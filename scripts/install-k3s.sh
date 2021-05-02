#!/bin/bash

set -e

/usr/local/bin/k3s-uninstall.sh
curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644

k3s kubectl create namespace keptn

helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
helm install keptn-nats-cluster nats/nats --kubeconfig=/etc/rancher/k3s/k3s.yaml --set stan.nats.url=nats://keptn-nats-cluster -n keptn