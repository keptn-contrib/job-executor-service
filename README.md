# Job Executor Service

![GitHub release (latest by date)](https://img.shields.io/github/v/release/keptn-sandbox/job-executor-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/keptn-sandbox/job-executor-service)](https://goreportcard.com/report/github.com/keptn-sandbox/job-executor-service)

- [Job Executor Service](#job-executor-service)
    - [Why?](#why)
    - [How?](#how)
        - [Specifying the working directory](#specifying-the-working-directory)
        - [Event Matching](#event-matching)
        - [Kubernetes Job](#kubernetes-job)
        - [Kubernetes Job Environment Variables](#kubernetes-job-environment-variables)
            - [From Events](#from-events)
            - [From Kubernetes Secrets](#from-kubernetes-secrets)
            - [From string literal](#from-string-literal)
        - [File Handling](#file-handling)
        - [Silent mode](#silent-mode)
        - [Resource quotas for jobs](#resource-quotas-for-jobs)
        - [Poll duration for jobs](#poll-duration-for-jobs)
        - [Remote Control Plane](#remote-control-plane)
    - [How to validate a job configuration](#how-to-validate-a-job-configuration)
    - [Endless Possibilities](#endless-possibilities)
    - [Credits](#credits)
    - [Compatibility Matrix](#compatibility-matrix)
    - [Installation](#installation)
        - [Deploy in your Kubernetes cluster](#deploy-in-your-kubernetes-cluster)
        - [Up- or Downgrading](#up--or-downgrading)
        - [Uninstall](#uninstall)
    - [Development](#development)
        - [Common tasks](#common-tasks)
        - [Testing Cloud Events](#testing-cloud-events)
    - [Automation](#automation)
        - [GitHub Actions: Automated Pull Request Review](#github-actions-automated-pull-request-review)
        - [GitHub Actions: Unit Tests](#github-actions-unit-tests)
        - [GH Actions/Workflow: Build Docker Images](#gh-actionsworkflow-build-docker-images)
    - [How to release a new version of this service](#how-to-release-a-new-version-of-this-service)
    - [License](#license)

This Keptn service introduces a radical new approach to running tasks with keptn. It provides the means to run any
container as a Kubernetes Job orchestrated by keptn.

## Why?

The job-executor-service aims to tackle several current pain points with the current approach of services running in the
keptn ecosystem:

| Problem | Solution |
|----------------|----------------|
| Keptn services are constantly running while listening for cloud events coming in over NATS. This consumes unnecessary resources on the Kubernetes cluster. | By running the defined keptn tasks as short-lived workloads (Kubernetes Jobs), they just consume resources while the task is executed. |
| Whenever some new functionality should be triggered by keptn, a new keptn service needs to be written. It usually wraps the functionality of the wanted framework and executes it under the hood. The downside: The code of the new keptn service needs to be written and maintained. This effort scales linearly with the amount of services. | This service can execute any framework with just a few lines of yaml configuration. No need to write or maintain any new code.  |
| Keptn services usually filter for a static list of events the trigger the included functionality. This is not configurable. Whenever the service should listen to a new event, the code of the service needs to be changed. | The Job Executor Service provides the means to trigger a task execution for any keptn event. This is done by matching a jsonpath to the received event payload.  |
| Keptn services are usually opinionized on how your framework execution looks like. E.g. the locust service just executes three different (statically named) files depending on the test strategy in the shipyard. It is not possible to write tests consisting of multiply files. | This service provides the possibility to write any specified file from the keptn git repository into a mounted folder (`/keptn`) of the Kubernetes job. This is done by a initcontainer running before the specified image. |
| Support for new functionality in keptn needs to be added to each keptn service individually. E.g. the new secret functionality needs to be included into all of the services running in the keptn execution plane. This slows down the availability of this new feature. | The Job Executor Service is a single service which provides the means to run any workload orchestrated by keptn. So, it is possible to support new functionality of keptn just once in this service - and all workloads profit from it. E.g. in the case of the secret functionality, one just needs to support it in this service and suddenly all the triggered Kubernetes Jobs have the correct secrets attached to it as environment variables. |

## How?

Just put a file into the keptn git repository (in folder `<service>/job/config.yaml`) to specify

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
      - name: "Run locust smoke tests"
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

### Specifying the working directory

Since all files are hosted by default under `/keptn` and some tools only operate on the current working directory, it is
also possible to switch the working directory of the container. This can be achieved by setting the `workingDirectory`
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
        workingDirectory: "/bin"
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
  - name: "Run locust smoke tests"
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

#### From Kubernetes Secrets

The following configuration looks up a kubernetes secret with the name `locust-secret` and all key/value pairs of the
secret will be available as separate environment variables in the job.

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

#### From string literal

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

### Resource quotas for jobs

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
  - name: "Run locust smoke tests"
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
  - name: "Run locust smoke tests"
    ...
    resources:
      limits:
        cpu: 1
      requests:
        cpu: 50m
```

would result in resource quotas for `cpu`, but in none for `memory`. If the `resources` block is present
(even if empty), all default resource quotas are ignored for this task.

### Poll duration for jobs

The default settings allow a job to run for 5 min until the job executor service cancels the task execution. The default
value can be overwritten for each task and is declared as seconds. The setting below would result in a poll duration of
20 minutes for this specific task:

```yaml
tasks:
  - name: "Run locust smoke tests"
    ...
    maxPollDuration: 1200
```

### Remote Control Plane

If you are using the service in a remote control plane setup make sure the distributor is configured to forward all
events used in the `job/config.yaml`. Just edit the `PUBSUB_TOPIC` environment variable in the distributor deployment
configuration to fit your needs.

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

## Credits

The credits of this service heavily go to @thschue and @yeahservice who originally came up with this idea. :rocket:

## Compatibility Matrix

| Keptn Version | [Job-Executor-Service Docker Image](https://hub.docker.com/r/keptnsandbox/job-executor-service/tags) | Config version |
| :-----------: | :--------------------------------------------------------------------------------------------------: | :------------: |
|     0.8.3     |                               keptnsandbox/job-executor-service:0.1.0                                |       -        |
|     0.8.3     |                               keptnsandbox/job-executor-service:0.1.1                                |       -        |
|     0.8.4     |                               keptnsandbox/job-executor-service:0.1.2                                |       v1       |
|     0.8.6     |                               keptnsandbox/job-executor-service:0.1.3                                |       v2       |

## Installation

The *job-executor-service* can be installed as a part of [Keptn's uniform](https://keptn.sh).

### Deploy in your Kubernetes cluster

To deploy the current version of the *job-executor-service* in your Keptn Kubernetes cluster, apply
the [`deploy/service.yaml`](deploy/service.yaml) file:

```console
kubectl apply -f deploy/service.yaml
```

This should install the `job-executor-service` together with a Keptn `distributor` into the `keptn` namespace, which you
can verify using

```console
kubectl -n keptn get deployment job-executor-service -o wide
kubectl -n keptn get pods -l run=job-executor-service
```

### Up- or Downgrading

Adapt and use the following command in case you want to up- or downgrade your installed version (specified by
the `$VERSION` placeholder):

```console
kubectl -n keptn set image deployment/job-executor-service job-executor-service=keptnsandbox/job-executor-service:$VERSION --record
```

### Uninstall

To delete a deployed *job-executor-service*, use the file `deploy/*.yaml` files from this repository and delete the
Kubernetes resources:

```console
kubectl delete -f deploy/service.yaml
```

## Development

Development can be conducted using any GoLang compatible IDE/editor (e.g., Jetbrains GoLand, VSCode with Go plugins).

It is recommended to make use of branches as follows:

* `master` contains the latest potentially unstable version
* `release-*` contains a stable version of the service (e.g., `release-0.1.0` contains version 0.1.0)
* create a new branch for any changes that you are working on, e.g., `feature/my-cool-stuff` or `bug/overflow`
* once ready, create a pull request from that branch back to the `master` branch

When writing code, it is recommended to follow the coding style suggested by
the [Golang community](https://github.com/golang/go/wiki/CodeReviewComments).

### Common tasks

* Build the binary: `go build -ldflags '-linkmode=external' -v -o job-executor-service`
* Run tests: `go test -race -v ./...`
* Build the docker image: `docker build . -t keptnsandbox/job-executor-service:dev` (Note: Ensure that you use the
  correct DockerHub account/organization)
* Run the docker image locally: `docker run --rm -it -p 8080:8080 keptnsandbox/job-executor-service:dev`
* Push the docker image to DockerHub: `docker push keptnsandbox/job-executor-service:dev` (Note: Ensure that you use the
  correct DockerHub account/organization)
* Deploy the service using `kubectl`: `kubectl apply -f deploy/`
* Delete/undeploy the service using `kubectl`: `kubectl delete -f deploy/`
* Watch the deployment using `kubectl`: `kubectl -n keptn get deployment job-executor-service -o wide`
* Get logs using `kubectl`: `kubectl -n keptn logs deployment/job-executor-service -f`
* Watch the deployed pods using `kubectl`: `kubectl -n keptn get pods -l run=job-executor-service`
* Deploy the service
  using [Skaffold](https://skaffold.dev/): `skaffold run --default-repo=your-docker-registry --tail` (Note:
  Replace `your-docker-registry` with your DockerHub username; also make sure to adapt the image name
  in [skaffold.yaml](skaffold.yaml))

### Testing Cloud Events

We have dummy cloud-events in the form of [RFC 2616](https://ietf.org/rfc/rfc2616.txt) requests in
the [test-events/](test-events/) directory. These can be easily executed using third party plugins such as
the [Huachao Mao REST Client in VS Code](https://marketplace.visualstudio.com/items?itemName=humao.rest-client).

## Automation

### GitHub Actions: Automated Pull Request Review

This repo uses [reviewdog](https://github.com/reviewdog/reviewdog) for automated reviews of Pull Requests.

You can find the details in [.github/workflows/reviewdog.yml](.github/workflows/reviewdog.yml).

### GitHub Actions: Unit Tests

This repo has automated unit tests for pull requests.

You can find the details in [.github/workflows/tests.yml](.github/workflows/tests.yml).

### GH Actions/Workflow: Build Docker Images

This repo uses GH Actions and Workflows to test the code and automatically build docker images.

Docker Images are automatically pushed based on the configuration done in [.ci_env](.ci_env) and the
two [GitHub Secrets](https://github.com/keptn-sandbox/job-executor-service/settings/secrets/actions)

* `REGISTRY_USER` - your DockerHub username
* `REGISTRY_PASSWORD` - a DockerHub [access token](https://hub.docker.com/settings/security) (alternatively, your
  DockerHub password)

## How to release a new version of this service

It is assumed that the current development takes place in the master branch (either via Pull Requests or directly).

To make use of the built-in automation using GH Actions for releasing a new version of this service, you should

* branch away from master to a branch called `release-x.y.z` (where `x.y.z` is your version),
* write release notes in the [releasenotes/](releasenotes/) folder,
* update the compatibility matrix,
* check the output of GH Actions builds for the release branch,
* verify that your image was built and pushed to DockerHub with the right tags,
* update the image tags for `job-executor-service` and `job-executor-service-initcontainer`
  in [`deploy/service.yaml`](deploy/service.yaml), [`helm/Chart.yaml`](helm/Chart.yaml),
  [`helm/values.yaml`](helm/values.yaml), [`helm/templates/configmap.yaml`](helm/templates/configmap.yaml) and
  the `app.kubernetes.io/version` in [`deploy/service.yaml`](deploy/service.yaml)
* test your service against a working Keptn installation.

If any problems occur, fix them in the release branch and test them again.

Once you have confirmed that everything works and your version is ready to go, you should

* create a new release on the release branch using
  the [GitHub releases page](https://github.com/keptn-sandbox/job-executor-service/releases), and
* merge any changes from the release branch back to the master branch.

## License

Please find more information in the [LICENSE](LICENSE) file.
