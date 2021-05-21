#!/bin/bash

set -e

k3s kubectl -n keptn port-forward deployment/job-executor-service 8080:8080