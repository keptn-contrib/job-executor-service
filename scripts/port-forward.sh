#!/bin/bash

set -e

k3s kubectl -n keptn port-forward deployment/keptn-generic-job-service 8080:8080