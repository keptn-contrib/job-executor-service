## Output Metrics

Often you will run an existing container / tool with the job executor service which outputs results. Most likely you want to use these metrics in a subsequent Keptn task (eg. a quality gate evaluation). To do so, you will need to push the metrics to a metric backend. This page demonstrates how this can be done.

- [Prometheus Integration](#output-metrics-to-prometheus)
- [Dynatrace Integration](#output-metrics-to-dynatrace)

Rather than directly running your container image, you'll need to wrap it in a "parent script" and run that parent script instead (shell scripts, python scripts, powershell scripts etc.).

The samples below assume you are using Python, but these instructions should provide enough information to adapt to any language you prefer.

### Output Metrics to Prometheus

![output metrics to prometheus](assets/output_metrics/prom_metrics.jpg)

Push metrics from any job executor service job to Prometheus using a [Prometheus Push Gateway](https://prometheus.io/docs/instrumenting/pushing/).

A Python base image with the requests module & the [prometheus client](https://pypi.org/project/prometheus-client/) is required.

You can use [gardnera/python:requests_prometheus_client](https://hub.docker.com/r/gardnera/python/tags) or build your own:
```
FROM python:slim
RUN pip install requests
RUN pip install prometheus-client
```

Use that image in your `job/config.yaml` file:
```
apiVersion: v2
actions:
  - name: "Run Your Tool"
    events:
      - name: "sh.keptn.event.YourEvent.triggered"
    tasks:
      - name: "Execute tool"
        files:
          - /files/app.py
        image: "gardnera/python:requests_prometheus_client"
        cmd: 
          - "python"
        args:
          - "/keptn/files/app.py"
```

Finally, create the `app.py` file using this boilerplate:

```
import json
import requests
from prometheus_client import CollectorRegistry, Counter, Gauge, push_to_gateway
import os

#####################
# Set these values  #
#####################

# The name of this integration. It will form part of the metric name. Eg. infracost
INTEGRATION_NAME = "infracost"

# Your Prometheus Push Gateway endpoint
# eg. "prometheus-pushgateway.monitoring.svc.cluster.local:9091"
PROM_GATEWAY = "prometheus-pushgateway.monitoring.svc.cluster.local:9091"

############################
# End configurable values  #
############################

# These variables are passed to job-executor-service automatically on job startup
# So you can assume they're available
KEPTN_PROJECT = os.getenv("KEPTN_PROJECT", "NULL")
KEPTN_SERVICE = os.getenv("KEPTN_SERVICE", "NULL")
KEPTN_STAGE = os.getenv("KEPTN_STAGE", "NULL")

PROM_LABELS = [
    "ci_platform",
    "keptn_project",
    "keptn_service",
    "keptn_stage"
]

########################
# Do your work here... #
########################

##########################
# PUSH METRICS TO PROM   #
##########################
reg = CollectorRegistry()

# pseudo-code for each metric
# create a new metric, set the labels and value
# this is just a sample, adjust based on your data structures
for metric_name in some_metrics:
  metric_value = some_metrics[name]

  # Create a Prometheus Gauge metric
  g = Gauge(name=f"keptn_{INTEGRATION_NAME}_{metric_name}", documentation='', registry=reg, labelnames=PROM_LABELS)
  # Set the labels and values
  g.labels(
    ci_platform="keptn",
    keptn_project=KEPTN_PROJECT,
    keptn_service=KEPTN_SERVICE,
    keptn_stage=KEPTN_STAGE
  ).set(metric_value)

# Send the metrics to Prometheus Push Gateway
push_to_gateway(gateway=PROM_GATEWAY,job=f"job-executor-service", registry=reg)
```

## Output Metrics to Dynatrace

![dt metrics](assets/output_metrics/dt_metrics.jpg)

Dynatrace offers an API endpoint so you can push metrics into Dynatrace by following the [line format](https://www.dynatrace.com/support/help/extend-dynatrace/extend-metrics/reference/metric-ingestion-protocol):

```
metric.key,dimensions payload
eg.
my.value,ci_platform="keptn",keptn_project="projectA",keptn_service="service1",keptn_stage="dev" 42
```

### Gather Details and Create Secret

The job executor service requires a details of your Dynatrace environment.

In the same namespace as the job executor service, create a secret to hold the `DT_TENANT` and `DT_API_TOKEN` values.

`DT_TENANT` should take the format (no trailing slashes):

- Dynatrace SaaS: `https://{your-environment-id}.live.dynatrace.com`
- Dynatrace Managed: `https://{your-domain}/e/{your-environment-id}`

`DT_API_TOKEN` requires the following permissions:

- `metrics.ingest` (Scope API: v2)

Example:
```
DT_TENANT=https://abc12345.live.dynatrace.com
DT_API_TOKEN=dtc01.******.*****
JES_NAMESPACE=SomeNamespace
kubectl create secret generic dt_details \
--namespace $JES_NAMESPACE \
--from-literal=DT_TENANT=$DT_TENANT \
--from-literal=DT_API_TOKEN
```

### Create job/config.yaml

Reference the above secret in the `job/config.yaml`. By doing so, `DT_TENANT` and `DT_API_TOKEN` become available as environment variables.

For `image`, use any image with Python and the `requests` module installed.

```
apiVersion: v2
actions:
  - name: "Run Your Tool"
    events:
      - name: "sh.keptn.event.YourEvent.triggered"
    tasks:
      - name: "Execute tool"
        env:
          - name: dt_details
            valueFrom: secret
        files:
          - /files/app.py
        image: "gardnera/requests:v0.0.1"
        cmd: 
          - "python"
        args:
          - "/keptn/files/app.py"
```

### Create app.py

The following is a sample Python script. Obviously you will need to adjust for your data and requirements.

```
import requests
import os

#####################
# Set these values  #
#####################

# The name of this integration. It will form part of the metric name. Eg. infracost
INTEGRATION_NAME = "infracost"

############################
# End configurable values  #
############################

# These variables are passed to job-executor-service automatically on job startup
# So you can assume they're available
KEPTN_PROJECT = os.getenv("KEPTN_PROJECT", "NULL")
KEPTN_SERVICE = os.getenv("KEPTN_SERVICE", "NULL")
KEPTN_STAGE = os.getenv("KEPTN_STAGE", "NULL")

# Available due to secret
DT_TENANT = os.getenv("DT_TENANT","NULL")
DT_API_TOKEN = os.getenv("DT_API_TOKEN","NULL")

########################
# Do your work here... #
########################

#################################################################
# Create Dynatrace compatible metrics string                    #
# This is a sample only                                         #
# You will need to solution this based on your data structures  #
#################################################################

# For example
some_list = [{
    "name": "metric1",
    "value": 42
}]

# Assumes you have a data structure with metric_name and metric_value available
metric_string = ""
for datapoint in some_list:
  metric_name = datapoint['name']
  metric_value = datapoint['value']

  metric_line = f"keptn_{INTEGRATION_NAME}_{metric_name},ci_platform=keptn,keptn_project={KEPTN_PROJECT},keptn_service={KEPTN_SERVICE},keptn_stage={KEPTN_STAGE} {metric_value}"
  
  # Add metric line to master list of metrics separated by a newline character
  metric_string += f"{metric_line}\n"
  
###############################
# PUSH METRICS TO DYNATRACE   #
###############################
headers = {
    "Authorization": f"Api-Token {DT_API_TOKEN}",
    "Content-Type": "text/plain; charset=utf-8"
}
dt_response = requests.post(url=f"{DT_TENANT}/api/v2/metrics/ingest",headers=headers, data=metric_string)
print(dt_response.status_code) # should be a 202
print(dt_response.text) # {"linesOK": 1, "linesInvalid": 0, "error": null, "warnings": null}
```