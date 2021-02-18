package kotsadm

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/replicatedhq/kots/pkg/kotsadm/types"
	"github.com/replicatedhq/kots/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func DistributionConfigMap(deployOptions types.DeployOptions) *corev1.ConfigMap {
	labels := types.GetKotsadmLabels()
	labels["kotsadm"] = "application"

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kotsadm-storage-registry-config",
			Namespace: deployOptions.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"config.yml": string(`
health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3
http:
  addr: :5000
  headers:
    X-Content-Type-Options:
      - nosniff
log:
  fields:
    service: registry
storage:
  cache:
    blobdescriptor: inmemory
version: 0.1`),
		},
	}

	return configMap
}

func DistributionService(deployOptions types.DeployOptions) *corev1.Service {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kotsadm-storage-registry",
			Namespace: deployOptions.Namespace,
			Labels:    types.GetKotsadmLabels(),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "kotsadm-storage-registry",
			},
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "registry",
					Port:       5000,
					TargetPort: intstr.FromInt(5000),
				},
			},
		},
	}

	return service
}

func DistributionStatefulset(deployOptions types.DeployOptions) *appsv1.StatefulSet {
	size := resource.MustParse("4Gi")

	if deployOptions.LimitRange != nil {
		var allowedMax *resource.Quantity
		var allowedMin *resource.Quantity

		for _, limit := range deployOptions.LimitRange.Spec.Limits {
			if limit.Type == corev1.LimitTypePersistentVolumeClaim {
				max, ok := limit.Max["storage"]
				if ok {
					allowedMax = &max
				}

				min, ok := limit.Min["storage"]
				if ok {
					allowedMin = &min
				}
			}
		}

		newSize := PromptForSizeIfNotBetween("registry", &size, allowedMin, allowedMax)
		if newSize == nil {
			os.Exit(-1)
		}

		size = *newSize
	}

	var securityContext corev1.PodSecurityContext
	if !deployOptions.IsOpenShift {
		securityContext = corev1.PodSecurityContext{
			RunAsUser: util.IntPointer(1000),
			FSGroup:   util.IntPointer(1000),
		}
	}

	statefulset := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kotsadm-storage-registry",
			Namespace: deployOptions.Namespace,
			Labels:    types.GetKotsadmLabels(),
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "kotsadm-storage-registry",
				},
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "kotsadm-storage-registry",
						Labels: types.GetKotsadmLabels(),
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceName(corev1.ResourceStorage): size,
							},
						},
					},
				},
			},
			ServiceName: "kotsadm-storage-registry",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: types.GetKotsadmLabels(map[string]string{
						"app": "kotsadm-storage-registry",
					}),
				},
				Spec: corev1.PodSpec{
					SecurityContext: &securityContext,
					Containers: []corev1.Container{
						{
							Name:            "docker-registry",
							Image:           "registry:2.7.1",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/bin/registry",
								"/serve",
								"/etc/docker/registry/config.yml",
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 5000,
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(5000),
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(5000),
									},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "REGISTRY_HTTP_SECERET",
									Value: "to-generate-",
								},
								{
									Name:  "REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY",
									Value: "/var/lib/registry",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "kotsadm-storage-registry",
									MountPath: "/var/lib/registry",
								},
								{
									Name:      "kotsadm-storage-registry-config",
									MountPath: "/etc/docker/registry",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "kotsadm-storage-registry",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "kotsadm-storage-registry",
								},
							},
						},
						{
							Name: "kotsadm-storage-registry-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "kotsadm-storage-registry-config",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return statefulset
}

func PromptForSizeIfNotBetween(label string, desired *resource.Quantity, min *resource.Quantity, max *resource.Quantity) *resource.Quantity {
	actualSize := desired

	if max != nil {
		if max.Cmp(*desired) == -1 {
			/// desired is too big
			actualSize = max
		}
	}
	if min != nil {
		if min.Cmp(*desired) == 1 {
			/// desired is too small, yeap, you read that right
			actualSize = min
		}
	}

	if actualSize.Cmp(*desired) == 0 {
		return desired
	}

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("The storage request for %s is not acceptable for the current namespace. KOTS recommends a size of %s, but will attempt to proceed with %s to meet the namespace limits. Do you want to continue", label, desired.String(), actualSize.String()),
		IsConfirm: true,
	}

	for {
		_, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(-1)
			}
			continue
		}

		return actualSize
	}
}
