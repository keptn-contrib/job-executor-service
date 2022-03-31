# Migration Guide from Generic-Executor-Service to Job-Executor-Service

There are multiple ways on how to migrate from Generic-Executor-Service to Job-Executor-Service, which we will try to 
describe within this document.

Before continuing with reading this document, please

* make sure you have a working Keptn environment with the latest version of job-executor-service installed,
* you are aware of the key concepts of job-executor-service (e.g., bring your own container, write job/config.yaml).

Suggested reads:

* https://keptn.sh/docs/concepts/glossary/
* https://keptn.sh/docs/concepts/delivery/
* https://keptn.sh/docs/concepts/architecture/
* https://engineering.dynatrace.com/blog/a-tool-to-execute-them-all-the-job-executor-service/

## What's still missing

Unfortunately, job-executor-service is not fully compatible with generic-executor-service. The following list provides
an overview of what's still missing.

### Returning errors or follow-up event

While [generic-executor-service supports returning errors or a follow-up event](https://github.com/keptn-sandbox/generic-executor-service#returning-errors-or-follow-up-event),
job-executor-service still lacks this, and will always put the output produced by the command as the `data.message` property
of the `.finished` cloud-event.

Example (using `helm upgrade` within job-executor-service) - output is written to `data.message`:
```json
{
  "data": {
    "message": "Job job-executor-service-job-a9338cab-2462-41e6-9414-599d-1 finished successfully!\n\nLogs:\nRelease \"helloservice\" has been upgraded. Happy Helming!\nNAME: helloservice\nLAST DEPLOYED: Wed Mar 16 13:43:41 2022\nNAMESPACE: podtato-head-qa\nSTATUS: deployed\nREVISION: 18\nTEST SUITE: None\n\n\n",
    "project": "podtato-head",
    "result": "pass",
    "service": "helloservice",
    "stage": "qa",
    "status": "succeeded"
  },
  "id": "7b8f6c80-b887-421a-8c90-cca9b51c2acd",
  "source": "job-executor-service",
  "specversion": "1.0",
  "time": "2022-03-16T13:43:49.785Z",
  "type": "sh.keptn.event.je-deployment.finished",
  "shkeptncontext": "2ebdf6a9-8300-42b0-80ea-ebd54061996e",
  "shkeptnspecversion": "0.2.3",
  "triggeredid": "a9338cab-2462-41e6-9414-599d711652e8"
}
```

**Plan**: There is a plan to include a similar feature in the future - see https://github.com/keptn-contrib/job-executor-service/discussions/129 for a related discussion.

**Recommendation**: For now, we recommend staying with generic-executor-service until this feature is implemented in job-executor-service.

## Migrating HTTP webhook-based setup

In generic-executor-service you can use webhooks by creating a `$event.http` file (where `$event` reflects the Cloud
event type you are listening for), with the following content:

```
POST https://webhook.site/YOURHOOKID
Accept: application/json
Cache-Control: no-cache
Content-Type: application/cloudevents+json

{
  "contenttype": "application/json",
  "deploymentstrategy": "blue_green_service",
  "project": "${data.project}",
  "service": "${data.service}",
  "stage": "${data.stage}",
  "mylabel" : "${data.label.gitcommit}",
  "mytoken" : "${env.testtoken}",
  "shkeptncontext": "${shkeptncontext}",
  "type": "${type}",
  "source": "${source}"
}
```

This can be migrated to [Keptn's built-in webhook-service](https://keptn.sh/docs/0.13.x/integrations/webhooks/). 
Some changes might be needed for the payload, e.g.:

* Convert variables / attribute access from `${data.project}` to `{{.data.project}}` (go templating syntax; remove `$`, use double curly braces `{{`/`}}`, use dot syntax to access attributes)
* Environment variables are no longer available (e.g., `${env.testtoken}` will no longer work). Instead, we recommend [using secrets to include sensitive data](https://keptn.sh/docs/0.13.x/integrations/webhooks/#include-sensitive-data)
* `${shkeptncontext}` becomes `{{.shkeptncontext}}`
* `${type}` becomes `{{.type}}`
* `${source}` becomes `{{.source}}`

## Migrating Bash Scripts

In generic-executor-service you can use bash scripts by creating a `$event.sh` file (where `$event` reflects the Cloud
event type you are listening for), with the following content:

**generic-executor/deployment.triggered.sh**
```bash
#!/bin/bash

# This is a script that will be executed by the Keptn Generic Executor Service for deployment.triggered events (based on the file name deployment.triggered.sh).
# It will be called with a couple of environment variables that are filled with Keptn Event Details, Env-Variables from the Service container as well as labels

echo "This is my deployment.triggered.sh script"
echo "Project = $DATA_PROJECT"
echo "Service = $DATA_SERVICE"
echo "Stage = $DATA_STAGE"
echo "Image = $DATA_CONFIGURATIONCHANGE_VALUES_IMAGE"
echo "DeploymentStrategy = $DATA_DEPLOYMENT_DEPLOYMENTSTRATEGY"
echo "TestToken = $ENV_TESTTOKEN"

# Here i could do whatever I want with these values, e.g: call an external tool :-)
```

In general, this can be achieved by re-using existing `.sh` files in job-executor-service, using a container that provides bash and the necessary tools, and providing all necessary variables within the code.

However, we recommend some additional steps before that:

* Change `$DATA_PROJECT` to `$KEPTN_PROJECT`
* Change `$DATA_SERVICE` to `$KEPTN_SERVICE`
* Change `$DATA_STAGE` to `$KEPTN_STAGE`
* Create secrets for sensitive data in the Kubernetes namespace that job-executor-service is installed (`kubectl -n keptn create secret generic my-super-secret --from-literal="ENV_TESTTOKEN=1234"`)
* Move the script into another folder (optional)

For more information regarding environment variables, please consult [FEATURES.md](../FEATURES.md).

**job/config.yaml**
```yaml
apiVersion: v2
actions:
  - name: "Deploy using bash scripts"
    events:
      - name: "sh.keptn.event.deployment.triggered"
    tasks:
      - name: "Run bash script"
        files:
          - /generic-executor-service/deployment.triggered.sh
        env:
          - name: DATA_CONFIGURATIONCHANGE_VALUES_IMAGE
            value: "$.data.configurationChange.values.image"
            valueFrom: event
          - name: DATA_DEPLOYMENT_DEPLOYMENTSTRATEGY
            value: "$.data.deployment.deploymentstrategy"
            valueFrom: event
          - name: my-super-secret
            valueFrom: secret
        workingDir: "/keptn"
        image: "debian:bookworm-slim"
        cmd: ["bash"]
        args: ["./generic-executor-service/deployment.triggered.sh"]
```

*Note*: We are using `debian:bookworm-slim` as an image here. This is a lightweight debian based image, which does not
provide a lot of tools (it's fine for scripting purpose, e.g., scripting using `echo`, `cat`, `grep` etc..., but not 
for using tools like `curl`, `wget`, etc... which you would have to install).

You can verify this locally by running
```console
docker run -it --rm debian:bookworm-slim bash
```
and pasting your script (or parts of it) into the terminal.


### Troubleshooting

#### Permission Denied

If you get a message like this:
```
bash: line 1: ./generic-executor-service/deployment.triggered.sh: Permission denied
```
you need to make sure that your script has the executable permission in the git upstream repo, e.g.: 
```console
chmod +x generic-executor-service/deployment.triggered.sh
keptn add-resource --project=<PROJECT> --service=<SERVICE> --stage=<STAGE> --resource=generic-executor-service/deployment.triggered.sh --resourceUri=generic-executor-service/deployment.triggered.sh
```


## Migrating Python Scripts

In generic-executor-service you can use python scripts by creating a `$event.py` file (where `$event` reflects the Cloud
event type you are listening for), with the following content:

**deployment.triggered.py**
```python
import os
import sys

# Lets get the first parameter which could potentially be a local file name
methodArg = ""
if len(sys.argv) > 1:
    methodArg = sys.argv[1]

print("This is my generic handler script and I got passed " + methodArg + " as parameter")
print("I also have some env variables, e.g: SHKEPTNCONTEXT=" + os.getenv('SHKEPTNCONTEXT', ""))
print("SOURCE=" + os.getenv('SOURCE',""))
print("PROJECT=" + os.getenv('DATA_PROJECT',""))
```

In general, this can be achieved by re-using existing `.py` files in job-executor-service, using a container that provides python and the necessary tools/packages, and providing all necessary variables within the code.

However, we recommend some additional steps before that:

* Change `DATA_PROJECT` to `KEPTN_PROJECT`
* Change `DATA_SERVICE` to `KEPTN_SERVICE`
* Change `DATA_STAGE` to `KEPTN_STAGE`
* Create secrets for sensitive data in the Kubernetes namespace that job-executor-service is installed (`kubectl create secret ...`)
* Move the script into another folder (optional)

For more information regarding environment variables, please consult [FEATURES.md](../FEATURES.md).

**job/config.yaml**
```yaml
apiVersion: v2
actions:
  - name: "Deploy using bash scripts"
    events:
      - name: "sh.keptn.event.deployment.triggered"
    tasks:
      - name: "Run bash script"
        files:
          - /generic-executor-service/deployment.triggered.py
        env:
          - name: SHKEPTNCONTEXT
            value: "$.shkeptncontext"
            valueFrom: event
          - name: SOURCE
            value: "$.source"
            valueFrom: event
        workingDir: "/keptn"
        image: "python:3.10"
        cmd: ["python3"]
        args: ["./generic-executor-service/deployment.triggered.py"]
```

*Note*: We are using `python:3.11` as an image here. This is a simple image with python3 included, but it does not
provide a lot of tools nor any additional python packages.

You can verify that your script is working by running this locally
```console
docker run -it --rm python:3.10 python3
```
and pasting your script (or parts of it) into the terminal.

