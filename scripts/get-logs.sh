#!/bin/bash

set -e

k3s kubectl -n keptn logs `k3s kubectl get pods -n keptn --selector=run=keptn-generic-job-service -o jsonpath='{.items[*].metadata.name}'` keptn-generic-job-service -f