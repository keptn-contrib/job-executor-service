#!/bin/bash

set -e

docker build . -f Dockerfile -t keptncontrib/job-executor-service:latest
docker push keptncontrib/job-executor-service:latest

docker build . -f initcontainer.Dockerfile -t keptncontrib/job-executor-service-initcontainer:latest
docker push keptncontrib/job-executor-service-initcontainer:latest

k3s kubectl -n keptn delete deployment job-executor-service
k3s kubectl -n keptn apply -f deploy/service.yaml -n keptn