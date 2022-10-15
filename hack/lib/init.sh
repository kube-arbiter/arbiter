#!/usr/bin/env bash

unset CDPATH

ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/../..

source "${ROOT_PATH}/hack/lib/util.sh"
source "${ROOT_PATH}/hack/lib/version.sh"
source "${ROOT_PATH}/hack/lib/golang.sh"
source "${ROOT_PATH}/hack/lib/docker.sh"
