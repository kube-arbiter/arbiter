#!/usr/bin/env bash
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/..
LOG_DIR=${LOG_DIR:-"/tmp/arbiter/logs"}

mkdir -p $LOG_DIR
kubectl api-resources --verbs=list -o name | xargs -n 1 kubectl get -A --ignore-not-found=true --show-kind=true >$LOG_DIR/get-all-resources-list.log
kubectl api-resources --verbs=list -o name | xargs -n 1 kubectl get -A -oyaml --ignore-not-found=true --show-kind=true >$LOG_DIR/get-all-resources-yaml.log
kubectl api-resources --verbs=list -o name | xargs -n 1 kubectl describe -A >$LOG_DIR/describe-all-resources.log
kind export logs --name arbiter-e2e $LOG_DIR
sudo chown -R $USER:$USER $LOG_DIR
