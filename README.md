# Job Executor Service

![GitHub release (latest by date)](https://img.shields.io/github/v/release/keptn-contrib/job-executor-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/keptn-contrib/job-executor-service)](https://goreportcard.com/report/github.com/keptn-contrib/job-executor-service)

This Keptn integration introduces a new approach of running customizable tasks with Keptn as Kubernetes Jobs.

## Motivation

The job-executor-service aims to tackle several current pain points with the current approach of services/integrations
running in the Keptn ecosystem:

| Problem                                                                                                                                                                                                                                                                                                                                        | Solution                                                                                                                                                                                                                                                                                                                                                                                                                                            |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Keptn services are constantly running on Kubernetes while listening for Cloud Events coming in over NATS, even when idle. This consumes unnecessary resources on the Kubernetes cluster.                                                                                                                                                       | By running the defined Keptn tasks as short-lived workloads (Kubernetes Jobs), they just consume resources while the task is executed.                                                                                                                                                                                                                                                                                                              |
| Whenever some new functionality should be triggered by Keptn, a new Keptn service needs to be written. It usually wraps the functionality of the wanted framework and executes it under the hood. The downside: The code of the new Keptn service needs to be written and maintained. This effort scales linearly with the amount of services. | This service can execute any framework with just a few lines of yaml configuration. No need to write or maintain any new code.                                                                                                                                                                                                                                                                                                                      |
| Keptn services usually filter for a static list of events the trigger the included functionality. This is not configurable. Whenever the service should listen to a new event, the code of the service needs to be changed.                                                                                                                    | The Job Executor Service provides the means to trigger a task execution for any Keptn event. This is done by matching a jsonpath to the received event payload.                                                                                                                                                                                                                                                                                     |
| Keptn services are usually opinionated on how your framework execution looks like. E.g. the locust service just executes three different (statically named) files depending on the test strategy in the shipyard. It is not possible to write tests consisting of multiple files.                                                              | This service provides the possibility to write any specified file from the Keptn git repository into a mounted folder (`/keptn`) of the Kubernetes job. This is done by a initcontainer running before the specified image.                                                                                                                                                                                                                         |
| Support for new functionality in Keptn needs to be added to each Keptn service individually. E.g. the new secret functionality needs to be included into all of the services running in the Keptn execution plane. This slows down the availability of this new feature.                                                                       | The Job Executor Service is a single service which provides the means to run any workload orchestrated by Keptn. So, it is possible to support new functionality of Keptn just once in this service - and all workloads profit from it. E.g. in the case of the secret functionality, one just needs to support it in this service and suddenly all the triggered Kubernetes Jobs have the correct secrets attached to it as environment variables. |



## Compatibility Matrix

| Keptn Version | [Job-Executor-Service Docker Image](https://hub.docker.com/r/keptncontrib/job-executor-service/tags) | Config version |
|:-------------:|:----------------------------------------------------------------------------------------------------:|:--------------:|
|    0.10.0     |                               keptncontrib/job-executor-service:0.1.5                                |       v2       |
|    0.10.0     |                               keptncontrib/job-executor-service:0.1.6                                |       v2       |
|    0.12.2     |                               keptncontrib/job-executor-service:0.1.7                                |       v2       |
|    0.12.6     |                               keptncontrib/job-executor-service:0.1.8                                |       v2       |
|    0.13.x     |                               keptncontrib/job-executor-service:0.1.9                                |       v2       |

Please note: Newer Keptn versions might be compatible, but compatibility has not been verified at the time of the release.

**Warning**: We are aware that there might be a problem when trying to install Job-Executor-Service in the same namespace as Keptn 0.14.x. 
For now, we advise to install Job-Executor-Service in a separate namespace, and set `remoteControlPlane.api.hostname`, `remoteControlPlane.api.token`, ... as detailed in the
[Installation section](#installation) below.


<details>
  <summary>Click here to show older versions</summary>
  
  | Keptn Version | [Job-Executor-Service Docker Image](https://hub.docker.com/r/keptncontrib/job-executor-service/tags) | Config version |
  |:-------------:|:----------------------------------------------------------------------------------------------------:|:--------------:|
  |     0.8.3     |                               keptncontrib/job-executor-service:0.1.0                                |       -        |
  |     0.8.3     |                               keptncontrib/job-executor-service:0.1.1                                |       -        |
  |     0.8.4     |                               keptncontrib/job-executor-service:0.1.2                                |       v1       |
  |     0.8.6     |                               keptncontrib/job-executor-service:0.1.3                                |       v2       |
  |     0.9.0     |                               keptncontrib/job-executor-service:0.1.4                                |       v2       |

</details>

## Installation

The *job-executor-service* can be installed as a part of [Keptn's uniform](https://keptn.sh) using `helm`.
It is recommended to install the *job-executor-service* on the remote execution-plane, which can be on the same cluster, or on a completely
separate Kubernetes environment (see [Keptn docs: Multi-cluster setup](https://keptn.sh/docs/0.11.x/operate/multi_cluster/) for details).

During the installation process various parameters of the *job-executor-service* can be customized, for a complete list
of these helm values have a look at the [documentation](chart/README.md).
In order to install the *job-executor-service* on the remote execution plane, some values of the helm chart need to be configured:
* `remoteControlPlane.topicSubscription` - list of Keptn CloudEvent types that this instance should listen to, e.g., `sh.keptn.event.remote-task.triggered`
* `remoteControlPlane.api.protocol` - protocol (`http` or `https`) used to connect to the remote control plane
* `remoteControlPlane.api.hostname` - Keptn API Hostname (e.g., `1.2.3.4.nip.io`). If Keptn is installed on the same cluster the API is usually reachable under `api-gateway-nginx.keptn`.
* `remoteControlPlane.api.token` - Keptn API Token (can be obtained from Bridge)

**Example**

```bash
KEPTN_API_PROTOCOL=http # or https
KEPTN_API_HOST=<INSERT-YOUR-HOSTNAME-HERE> # e.g., 1.2.3.4.nip.io
 KEPTN_API_TOKEN=<INSERT-YOUR-KEPTN-API-TOKEN-HERE>

TASK_SUBSCRIPTION=sh.keptn.event.remote-task.triggered

helm upgrade --install --create-namespace -n <NAMESPACE> \
  job-executor-service https://github.com/keptn-contrib/job-executor-service/releases/download/<VERSION>/job-executor-service-<VERSION>.tgz \
 --set remoteControlPlane.topicSubscription=${TASK_SUBSCRIPTION},remoteControlPlane.api.protocol=${KEPTN_API_PROTOCOL},remoteControlPlane.api.hostname=${KEPTN_API_HOST},remoteControlPlane.api.token=${KEPTN_API_TOKEN}
```

Please replace `<VERSION>` with the actual version you want to install from the compatibility matrix above or the
[GitHub releases page](https://github.com/keptn-contrib/job-executor-service/releases).

To verify that everything works you can visit Bridge, select a project, go to Uniform, and verify that `job-executor-service`  is registered as "remote execution plane" with the correct version and event type.

**Example (with auto-detection)**

For easier installation of the *job-executor-service* a Keptn installation on the same kubernetes cluster can be discovered automatically. This only
works if no `remoteControlPlane.api.token`, `remoteControlPlane.api.protocol` or `remoteControlPlane.api.hostname` is provided.
If multiple Keptn installations are present, the `remoteControlPlane.autoDetect.namespace` must be set to the desired Keptn instance.
The auto-detection feature can be enabled by setting `remoteControlPlane.autoDetect.enabled` to `true`.

```bash
TASK_SUBSCRIPTION=sh.keptn.event.remote-task.triggered

helm upgrade --install --create-namespace -n <NAMESPACE> \
  job-executor-service https://github.com/keptn-contrib/job-executor-service/releases/download/<VERSION>/job-executor-service-<VERSION>.tgz \
 --set remoteControlPlane.autoDetect.enabled=true,remoteControlPlane.topicSubscription=${TASK_SUBSCRIPTION},remoteControlPlane.api.token="",remoteControlPlane.api.hostname="",remoteControlPlane.api.protocol=""
```


### Update API Token on Remote Execution-Plane

To update your `KEPTN_API_TOKEN` on an existing installation, please execute the following command (make sure to use the same `<VERSION>` is currently installed):

```bash
 KEPTN_API_TOKEN=<INSERT-YOUR-NEW-KEPTN-API-TOKEN-HERE>

helm upgrade -n <NAMESPACE> \
  job-executor-service https://github.com/keptn-contrib/job-executor-service/releases/download/<VERSION>/job-executor-service-<VERSION>.tgz \
  --reuse-values \
 --set remoteControlPlane.api.token=${KEPTN_API_TOKEN}
```

### Update Topic Subscriptions

To update your `TASK_SUBSCRIPTION` (as in the Cloud Event types that job-executor-service is listening to), please execute the following command (make sure to use the same `<VERSION>` is currently installed):

```bash
TASK_SUBSCRIPTION=sh.keptn.event.remote-task.triggered,sh.keptn.event.some-other-task.triggered

helm upgrade -n <NAMESPACE> \
  job-executor-service https://github.com/keptn-contrib/job-executor-service/releases/download/<VERSION>/job-executor-service-<VERSION>.tgz \
  --reuse-values \
 --set remoteControlPlane.topicSubscription=${TASK_SUBSCRIPTION}
```

## Upgrade

To upgrade to a newer version of *job-executor-service*, run

```bash
helm upgrade -n <NAMESPACE> \
  job-executor-service https://github.com/keptn-contrib/job-executor-service/releases/download/<VERSION>/job-executor-service-<VERSION>.tgz \
  --reuse-values
```

To upgrade to a newer version of *job-executor-service* and automatically use the auto-detection to configure the Keptn API token, run

```bash
helm upgrade -n <NAMESPACE> \
  job-executor-service https://github.com/keptn-contrib/job-executor-service/releases/download/<VERSION>/job-executor-service-<VERSION>.tgz \
  --reuse-values \
  --set remoteControlPlane.autoDetect.enabled=true,remoteControlPlane.topicSubscription=${TASK_SUBSCRIPTION},remoteControlPlane.api.token="",remoteControlPlane.api.hostname="",remoteControlPlane.api.protocol=""
```

## Uninstall

To uninstall *job-executor-service*, run

```bash
helm uninstall -n <NAMESPACE> job-executor-service
```

## Development

Development can be conducted using any GoLang compatible IDE/editor (e.g., Jetbrains GoLand, VSCode with Go plugins).

When writing code, it is recommended to follow the coding style suggested by the [Golang community](https://github.com/golang/go/wiki/CodeReviewComments).

### Common tasks

* Run tests: `go test -race -v ./...`
* Deploy the service during development using [Skaffold](https://skaffold.dev/):  
  `skaffold run --default-repo=<your-docker-registry> --tail`
(Note: replace `<your-docker-registry>` with your DockerHub username; The development deployment assumes that only one Keptn version is present on the cluster)
* Watch the deployment using `kubectl`: `kubectl -n keptn get deployment job-executor-service -o wide`
* Get logs using `kubectl`: `kubectl -n keptn logs deployment/job-executor-service -f`
* Watch the deployed pods using `kubectl`: `kubectl -n keptn get pods -l run=job-executor-service`

### How to release a new version of this service

It is assumed that the current development takes place in the `main` branch (either via Pull Requests or directly).

Creating a release is as simple as using the 
[Create pre-release](https://github.com/keptn-contrib/job-executor-service/actions/workflows/pre-release.yml) and 
[Create release](https://github.com/keptn-contrib/job-executor-service/actions/workflows/release.yml) workflows.

**Note**: Creating a pre-release will actually create a GitHub pre-release and tag the latest commit on the specified branch.
When creating a release, only a draft release as well as a pull request are created. You still need to publish the draft
release and merge the Pull Request.


## Quickstart

Just put a file into the Keptn config git repository of a service (in folder `<service>/job/config.yaml`) to specify

* the containers which should be run as Kubernetes Jobs and
* the events for which they should be triggered.

```yaml
apiVersion: v2
actions:
  - name: "Run something"
    events:
      - name: "sh.keptn.event.test.triggered"
    tasks:
      - name: "Greet the world"
        image: "alpine"
        cmd:
          - echo
        args:
          - "Hello World"

```

The easiest way to add the `config.yaml` to the keptn git repository is to use the `keptn` cli:

```shell
keptn add-resource --project=myproject --service=myservice --stage=mystage --resource=config.yaml --resourceUri=job/config.yaml
```


## How to validate a job configuration

`job-lint` is a simple cli tool that validates any given job configuration file and shows possible errors. You can download it on the [GH Releases page](https://github.com/keptn-contrib/job-executor-service/releases).

```shell
./job-lint test-data/config.yaml

2021/07/13 16:18:49 config ../test-data/config.yaml is valid
```

*Note*: The default behavior when linting a job configuration is to not allow privileged job workloads.
This behavior can be changed by specifying `-allow-privileged-jobs=true`.
The flag should match the job-executor-service [configuration](/chart/README.md) to avoid problems during runtime.

## Features

A more comprehensive list of use-cases and features that this integration supports is provided in [FEATURES.md](FEATURES.md).

## Architecture

A deep look into the architecture of job-executor-service is provided in [ARCHITECTURE.md](ARCHITECTURE.md).

## Credits

The credits of this service heavily go to @thschue and @yeahservice who originally came up with this idea. :rocket:

## License

Please find more information in the [LICENSE](LICENSE) file.
