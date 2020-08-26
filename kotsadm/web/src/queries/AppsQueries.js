import gql from "graphql-tag";

export const ping = gql(`
  query ping {
    ping
  }
`);

export const listAppsRaw = `
  query listApps {
    listApps {
      kotsApps {
        id
        name
        iconUri
        createdAt
        updatedAt
        slug
        currentSequence
        isGitOpsSupported
        allowSnapshots
        licenseType
        updateCheckerSpec
        currentVersion {
          title
          status
          createdOn
          sequence
          deployedAt
          yamlErrors {
            path
            error
          }
        }
        lastUpdateCheckAt
        downstreams {
          name
          currentVersion {
            title
            status
            createdOn
            sequence
            deployedAt
            yamlErrors {
              path
              error
            }
          }
          gitops {
            enabled
            provider
            uri
            hostname
            path
            branch
            format
            action
            isConnected
          }
          pendingVersions {
            title
            status
            createdOn
            sequence
            deployedAt
            yamlErrors {
              path
              error
            }
          }
          pastVersions {
            title
            status
            createdOn
            sequence
            deployedAt
            yamlErrors {
              path
              error
            }
          }
          cluster {
            id
            title
            slug
            createdOn
            lastUpdated
            currentVersion {
              title
              status
              createdOn
              sequence
              deployedAt
              yamlErrors {
                path
                error
              }
            }
            shipOpsRef {
              token
            }
            totalApplicationCount
          }
        }
      }
    }
  }
`;
export const listApps = gql(listAppsRaw);

export const getKotsAppRaw = `
  query getKotsApp($slug: String!) {
    getKotsApp(slug: $slug) {
      id
      name
      iconUri
      createdAt
      updatedAt
      slug
      upstreamUri
      currentSequence
      hasPreflight
      isAirgap
      isConfigurable
      isGitOpsSupported
      allowRollback
      allowSnapshots
      licenseType
      updateCheckerSpec
      currentVersion {
        title
        status
        createdOn
        sequence
        releaseNotes
        deployedAt
        yamlErrors {
          path
          error
        }
      }
      lastUpdateCheckAt
      bundleCommand
      downstreams {
        name
        links {
          title
          uri
        }
        currentVersion {
          title
          status
          createdOn
          sequence
          deployedAt
          source
          releaseNotes
          parentSequence
          yamlErrors {
            path
            error
          }
        }
        pendingVersions {
          title
          status
          createdOn
          sequence
          deployedAt
          parentSequence
          yamlErrors {
            path
            error
          }
        }
        pastVersions {
          title
          status
          createdOn
          sequence
          deployedAt
          parentSequence
          yamlErrors {
            path
            error
          }
        }
        gitops {
          enabled
          provider
          uri
          hostname
          path
          branch
          format
          action
          deployKey
          isConnected
        }
        cluster {
          id
          title
          slug
          createdOn
          lastUpdated
          currentVersion {
            title
            status
            createdOn
            sequence
            deployedAt
            yamlErrors {
              path
              error
            }
          }
          shipOpsRef {
            token
          }
          totalApplicationCount
        }
      }
    }
  }
`;
export const getKotsApp = gql(getKotsAppRaw);

export const getAirgapInstallStatusRaw = `
  query getAirgapInstallStatus {
    getAirgapInstallStatus {
      installStatus
      currentMessage
    }
  }
`;
export const getAirgapInstallStatus = gql(getAirgapInstallStatusRaw);

export const getOnlineInstallStatusRaw = `
  query getOnlineInstallStatus {
    getOnlineInstallStatus {
      installStatus
      currentMessage
    }
  }
`;
export const getOnlineInstallStatus = gql(getOnlineInstallStatusRaw);

export const getImageRewriteStatusRaw = `
  query getImageRewriteStatus {
    getImageRewriteStatus {
      currentMessage
      status
    }
  }
`;
export const getImageRewriteStatus = gql(getImageRewriteStatusRaw);

export const getKotsDownstreamOutput = gql`
  query getKotsDownstreamOutput($appSlug: String!, $clusterSlug: String!, $sequence: Int!) {
    getKotsDownstreamOutput(appSlug: $appSlug, clusterSlug: $clusterSlug, sequence: $sequence) {
      dryrunStdout
      dryrunStderr
      applyStdout
      applyStderr
      renderError
    }
  }
`;

export const getKotsAppDashboard = gql`
  query getKotsAppDashboard($slug: String!, $clusterId: String) {
    getKotsAppDashboard(slug: $slug, clusterId: $clusterId) {
      appStatus {
        appId
        updatedAt
        state
        resourceStates {
          kind
          name
          namespace
          state
        }
      }
      metrics {
        title
        tickFormat
        tickTemplate
        series {
          legendTemplate
          metric {
            name
            value
          }
          data {
            timestamp
            value
          }
        }
      }
      prometheusAddress
    }
  }
`;

export const getPrometheusAddress = gql`
  query getPrometheusAddress {
    getPrometheusAddress
  }
`;
