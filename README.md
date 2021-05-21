# Job Executor Service
![GitHub release (latest by date)](https://img.shields.io/github/v/release/keptn-sandbox/job-executor-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/keptn-sandbox/job-executor-service)](https://goreportcard.com/report/github.com/keptn-sandbox/job-executor-service)

(naming not final)

This Keptn service introduces a radical new approach to running tasks with keptn. It provides the means
to run any container as a Kubernetes Job orchestrated by keptn.

## Why?

The job-executor-service aims to tackle several current pain points with the current approach of services running in the keptn ecosystem:

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
actions:
  - name: "Run locust"
    event: "sh.keptn.event.test.triggered"
    jsonpath:
      property: "$.test.teststrategy" 
      match: "locust"
    tasks:
      - name: "Run locust smoke tests"
        files: 
          - locust/basic.py
          - locust/import.py
        image: "locustio/locust"
        cmd: "locust -f /keptn/locust/basic.py"
```

### Event Matching

The configuration located in `<service>/job/config.yaml` contains the following section:

```
    jsonpath:
      property: "$.test.teststrategy" 
      match: "locust"
```

If the service receives an event which matches the jsonpath match expression, the specified tasks are executed. E.g. the 
following cloud event would match the jsonpath above:

```
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
        teststrategy": "locust"
      }
    }
  }
```

### Kubernetes Job

The configuration contains the following section:

```
    tasks:
      - name: "Run locust smoke tests"
        files: 
          - locust/basic.py
          - locust/import.py
        image: "locustio/locust"
        cmd: "locust -f /keptn/locust/basic.py"
```

It contains the tasks which should be executed as Kubernetes job. The service schedules a different job for each of these
tasks in the order, they are listed within the config. The service waits for the successful execution of all of the tasks 
to respond with a `StatusSucceeded` finished event. When one of the events fail, it responds with `StatusErrored` 
finished cloud event. 

### File Handling

Files can be added to your running tasks by specifying them in the `files` section of your tasks:

```
        files: 
          - locust/basic.py
          - locust/import.py
```

This is done by using an `initcontainer` for the scheduled Kubernetes Job which prepares the `èmptyDir` volume mounted to 
the Kubernetes Job. Within the Job itself, the files will be available within the `keptn` folder. The naming of the files 
and the location will be preserved.

When using these files in your comtainer command, please make sure to reference them by prepending the `keptn` path. E.g.:

```
        cmd: "locust -f /keptn/locust/locustfile.py"
```


## Endless Possibilities

* Run the helm service as a kubernetes job and limit the permissions of it by assigning a different service account to it.
* Execute infrastructure as code with keptn (e.g. run terraform, pulumi, etc)
* Run kaniko inside the cluster and build and push your images. Suddenly, keptn turns into your CI
* ...

## Credits

The credits of this service heavily go to @thschue and @augustin-dt who originally came up with this idea. :rocket:

## Compatibility Matrix

*Please fill in your versions accordingly*

| Keptn Version    | [Job-Executor-Service Docker Image](https://hub.docker.com/r/didiladi/job-executor-service/tags) |
|:----------------:|:----------------------------------------:|
|       0.8.2      | didiladi/job-executor-service:latest |

## Installation

The *job-executor-service* can be installed as a part of [Keptn's uniform](https://keptn.sh).

### Deploy in your Kubernetes cluster

To deploy the current version of the *job-executor-service* in your Keptn Kubernetes cluster, apply the [`deploy/service.yaml`](deploy/service.yaml) file:

```console
kubectl apply -f deploy/service.yaml
```

This should install the `job-executor-service` together with a Keptn `distributor` into the `keptn` namespace, which you can verify using

```console
kubectl -n keptn get deployment job-executor-service -o wide
kubectl -n keptn get pods -l run=job-executor-service
```

### Up- or Downgrading

Adapt and use the following command in case you want to up- or downgrade your installed version (specified by the `$VERSION` placeholder):

```console
kubectl -n keptn set image deployment/job-executor-service job-executor-service=didiladi/job-executor-service:$VERSION --record
```

### Uninstall

To delete a deployed *job-executor-service*, use the file `deploy/*.yaml` files from this repository and delete the Kubernetes resources:

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

When writing code, it is recommended to follow the coding style suggested by the [Golang community](https://github.com/golang/go/wiki/CodeReviewComments).

### Common tasks

* Build the binary: `go build -ldflags '-linkmode=external' -v -o job-executor-service`
* Run tests: `go test -race -v ./...`
* Build the docker image: `docker build . -t didiladi/job-executor-service:dev` (Note: Ensure that you use the correct DockerHub account/organization)
* Run the docker image locally: `docker run --rm -it -p 8080:8080 didiladi/job-executor-service:dev`
* Push the docker image to DockerHub: `docker push didiladi/job-executor-service:dev` (Note: Ensure that you use the correct DockerHub account/organization)
* Deploy the service using `kubectl`: `kubectl apply -f deploy/`
* Delete/undeploy the service using `kubectl`: `kubectl delete -f deploy/`
* Watch the deployment using `kubectl`: `kubectl -n keptn get deployment job-executor-service -o wide`
* Get logs using `kubectl`: `kubectl -n keptn logs deployment/job-executor-service -f`
* Watch the deployed pods using `kubectl`: `kubectl -n keptn get pods -l run=job-executor-service`
* Deploy the service using [Skaffold](https://skaffold.dev/): `skaffold run --default-repo=your-docker-registry --tail` (Note: Replace `your-docker-registry` with your DockerHub username; also make sure to adapt the image name in [skaffold.yaml](skaffold.yaml))


### Testing Cloud Events

We have dummy cloud-events in the form of [RFC 2616](https://ietf.org/rfc/rfc2616.txt) requests in the [test-events/](test-events/) directory. These can be easily executed using third party plugins such as the [Huachao Mao REST Client in VS Code](https://marketplace.visualstudio.com/items?itemName=humao.rest-client).

## Automation

### GitHub Actions: Automated Pull Request Review

This repo uses [reviewdog](https://github.com/reviewdog/reviewdog) for automated reviews of Pull Requests. 

You can find the details in [.github/workflows/reviewdog.yml](.github/workflows/reviewdog.yml).

### GitHub Actions: Unit Tests

This repo has automated unit tests for pull requests. 

You can find the details in [.github/workflows/tests.yml](.github/workflows/tests.yml).

### GH Actions/Workflow: Build Docker Images

This repo uses GH Actions and Workflows to test the code and automatically build docker images.

Docker Images are automatically pushed based on the configuration done in [.ci_env](.ci_env) and the two [GitHub Secrets](https://github.com/keptn-sandbox/job-executor-service/settings/secrets/actions)
* `REGISTRY_USER` - your DockerHub username
* `REGISTRY_PASSWORD` - a DockerHub [access token](https://hub.docker.com/settings/security) (alternatively, your DockerHub password)

## How to release a new version of this service

It is assumed that the current development takes place in the master branch (either via Pull Requests or directly).

To make use of the built-in automation using GH Actions for releasing a new version of this service, you should

* branch away from master to a branch called `release-x.y.z` (where `x.y.z` is your version),
* write release notes in the [releasenotes/](releasenotes/) folder,
* check the output of GH Actions builds for the release branch, 
* verify that your image was built and pushed to DockerHub with the right tags,
* update the image tags in [deploy/service.yaml], and
* test your service against a working Keptn installation.

If any problems occur, fix them in the release branch and test them again.

Once you have confirmed that everything works and your version is ready to go, you should

* create a new release on the release branch using the [GitHub releases page](https://github.com/keptn-sandbox/job-executor-service/releases), and
* merge any changes from the release branch back to the master branch.

## License

Please find more information in the [LICENSE](LICENSE) file.
