
keptn-service-template-go
===========

Helm Chart for the keptn keptn-service-template-go


## Configuration

The following table lists the configurable parameters of the keptn-service-template-go chart and their default values.

| Parameter                | Description             | Default        |
| ------------------------ | ----------------------- | -------------- |
| `keptnservice.image.repository` | Container image name | `"docker.io/keptnsandbox/keptn-service-template-go"` |
| `keptnservice.image.pullPolicy` | Kubernetes image pull policy | `"IfNotPresent"` |
| `keptnservice.image.tag` | Container tag | `""` |
| `keptnservice.service.enabled` | Creates a kubernetes service for the keptn-service-template-go | `true` |
| `distributor.stageFilter` | Sets the stage this helm service belongs to | `""` |
| `distributor.serviceFilter` | Sets the service this helm service belongs to | `""` |
| `distributor.projectFilter` | Sets the project this helm service belongs to | `""` |
| `distributor.image.repository` | Container image name | `"docker.io/keptn/distributor"` |
| `distributor.image.pullPolicy` | Kubernetes image pull policy | `"IfNotPresent"` |
| `distributor.image.tag` | Container tag | `""` |
| `remoteControlPlane.enabled` | Enables remote execution plane mode | `false` |
| `remoteControlPlane.api.protocol` | Used protocol (http, https | `"https"` |
| `remoteControlPlane.api.hostname` | Hostname of the control plane cluster (and port) | `""` |
| `remoteControlPlane.api.apiValidateTls` | Defines if the control plane certificate should be validated | `true` |
| `remoteControlPlane.api.token` | Keptn api token | `""` |
| `imagePullSecrets` | Secrets to use for container registry credentials | `[]` |
| `serviceAccount.create` | Enables the service account creation | `true` |
| `serviceAccount.annotations` | Annotations to add to the service account | `{}` |
| `serviceAccount.name` | The name of the service account to use. | `""` |
| `podAnnotations` | Annotations to add to the created pods | `{}` |
| `podSecurityContext` | Set the pod security context (e.g. fsgroups) | `{}` |
| `securityContext` | Set the security context (e.g. runasuser) | `{}` |
| `resources` | Resource limits and requests | `{}` |
| `nodeSelector` | Node selector configuration | `{}` |
| `tolerations` | Tolerations for the pods | `[]` |
| `affinity` | Affinity rules | `{}` |





