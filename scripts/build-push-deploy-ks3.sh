#!/bin/bash

set -e

docker build . -t didiladi/keptn-generic-job-service:latest
docker push didiladi/keptn-generic-job-service:latest
k3s kubectl -n keptn delete deployment keptn-generic-job-service
k3s kubectl -n keptn apply -f deploy/service.yaml -n keptn