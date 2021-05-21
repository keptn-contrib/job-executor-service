#!/bin/bash

set -e

k3s kubectl -n keptn logs `k3s kubectl get pods -n keptn --selector=run=job-executor-service -o jsonpath='{.items[*].metadata.name}'` job-executor-service -f