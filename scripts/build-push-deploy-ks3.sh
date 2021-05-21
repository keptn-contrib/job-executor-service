#!/bin/bash

set -e

docker build . -f Dockerfile -t didiladi/job-executor-service:latest
docker push didiladi/job-executor-service:latest

docker build . -f initcontainer.Dockerfile -t didiladi/job-executor-service-initcontainer:latest
docker push didiladi/job-executor-service-initcontainer:latest

k3s kubectl -n keptn delete deployment job-executor-service
k3s kubectl -n keptn apply -f deploy/service.yaml -n keptn