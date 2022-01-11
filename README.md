# Job Executor Service

![GitHub release (latest by date)](https://img.shields.io/github/v/release/keptn-contrib/job-executor-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/keptn-contrib/job-executor-service)](https://goreportcard.com/report/github.com/keptn-contrib/job-executor-service)

This Keptn integration introduces a new approach of running customizable tasks with Keptn as Kubernetes Jobs.

## Compatibility Matrix

| Keptn Version | [Job-Executor-Service Docker Image](https://hub.docker.com/r/keptncontrib/job-executor-service/tags) | Config version |
|:-------------:|:----------------------------------------------------------------------------------------------------:|:--------------:|
|     0.8.3     |                               keptncontrib/job-executor-service:0.1.0                                |       -        |
|     0.8.3     |                               keptncontrib/job-executor-service:0.1.1                                |       -        |
|     0.8.4     |                               keptncontrib/job-executor-service:0.1.2                                |       v1       |
|     0.8.6     |                               keptncontrib/job-executor-service:0.1.3                                |       v2       |
|     0.9.0     |                               keptncontrib/job-executor-service:0.1.4                                |       v2       |
|    0.10.0     |                               keptncontrib/job-executor-service:0.1.5                                |       v2       |

## Installation

The *job-executor-service* can be installed as a part of [Keptn's uniform](https://keptn.sh) using `helm`:

```bash
helm install -n keptn job-executor-service https://github.com/keptn-contrib/job-executor-service/releases/download/<VERSION>/job-executor-service-<VERSION>.tgz
```

Please replace `<VERSION>` with the actual version you want to install from the compatibility matrix above or the 
[GitHub releases page](https://github.com/keptn-contrib/job-executor-service/releases).


**Note**: Versions 0.1.4 and older need to be installed using `kubectl`, e.g.:
```bash
kubectl apply -f https://raw.githubusercontent.com/keptn-contrib/job-executor-service/release-<VERSION>/deploy/service.yaml
```


## Uninstall

To uninstall *job-executor-service*, run

```bash
helm uninstall -n keptn job-executor-service
```

## Development

Development can be conducted using any GoLang compatible IDE/editor (e.g., Jetbrains GoLand, VSCode with Go plugins).

When writing code, it is recommended to follow the coding style suggested by the [Golang community](https://github.com/golang/go/wiki/CodeReviewComments).

### Common tasks

* Build the binary: `go build -ldflags '-linkmode=external' -v -o job-executor-service`
* Run tests: `go test -race -v ./...`
* Watch the deployment using `kubectl`: `kubectl -n keptn get deployment job-executor-service -o wide`
* Get logs using `kubectl`: `kubectl -n keptn logs deployment/job-executor-service -f`
* Watch the deployed pods using `kubectl`: `kubectl -n keptn get pods -l run=job-executor-service`
* Deploy the service
  using [Skaffold](https://skaffold.dev/): `skaffold run --default-repo=<your-docker-registry> --tail` (Note:
  Replace `<your-docker-registry>` with your DockerHub username


### How to release a new version of this service

It is assumed that the current development takes place in the `main` branch (either via Pull Requests or directly).

Creating a release is as simple as using the 
[Create pre-release](https://github.com/keptn-contrib/job-executor-service/actions/workflows/pre-release.yml) and 
[Create release](https://github.com/keptn-contrib/job-executor-service/actions/workflows/release.yml) workflows.

**Note**: Creating a pre-release will actually create a GitHub pre-release and tag the latest commit on the specified branch.
When creating a release, only a draft release as well as a pull request are created. You still need to publish the draft
release and merge the Pull Request.

## Credits

The credits of this service heavily go to @thschue and @yeahservice who originally came up with this idea. :rocket:

## License

Please find more information in the [LICENSE](LICENSE) file.

## Documentation

- [Documentation](#documentation)
    - [Why?](#why)
    - [How?](#how)
        - [Specifying the working directory](#specifying-the-working-directory)
        - [Event Matching](#event-matching)
        - [Kubernetes Job](#kubernetes-job)
        - [Kubernetes Job Environment Variables](#kubernetes-job-environment-variables)
            - [From Events](#from-events)
            - [From Kubernetes Secrets](#from-kubernetes-secrets)
            - [From String Literal](#from-string-literal)
        - [File Handling](#file-handling)
        - [Silent mode](#silent-mode)
        - [Resource quotas](#resource-quotas)
        - [Poll duration](#poll-duration)
        - [Job namespace](#job-namespace)
        - [Job image pull policy](#job-image-pull-policy)
        - [Send start/finished event if the job config.yaml can't be found](#send-startfinished-event-if-the-job-configyaml-cant-be-found)
        - [Additional Event Data](#additional-event-data)
        - [Remote Control Plane](#remote-control-plane)
        - [Job clean-up](#job-clean-up)
    - [How to validate a job configuration](#how-to-validate-a-job-configuration)
    - [Endless Possibilities](#endless-possibilities)

## Why?

The job-executor-service aims to tackle several current pain points with the current approach of services/integrations
running in the Keptn ecosystem:

| Problem                                                                                                                                                                                                                                                                                                                                        | Solution                                                                                                                                                                                                                                                                                                                                                                                                                                            |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Keptn services are constantly running on Kubernetes while listening for Cloud Events coming in over NATS, even when idle. This consumes unnecessary resources on the Kubernetes cluster.                                                                                                                                                       | By running the defined Keptn tasks as short-lived workloads (Kubernetes Jobs), they just consume resources while the task is executed.                                                                                                                                                                                                                                                                                                              |
| Whenever some new functionality should be triggered by Keptn, a new Keptn service needs to be written. It usually wraps the functionality of the wanted framework and executes it under the hood. The downside: The code of the new Keptn service needs to be written and maintained. This effort scales linearly with the amount of services. | This service can execute any framework with just a few lines of yaml configuration. No need to write or maintain any new code.                                                                                                                                                                                                                                                                                                                      |
| Keptn services usually filter for a static list of events the trigger the included functionality. This is not configurable. Whenever the service should listen to a new event, the code of the service needs to be changed.                                                                                                                    | The Job Executor Service provides the means to trigger a task execution for any Keptn event. This is done by matching a jsonpath to the received event payload.                                                                                                                                                                                                                                                                                     |
| Keptn services are usually opinionized on how your framework execution looks like. E.g. the locust service just executes three different (statically named) files depending on the test strategy in the shipyard. It is not possible to write tests consisting of multiply files.                                                              | This service provides the possibility to write any specified file from the Keptn git repository into a mounted folder (`/keptn`) of the Kubernetes job. This is done by a initcontainer running before the specified image.                                                                                                                                                                                                                         |
| Support for new functionality in Keptn needs to be added to each Keptn service individually. E.g. the new secret functionality needs to be included into all of the services running in the Keptn execution plane. This slows down the availability of this new feature.                                                                       | The Job Executor Service is a single service which provides the means to run any workload orchestrated by Keptn. So, it is possible to support new functionality of Keptn just once in this service - and all workloads profit from it. E.g. in the case of the secret functionality, one just needs to support it in this service and suddenly all the triggered Kubernetes Jobs have the correct secrets attached to it as environment variables. |

## How?

Just put a file into the Keptn config git repository (in folder `<service>/job/config.yaml`) to specify

* the containers which should be run as Kubernetes Jobs and
* the events for which they should be triggered.

```yaml
apiVersion: v2
actions:
  - name: "Run locust"
    events:
      - name: "sh.keptn.event.test.triggered"
        jsonpath:
          property: "$.data.test.teststrategy"
          match: "health"
      - name: "sh.keptn.event.test.triggered"
        jsonpath:
          property: "$.data.test.teststrategy"
          match: "load"
    tasks:
      - name: "Run locust tests"
        files:
          - locust/basic.py
          - locust/import.py
          - locust/locust.conf
        image: "locustio/locust"
        cmd:
          - locust
        args:
          - '--config'
          - /keptn/locust/locust.conf
          - '-f'
          - /keptn/locust/basic.py
          - '--host'
          - $(HOST)
        env:
          - name: HOST
            value: "$.data.deployment.deploymentURIsLocal[0]"
            valueFrom: event
```

The easiest way to add the `config.yaml` to the keptn git repository is to use the `keptn` cli:

```shell
keptn add-resource --project=myproject --service=myservice --stage=mystage --resource=config.yaml --resourceUri=job/config.yaml
```

### Specifying the working directory

Since all files are hosted by default under `/keptn` and some tools only operate on the current working directory, it is
also possible to switch the working directory of the container. This can be achieved by setting the `workingDir`
property in a task object.

Here an example:

```yaml
apiVersion: v2
actions:
  - name: "Print files"
    events:
      - name: "sh.keptn.event.sample.triggered"
    tasks:
      - name: "Show files in bin"
        image: "alpine"
        workingDir: "/bin"
        cmd:
          - ls
```

In this example, the `ls` command will be run in the `/bin` folder.

### Event Matching

The tasks of an action are executed if the event name matches. Wildcards can also be used, e.g.

```yaml
- name: "sh.keptn.event.*.triggered"
```

Would match events `sh.keptn.event.test.triggered`, `sh.keptn.event.deployment.triggered` and so on.

Optionally the following section can be added to an event:

```yaml
jsonpath:
  property: "$.data.test.teststrategy"
  match: "locust"
```

If the service receives an event which matches the name, and the jsonpath match expression, the specified tasks are
executed. E.g. the following cloud event would match the jsonpath above:

```json
{
  "type": "sh.keptn.event.test.triggered",
  "specversion": "1.0",
  "source": "test-events",
  "id": "f2b878d3-03c0-4e8f-bc3f-454bc1b3d79b",
  "time": "2019-06-07T07:02:15.64489Z",
  "contenttype": "application/json",
  "shkeptncontext": "08735340-6f9e-4b32-97ff-3b6c292bc50i",
  "data": {
    "project": "sockshop",
    "stage": "dev",
    "service": "carts",
    "labels": {
      "testId": "4711",
      "buildId": "build-17",
      "owner": "JohnDoe"
    },
    "status": "succeeded",
    "result": "pass",
    "test": {
      "teststrategy": "locust"
    }
  }
}
```

### Kubernetes Job

The configuration contains the following section:

```yaml
tasks:
  - name: "Run locust tests"
    files:
      - locust/basic.py
      - locust/import.py
      - locust/locust.conf
    image: "locustio/locust"
    cmd:
      - locust
    args:
      - '--config'
      - /keptn/locust/locust.conf
      - '-f'
      - /keptn/locust/basic.py
      - '--host'
      - $(HOST)
```

It contains the tasks which should be executed as Kubernetes job. The service schedules a different job for each of
these tasks in the order, they are listed within the config. The service waits for the successful execution of all the
tasks to respond with a `StatusSucceeded` finished event. When one of the events fail, it responds with `StatusErrored`
finished cloud event.

### Kubernetes Job Environment Variables

In the `env` section of a task, a list of environment variables can be declared, with their source either from the
incoming cloud event (`valueFrom: event`) or from kubernetes secrets (`valueFrom: secret`).

Environment variables in `cmd` can be accessed with parentheses, e.g. `"$(HOST)"`. This is required for the variable to
be expanded in the command.

#### From Events

The following environment variable has the name `HOST`, and the value is whatever the given
jsonpath `$.data.deployment.deploymentURIsLocal[0]` resolves to.

```yaml
cmd:
  - locust
args:
  - '--config'
  - /keptn/locust/locust.conf
  - '-f'
  - /keptn/locust/basic.py
  - '--host'
  - $(HOST)
env:
  - name: HOST
    value: "$.data.deployment.deploymentURIsLocal[0]"
    valueFrom: event
```

In the above example the json path for `HOST` would resolve into `https://keptn.sh` for the below event

```yaml
{
  "data": {
    "deployment": {
      "deploymentNames": [
          "user_managed"
      ],
      "deploymentURIsLocal": [
          "https://keptn.sh"
      ],
      "deploymentURIsPublic": [
          ""
      ],
      "deploymentstrategy": "user_managed",
      "gitCommit": "eb5fc3d5253b1845d3d399c880c329374bbbb30e"
    },
    "message": "",
    "project": "sockshop",
    "stage": "dev",
    "service": "carts",
    "status": "succeeded",
    "test": {
      "teststrategy": "health"
    }
  },
  "id": "4fe1eed1-49e2-49a9-91af-a42c8b0f7811",
  "source": "shipyard-controller",
  "specversion": "1.0",
  "time": "2021-05-13T07:46:09.546Z",
  "type": "sh.keptn.event.test.triggered",
  "shkeptncontext": "138f7bf1-f027-42c4-b705-9033b5f5871e"
}
```

The Job executor service also allows you to format the event data in different formats including `JSON` and `YAML` using
the `as` keyword. This can be useful when working with the whole event or an object with multiple fields.

```yaml
cmd:
  - locust
args:
  - '--config'
  - /keptn/locust/locust.conf
  - '-f'
  - /keptn/locust/basic.py
  - '--host'
  - $(HOST)
env:
  - name: EVENT
    value: "$"
    valueFrom: event
    as: json
```

If the `as` keyword is omitted the job executor defaults to `string` for a single value and `JSON` for a `map` type.

#### From Kubernetes Secrets

The following configuration looks up a kubernetes secret with the name `locust-secret` and all key/value pairs of the
secret will be available as separate environment variables in the job.

The kubernetes secret is always looked up in the [namespace](#job-namespace) the respective task runs in.

```yaml
cmd:
  - locust
args:
  - '--config'
  - /keptn/locust/locust.conf
  - '-f'
  - /keptn/locust/$(FILE)
  - '--host'
  - $(HOST)
env:
  - name: locust-secret
    valueFrom: secret
```

With the secret below, there will be two environment variables available in the job. `HOST` with the
value `https://keptn.sh` and `FILE` with the value `basic.py`

```shell
kubectl -n keptn create secret generic locust-secret --from-literal=HOST=https://keptn.sh --from-literal=FILE=basic.py -oyaml --dry-run=client
```

```yaml
apiVersion: v1
data:
  FILE: YmFzaWMucHk=
  HOST: aHR0cHM6Ly9rZXB0bi5zaA==
kind: Secret
metadata:
  creationTimestamp: null
  name: locust-secret
  namespace: keptn
```

#### From String Literal

It sometimes makes sense to provide a static string value as an environment variable. This can be done by specifying
a `string` as the `valueFrom` value. The value of the environment variable can then be specified by the `value`
property.

Here an example

```yaml
cmd:
  - locust
args:
  - '--config'
  - /keptn/locust/locust.conf
  - '-f'
  - /keptn/locust/$(FILE)
  - '--host'
  - $(HOST)
env:
  - name: DATA_DIR
    valueFrom: string
    value: /tmp/data
```

This makes the `DATA_DIR` env variable with the value `/tmp/data`
available to the cmd.

### File Handling

Single files or all files in a directory can be added to your running tasks by specifying them in the `files` section of
your tasks:

```yaml
files:
  - locust/basic.py
  - locust/import.py
  - locust/locust.conf
  - /helm
```

The above settings will make the listed single files and all files in the `helm` directory and its subdirectories
available in your task. Files can be listed with or without a starting `/`, it will be handled as absolute path for both
cases.

This setup is done by using an `initcontainer` for the scheduled Kubernetes Job which prepares the `emptyDir` volume
mounted to the Kubernetes Job. Within the Job itself, the files will be available within the `keptn` folder. The naming
of the files and the location will be preserved.

When using these files in your container command, please make sure to reference them by prepending the `keptn` path.
E.g.:

```yaml
cmd:
  - locust
args:
  - '--config'
  - /keptn/locust/locust.conf
  - '-f'
  - /keptn/locust/basic.py
```

### Silent mode

Actions can be run in silent mode, meaning no `.started/.finished` events will be sent by the job-executor-service. This
is particular useful when not matching `.triggered` events but e.g. `.finished` events where responding
with `.started/.finished` events does not make sense. To enable silent mode simply set it to true for the corresponding
action. By default silent mode is disabled for each action.

```yaml
actions:
  - name: "Run locust"
    silent: true
```

### Resource quotas

The `initcontainer` and the `job` container will use the default resource quotas defined as environment variables. They
can be set in [`deploy/service.yaml`](deploy/service.yaml):

```yaml
- name: DEFAULT_RESOURCE_LIMITS_CPU
  value: "1"
- name: DEFAULT_RESOURCE_LIMITS_MEMORY
  value: "512Mi"
- name: DEFAULT_RESOURCE_REQUESTS_CPU
  value: "50m"
- name: DEFAULT_RESOURCE_REQUESTS_MEMORY
  value: "128Mi"
```

or for helm in [`helm/templates/configmap.yaml`](helm/templates/configmap.yaml):

```yaml
default_resource_limits_cpu: "1"
default_resource_limits_memory: "512Mi"
default_resource_requests_cpu: "50m"
default_resource_requests_memory: "128Mi"
```

The default resource quotas can be easily overwritten for each task. Add the following block to the configuration on
task level:

```yaml
tasks:
  - name: "Run locust tests"
    ...
    resources:
      limits:
        cpu: 1
        memory: 512Mi
      requests:
        cpu: 50m
        memory: 128Mi
```

Now each job that gets spawned for the task will have the configured resource quotas. There is no need to specify all
values, as long as the configuration makes sense for kubernetes. E.g. the following configuration

```yaml
tasks:
  - name: "Run locust tests"
    ...
    resources:
      limits:
        cpu: 1
      requests:
        cpu: 50m
```

would result in resource quotas for `cpu`, but in none for `memory`. If the `resources` block is present
(even if empty), all default resource quotas are ignored for this task.

### Poll duration

The default settings allow a job to run for 5 min until the job executor service cancels the task execution. The default
value can be overwritten for each task and is declared as seconds. The setting below would result in a poll duration of
20 minutes for this specific task:

```yaml
tasks:
  - name: "Run locust tests"
    ...
    maxPollDuration: 1200
```

### Job namespace

By default the jobs run in the `keptn` namespace. This can be configured with the `JOB_NAMESPACE` environment variable.
If you want to run your jobs in a different namespace than the job executor runs in, make sure a kubernetes role is
configured so that the job executor can deploy jobs to it.

In addition, for each task the default namespace can be overwritten in the following way:

```yaml
tasks:
  - name: "Run locust tests"
    ...
    namespace: carts
```

### Job Image Pull Policy
By default the image for the tasks will be pulled according to 
[kubernetes pull policy defaults](https://kubernetes.io/docs/concepts/containers/images/#imagepullpolicy-defaulting).

It's possible to override the pull policy by specifying the desired value in the task:

```yaml
tasks:
  - name: "Run locust tests"
    files:
      - locust/basic.py
      - locust/import.py
      - locust/locust.conf
    image: "locustio/locust"
    imagePullPolicy: "Always"
    cmd:
      - locust
    args:
      - '--config'
      - /keptn/locust/locust.conf
      - '-f'
      - /keptn/locust/basic.py
      - '--host'
      - $(HOST)
```

Allowed values for image pull policy are the same as the [ones accepted by kubernetes](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy)

Note: the job executor service does not perform any validation on the image pull policy value. We delegate any validation
to kubernetes api server.

### Send start/finished event if the job config.yaml can't be found

By default, the job executor service does not send any started/finished events if can't find its `config.yaml` in the
keptn repository. In the case it is desired that the job executor service sends a start/finished event with the
respective error, just set the following environment variable on the pod to true:

```yaml
ALWAYS_SEND_FINISHED_EVENT = true
```

### Additional Event Data

In some cases it is required that the events returned by the job executor service contains additional data, so that the
following tasks defined in the keptn sequence can do their job. E.g. the lighthouse service needs a `start` and `end`
timestamp to do an evaluation.

The job executor service currently adds the following data to specific event types:

* Incoming `sh.keptn.event.test.triggered`
    * Outgoing `sh.keptn.event.test.finished` events contain a `start` and `end` timestamp, marking the beginning and
      the end time of the job responsible for handling the event
      ```json
      {
        "type": "sh.keptn.event.test.finished",
        "data": {
          "project": "sockshop",
          "stage": "dev",
          "service": "carts",
          "result": "pass",
          "status": "succeeded",
          "test": {
            "start": "2021-08-24T16:10:25+00:00",
            "end": "2021-08-24T16:10:30+00:00"
          }
        }
      }
      ```

### Remote Control Plane

If you are using the service in a remote control plane setup make sure the distributor is configured to forward all
events used in the `job/config.yaml`. Just edit the `PUBSUB_TOPIC` environment variable in the distributor deployment
configuration to fit your needs.

### Job clean-up

Jobs objects are kept in kubernetes after completion to allow checking for status or logs inspections/retrieval.
This is not always desirable so kubernetes allows for [automatic clean-up of finished jobs](https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/)
using `ttlSecondsAfterFinished` property in the job spec.

Jobs created by the executor service will still be available for a time after (successful or failed) completion.
The default value of the time-to-live (TTL) for completed jobs is `21600` seconds (6 hours).

In order to set a different TTL for jobs add the `ttlSecondsAfterFinished` property in the task definition, for example:

```yaml
tasks:
  - name: "Run locust tests"
    files:
      - locust/basic.py
      - locust/import.py
      - locust/locust.conf
    image: "locustio/locust"
    cmd:
      - locust
    args:
      - '--config'
      - /keptn/locust/locust.conf
      - '-f'
      - /keptn/locust/basic.py
      - '--host'
      - $(HOST)
    # the corresponding job for this task will be cleaned up 10 minutes (600 seconds) after completion
    ttlSecondsAfterFinished: 600
```

## How to validate a job configuration

`job-lint` is a simple cli tool that validates any given job configuration file and shows possible errors.

```shell
./job-lint test-data/config.yaml

2021/07/13 16:18:49 config ../test-data/config.yaml is valid
```

For each release beginning with `0.1.3` compatible binaries are attached.

## Endless Possibilities

* Run the helm service as a kubernetes job and limit the permissions of it by assigning a different service account to
  it.
* Execute infrastructure as code with keptn (e.g. run terraform, pulumi, etc)
* Run kaniko inside the cluster and build and push your images. Suddenly, keptn turns into your CI
* ...

