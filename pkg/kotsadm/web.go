package kotsadm

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
)

var (
	executableMode = int32(484)
)

func ensureWeb(deployOptions *DeployOptions, clientset *kubernetes.Clientset) error {
	if deployOptions.Hostname == "" {
		hostname, err := promptForHostname()
		if err != nil {
			return errors.Wrap(err, "failed to prompt for hostname")
		}

		deployOptions.Hostname = hostname
	}

	if deployOptions.ServiceType == "" {
		serviceType, err := promptForWebServiceType(deployOptions)
		if err != nil {
			return errors.Wrap(err, "failed to prompt for service type")
		}

		deployOptions.ServiceType = serviceType
	}

	if err := ensureWebConfig(deployOptions, clientset); err != nil {
		return errors.Wrap(err, "failed to ensure web configmap")
	}

	if err := ensureWebDeployment(deployOptions.Namespace, clientset); err != nil {
		return errors.Wrap(err, "failed to ensure web deployment")
	}

	if err := ensureWebService(deployOptions, clientset); err != nil {
		return errors.Wrap(err, "failed to ensure web service")
	}

	return nil
}

func ensureWebConfig(deployOptions *DeployOptions, clientset *kubernetes.Clientset) error {
	_, err := clientset.CoreV1().ConfigMaps(deployOptions.Namespace).Get("kotsadm-web-scripts", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get existing config map")
		}

		configMap := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kotsadm-web-scripts",
				Namespace: deployOptions.Namespace,
			},
			Data: map[string]string{
				"start-kotsadm-web.sh": fmt.Sprintf(`#!/bin/bash
sed -i 's/###_GRAPHQL_ENDPOINT_###/http:\/\/%s\/graphql/g' /usr/share/nginx/html/index.html
sed -i 's/###_REST_ENDPOINT_###/http:\/\/%s\/api/g' /usr/share/nginx/html/index.html
sed -i 's/###_GITHUB_CLIENT_ID_###/not-supported/g' /usr/share/nginx/html/index.html
sed -i 's/###_SHIPDOWNLOAD_ENDPOINT_###/http:\/\/%s\/api\/v1\/download/g' /usr/share/nginx/html/index.html
sed -i 's/###_SHIPINIT_ENDPOINT_###/http:\/\/%s\/api\/v1\/init\//g' /usr/share/nginx/html/index.html
sed -i 's/###_SHIPUPDATE_ENDPOINT_###/http:\/\/%s\/api\/v1\/update\//g' /usr/share/nginx/html/index.html
sed -i 's/###_SHIPEDIT_ENDPOINT_###/http:\/\/%s\/api\/v1\/edit\//g' /usr/share/nginx/html/index.html
sed -i 's/###_GITHUB_REDIRECT_URI_###/http:\/\/%s\/auth\/github\/callback/g' /usr/share/nginx/html/index.html
sed -i 's/###_GITHUB_INSTALL_URL_###/not-supportetd/g' /usr/share/nginx/html/index.html
sed -i 's/###_INSTALL_ENDPOINT_###/http:\/\/%s\/api\/install/g' /usr/share/nginx/html/index.html

nginx -g "daemon off;"`, deployOptions.Hostname, deployOptions.Hostname,
					deployOptions.Hostname, deployOptions.Hostname, deployOptions.Hostname,
					deployOptions.Hostname, deployOptions.Hostname, deployOptions.Hostname),
			},
		}

		_, err := clientset.CoreV1().ConfigMaps(deployOptions.Namespace).Create(configMap)
		if err != nil {
			return errors.Wrap(err, "failed to create configmap")
		}
	}

	return nil
}

func ensureWebDeployment(namespace string, clientset *kubernetes.Clientset) error {
	_, err := clientset.AppsV1().Deployments(namespace).Get("kotsadm-web", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get existing deployment")
		}

		deployment := &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Deployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kotsadm-web",
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "kotsadm-web",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "kotsadm-web",
						},
					},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "kotsadm-web-scripts",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										DefaultMode: &executableMode,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "kotsadm-web-scripts",
										},
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Image:           "kotsadm/kotsadm-web:alpha",
								ImagePullPolicy: corev1.PullAlways,
								Name:            "kotsadm-web",
								Args: []string{
									"/scripts/start-kotsadm-web.sh",
								},
								Ports: []corev1.ContainerPort{
									{
										Name:          "http",
										ContainerPort: 3000,
									},
								},
								ReadinessProbe: &corev1.Probe{
									FailureThreshold:    3,
									InitialDelaySeconds: 2,
									PeriodSeconds:       2,
									SuccessThreshold:    1,
									TimeoutSeconds:      1,
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path:   "/healthz",
											Port:   intstr.FromInt(3000),
											Scheme: corev1.URISchemeHTTP,
										},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "kotsadm-web-scripts",
										MountPath: "/scripts/start-kotsadm-web.sh",
										SubPath:   "start-kotsadm-web.sh",
									},
								},
							},
						},
					},
				},
			},
		}

		_, err := clientset.AppsV1().Deployments(namespace).Create(deployment)
		if err != nil {
			return errors.Wrap(err, "failed to create deployment")
		}
	}

	return nil
}

func ensureWebService(deployOptions *DeployOptions, clientset *kubernetes.Clientset) error {
	port := corev1.ServicePort{
		Name:       "http",
		Port:       3000,
		TargetPort: intstr.FromString("http"),
	}

	serviceType := corev1.ServiceTypeClusterIP
	if deployOptions.ServiceType == "NodePort" {
		serviceType = corev1.ServiceTypeNodePort
		port.NodePort = int32(deployOptions.NodePort)
	} else if deployOptions.ServiceType == "LoadBalancer" {
		serviceType = corev1.ServiceTypeLoadBalancer
	}

	_, err := clientset.CoreV1().Services(deployOptions.Namespace).Get("kotsadm-web", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get existing service")
		}

		service := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kotsadm-web",
				Namespace: deployOptions.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "kotsadm-web",
				},
				Type: serviceType,
				Ports: []corev1.ServicePort{
					port,
				},
			},
		}

		_, err := clientset.CoreV1().Services(deployOptions.Namespace).Create(service)
		if err != nil {
			return errors.Wrap(err, "Failed to create service")
		}
	}

	return nil
}

func promptForWebServiceType(deployOptions *DeployOptions) (string, error) {
	prompt := promptui.Select{
		Label: "Web/UI Service Type:",
		Items: []string{"ClusterIP", "NodePort", "LoadBalancer"},
	}

	for {
		_, result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(-1)
			}
			continue
		}

		if result == "NodePort" {
			nodePort, err := promptForWebNodePort()
			if err != nil {
				return "", errors.Wrap(err, "failed to prompt for node port")
			}

			deployOptions.NodePort = int32(nodePort)
		}
		return result, nil
	}

}

func promptForWebNodePort() (int, error) {
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | bold }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "Node Port:",
		Templates: templates,
		Default:   "30000",
		Validate: func(input string) error {
			_, err := strconv.Atoi(input)
			return err
		},
	}

	for {
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(-1)
			}
			continue
		}

		nodePort, err := strconv.Atoi(result)
		if err != nil {
			return 0, errors.Wrap(err, "failed to convert nodeport")
		}

		return nodePort, nil
	}

}

func promptForHostname() (string, error) {
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | bold }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "Hostname:",
		Templates: templates,
		Default:   "",
		Validate: func(input string) error {
			if !strings.Contains(input, ":") {
				errs := validation.IsDNS1123Subdomain(input)
				if len(errs) > 0 {
					return errors.New(errs[0])
				}

				return nil
			}

			split := strings.Split(input, ":")
			if len(split) != 2 {
				return errors.New("only hostname or hostname:port are allowed formats")
			}

			errs := validation.IsDNS1123Subdomain(split[0])
			if len(errs) > 0 {
				return errors.New(errs[0])
			}

			_, err := strconv.Atoi(split[1])
			if err != nil {
				return err
			}

			return nil
		},
	}

	for {
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(-1)
			}
			continue
		}

		return result, nil
	}

}
