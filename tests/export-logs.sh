#!/usr/bin/env bash
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/..
LOG_DIR=${LOG_DIR:-"/tmp/arbiter/logs"}

mkdir -p $LOG_DIR
kind export logs --name arbiter-e2e $LOG_DIR
sudo chown -R $USER:$USER $LOG_DIR
