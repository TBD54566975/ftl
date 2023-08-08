#!/bin/bash
set -euo pipefail

# https://www.kubegres.io/doc/getting-started.html
kubectl apply -f https://raw.githubusercontent.com/reactive-tech/kubegres/v1.16/kubegres.yaml
kubectl apply -f pg-cluster.yaml
kubectl apply -f ftl-controller.yaml

