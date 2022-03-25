
job-executor-service
===========

Helm Chart for the keptn job-executor-service


## Configuration

The following table lists the configurable parameters of the job-executor-service chart and their default values.

| Parameter                | Description             | Default        |
| ------------------------ | ----------------------- | -------------- |
| `jobexecutorservice.image.repository` | Container image name | `"docker.io/keptncontrib/job-executor-service"` |
| `jobexecutorservice.image.pullPolicy` | Kubernetes image pull policy | `"IfNotPresent"` |
| `jobexecutorservice.image.tag` | Container tag | `""` |
| `jobexecutorservice.service.enabled` | Creates a kubernetes service for the job-executor-service | `true` |
| `distributor.stageFilter` | Sets the stage this helm service belongs to | `""` |
| `distributor.serviceFilter` | Sets the service this helm service belongs to | `""` |
| `distributor.projectFilter` | Sets the project this helm service belongs to | `""` |
| `distributor.image.repository` | Container image name | `"docker.io/keptn/distributor"` |
| `distributor.image.pullPolicy` | Kubernetes image pull policy | `"IfNotPresent"` |
| `distributor.image.tag` | Container tag | `""` |
| `remoteControlPlane.enabled` | Enables remote execution plane mode | `true` |
| `remoteControlPlane.api.protocol` | Used protocol (http, https | `"https"` |
| `remoteControlPlane.api.hostname` | Hostname of the control plane cluster (and port) | `""` |
| `remoteControlPlane.api.apiValidateTls` | Defines if the control plane certificate should be validated | `true` |
| `remoteControlPlane.api.token` | Keptn api token | `""` |
| `imagePullSecrets` | Secrets to use for container registry credentials | `[]` |
| `serviceAccount.create` | Enables the service account creation | `true` |
| `serviceAccount.annotations` | Annotations to add to the service account | `{}` |
| `serviceAccount.name` | The name of the service account to use. | `""` |
| `podAnnotations` | Annotations to add to the created pods | `{}` |
| `podSecurityContext` | Set the pod security context. ***For security puropses the podSecurityContext value should not be changed!*** | [See values.yaml](chart/values.yaml)            |
| `securityContext` | Set the security context. ***For security puropses the securityContext value should not be changed!***        | [See values.yaml](chart/values.yaml)  |
| `resources` | Resource limits and requests | `{}` |
| `nodeSelector` | Node selector configuration | `{}` |
| `tolerations` | Tolerations for the pods | `[]` |
| `affinity` | Affinity rules | `{}` |





