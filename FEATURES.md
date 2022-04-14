
## Features

- [Features](#Features)
    - [Getting started](#getting-started)
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
    - [Job security context](#job-security-context)
    - [Job image pull policy](#job-image-pull-policy)
    - [Restrict job images](#restrict-job-images)
    - [Send start/finished event if the job config.yaml can't be found](#send-startfinished-event-if-the-job-configyaml-cant-be-found)
    - [Additional Event Data](#additional-event-data)
    - [Remote Control Plane](#remote-control-plane)
    - [Job clean-up](#job-clean-up)


### Getting started

To get started with job-executor-service, please follow the [Quickstart](README.md#quickstart).


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

By default, the following environment variables are automatically available:
* `$(KEPTN_PROJECT)` - project name from Cloud Event (`.data.project`)
* `$(KEPTN_SERVICE)` - project name from Cloud Event (`.data.service`)
* `$(KEPTN_STAGE)` - project name from Cloud Event (`.data.stage`)
* For every label of the Cloud Event, we provide `$(LABELS_KEY)` making the key uppercase and transforming spacing/hyphens in underscores (e.g., the label `build-id` can be accessed using `$(LABELS_BUILD_ID)`) 


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

### Job security context

By default the jobs will use the default security context, which was specified at the time of the installation of the 
job-executor-service. The job configuration provides a way to overwrite this context on a task level:

```yaml
tasks:
  - name: "Run as different user"
    image: "alpine"
    securityContext:
      runAsUser: 7000
      runAsGroup: 9000
    cmd:
      - id
  - name: "Allow modifications to root FS"
    image: "alpine"
    securityContext:
      readOnlyRootFilesystem: false
    cmd:
      - sh
    args:
      - -c
      - "echo WriteableFilesystem > test.txt"
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

### Restrict job images

During the installation of the *job-executor-service* a comma separated allow-list of images can be specified
(`jobConfig.allowedImageList`), which is used to restrict the amount of images that can be used in job workloads. 
This allow-list also supports simple globs like `docker.io/*`.

For example, to allow only images from a specific user from `docker.io` and images from a custom registry, the installation
of the *job-executor-service* can be adapted as follows:
```bash
KEPTN_API_PROTOCOL=http # or https
KEPTN_API_HOST=<INSERT-YOUR-HOSTNAME-HERE> # e.g., 1.2.3.4.nip.io
 KEPTN_API_TOKEN=<INSERT-YOUR-KEPTN-API-TOKEN-HERE>

TASK_SUBSCRIPTION=sh.keptn.event.remote-task.triggered
ALLOWED_IMAGE_LIST="docker.io/my-user/*,custom.registry.io/*"

helm upgrade --install --create-namespace -n <NAMESPACE> \
  job-executor-service https://github.com/keptn-contrib/job-executor-service/releases/download/<VERSION>/job-executor-service-<VERSION>.tgz \
 --set jobConfig.allowedImageList=${ALLOWED_IMAGE_LIST},remoteControlPlane.topicSubscription=${TASK_SUBSCRIPTION},remoteControlPlane.api.protocol=${KEPTN_API_PROTOCOL},remoteControlPlane.api.hostname=${KEPTN_API_HOST},remoteControlPlane.api.token=${KEPTN_API_TOKEN}
```


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

**Note:** `ttlSecondsAfterFinished` relies on setting the [same property](https://kubernetes.io/docs/concepts/workloads/controllers/job/#ttl-mechanism-for-finished-jobs)
in kubernetes job workloads spec. The TTL controller (alpha from Kubernetes v1.12-v1.20, beta in v1.21-v1.22, GA in v1.23)
will then take care of cleanup.
More information about feature stages in kubernetes (that is what alpha, beta and GA implies in that context) have a look at
the [official kubernetes documentation](https://v1-20.docs.kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/#feature-stages).

