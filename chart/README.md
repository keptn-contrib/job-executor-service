
job-executor-service
===========

Helm Chart for the keptn job-executor-service


## Configuration

The following table lists the configurable parameters of the job-executor-service chart and their default values.

| Parameter                                    | Description                                                                                                                                                  | Default                                         |
|----------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------|
| `jobexecutorservice.image.repository`        | Container image name                                                                                                                                         | `"docker.io/keptncontrib/job-executor-service"` |
| `jobexecutorservice.image.pullPolicy`        | Kubernetes image pull policy                                                                                                                                 | `"IfNotPresent"`                                |
| `jobexecutorservice.image.tag`               | Container tag                                                                                                                                                | `""`                                            |
| `jobexecutorservice.service.enabled`         | Creates a kubernetes service for the job-executor-service                                                                                                    | `true`                                          |
| `distributor.stageFilter`                    | Sets the stage this helm service belongs to                                                                                                                  | `""`                                            |
| `distributor.serviceFilter`                  | Sets the service this helm service belongs to                                                                                                                | `""`                                            |
| `distributor.projectFilter`                  | Sets the project this helm service belongs to                                                                                                                | `""`                                            |
| `distributor.image.repository`               | Container image name                                                                                                                                         | `"docker.io/keptn/distributor"`                 |
| `distributor.image.pullPolicy`               | Kubernetes image pull policy                                                                                                                                 | `"IfNotPresent"`                                |
| `distributor.image.tag`                      | Container tag                                                                                                                                                | `""`                                            |
| `jobConfig.allowedImageList`                 | A comma separated list of images that are allowed in job workloads                                                                                           | `""`                                            |
| `jobConfig.allowPrivilegedJobs`              | Allows privileged job workloads. ***Allowing privileged job workloads can be considered dangerous!***                                                        | `false`                                         |
| `jobConfig.podSecurityContext`               | The default pod security context for job workloads                                                                                                           | [See values.yaml](values.yaml)                  |
| `jobConfig.jobSecurityContext`               | The default security context for job workloads                                                                                                               | [See values.yaml](values.yaml)                  |
| `jobConfig.serviceAccount.create`            | Enables the creation of the default service account used for job workloads                                                                                   | `true`                                          | 
| `jobConfig.serviceAccount.name`              | The name of the default service account used for job workloads                                                                                               | `default-job-account`                           | 
| `jobConfig.serviceAccount.annotations`       | Additional annotations for the default service account used for job workloads                                                                                | `{}`                                            |
| `jobConfig.taskDeadlineSeconds`              | Maximum duration for a kubernetes job run in seconds (0 means no limit, set it to an integer > 0 to enforce it)                                              | `0`                                             |
| `jobConfig.labels`                           | Additional labels that are added to all kubernetes jobs                                                                                                      | `{}`                                            |
 | `jobConfig.networkPolicy.enabled`            | Enable a network policy for jobs such that they can not access cluster internal resources                                                                    | `false`                                         |
 | `jobConfig.networkPolicy.allowAccessToKeptn` | All jobs to access the Keptn API server while the network policy is active                                                                                   | `false`                                         |
| `remoteControlPlane.autoDetect.enabled`      | Enables auto detection of a Keptn installation                                                                                                               | `false`                                         |
| `remoteControlPlane.autoDetect.namespace`    | Namespace which should be used by the auto-detection                                                                                                         | `""`                                            |
| `remoteControlPlane.api.protocol`            | Used protocol (http, https                                                                                                                                   | `"https"`                                       |
| `remoteControlPlane.api.hostname`            | Hostname of the control plane cluster (and port)                                                                                                             | `"api-gateway-nginx.keptn"`                     |
| `remoteControlPlane.api.apiValidateTls`      | Defines if the control plane certificate should be validated                                                                                                 | `true`                                          |
| `remoteControlPlane.api.token`               | Keptn api token                                                                                                                                              | `""`                                            |
| `imagePullSecrets`                           | Secrets to use for container registry credentials                                                                                                            | `[]`                                            |
| `serviceAccount.create`                      | Enables the service account creation                                                                                                                         | `true`                                          |
| `serviceAccount.annotations`                 | Annotations to add to the service account                                                                                                                    | `{}`                                            |
| `serviceAccount.name`                        | The name of the service account to use.                                                                                                                      | `""`                                            |
| `podAnnotations`                             | Annotations to add to the created pods                                                                                                                       | `{}`                                            |
| `podSecurityContext`                         | Set the pod security context. ***For security purposes the podSecurityContext value should not be changed!***                                                | [See values.yaml](values.yaml)                  |
| `securityContext`                            | Set the security context. ***For security purposes the securityContext value should not be changed!***                                                       | [See values.yaml](values.yaml)                  |
| `resources`                                  | Resource limits and requests                                                                                                                                 | `{}`                                            |
| `nodeSelector`                               | Node selector configuration                                                                                                                                  | `{}`                                            |
| `tolerations`                                | Tolerations for the pods                                                                                                                                     | `[]`                                            |
| `affinity`                                   | Affinity rules                                                                                                                                               | `{}`                                            |
| `networkPolicy.ingress.enabled`           | Enable job-executor-service ingress network policy (no ingress traffic allowed)                                                                                           | false                                           |
| `networkPolicy.egress.enabled`                      | Enable job-executor-service egress network policy: only egress allowed to OAuth provider, k8s master and Keptn API (may not work in case of dynamic/elastic IP addresses) | false                                           |
| `networkPolicy.egress.k8sMasterCIDR`                | Define kubernetes master(s) CIDR, if left empty we'll try to autodetect the master(s) IP address by looking up the endpoints of `kubernetes.default` service | ""                                              |
| `networkPolicy.egress.k8sMasterPort`                | Define kubernetes master(s) https port, if set to 0 we'll try to autodetect the port by looking up the endpoints of `kubernetes.default` service             | ""                                              |




