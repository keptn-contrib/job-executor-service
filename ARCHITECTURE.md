# Architecture

Within this document the architecture of job-executor-service is described.

## Basics

Before reading this document, we expect that you already have some knowledge about [Keptn](https://keptn.sh).
Primarily, you should be familiar with the following terms and concepts of Keptn:
* Project
* Service
* Shipyard file
* Stage
* Sequence
* Tasks
* Execution-plane vs. control-plane
* Job-executor-service basics

Suggested reads:
* https://keptn.sh/docs/concepts/glossary/
* https://keptn.sh/docs/concepts/delivery/
* https://keptn.sh/docs/concepts/architecture/
* https://engineering.dynatrace.com/blog/a-tool-to-execute-them-all-the-job-executor-service/

In addition, it is useful, but not strictly required, to have a basic understanding of Kubernetes jobs, pods and containers.
As a bare minimum to continue reading, it makes sense to understand what a (Docker) container is and how you can create 
one yourself.

## General Architecture

![](assets/architecture/job-exec-architecture.jpeg)

Job-Executor consists of a `keptn/distributor`, which handles the connection to Keptn control-plane, and the
job-Executor Keptn service itself. Job-Executor needs to be installed within a Kubernetes cluster, and connected to
a Keptn installation.

When a certain Cloud Event is emitted by Keptn, Job-Executor will create a new Kubernetes job within the same Kubernetes
cluster, with the details configured in `job/config.yaml`. This job will consist of an `initcontainer`, which is supposed
to fetch files from the projects config git repo (served by Keptns `configuration-service`), and the actual container
based on the `image` defined in `job/config.yaml`, where a `command` is executed.

## Example Configuration

A major part to understand the architecture is understanding how events are flowing from and to Keptn.

For this, please imagine a project with a simple shipyard file with a single stage `production`, and a sequence `pipeline`
that only consists of one task `test`.

**shipyard.yaml**
```yaml
apiVersion: "spec.keptn.sh/0.2.2"
kind: "Shipyard"
metadata:
  name: "shipyard-pipeline"
spec:
  stages:
    - name: "production"
      sequences:
        - name: "pipeline"
          tasks:
            - name: "test"
```

In addition, please consider the following (simplified) job-executor configuration:
```yaml
apiVersion: v2
actions:
  - name: "Run tests"
    event: "sh.keptn.event.test.triggered"
    files:
      - locust/basic.py
      - locust/locust.conf
    image: "docker.io/locustio/locust"
    cmd: "locust"
    args: ['--config', '/keptn/locust/locust.conf', '-f', '/keptn/locust/basic.py', '--host', '$(HOST)']
```
This configuration basically means:
* When `test.triggered` is emitted from Keptn,
* fetch the files `locust/basic.py` and `locust/locust.conf` from the projects config repo,
* spawn a Kubernetes job with the image `docker.io/locustio/locust`, 
* and run the command `locust --config /keptn/locust/locust.conf -f /keptn/locust/basic.py --host $(HOST)`

*Note*: Locust is a tool for running performance/load-tests

## Full Event flow 

This diagram shows the full event flow with almost all Keptn components involved (based on Keptn 0.12), job-executor-service
itself, as well as Kubernetes and Git.

![](assets/architecture/user-flow-keptn-full.png)

## Event Flow focused on job-executor-centric

This diagram shows the full event flow, but limited to user-facing Keptn components, job-executor-service
itself, as well as Kubernetes and Git.

![](assets/architecture/user-flow-keptn-job-exec.png)
