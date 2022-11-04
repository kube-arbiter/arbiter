#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_PATH}/hack/lib/init.sh"

go::setup_env

echo "installing golangci-lint"
go install github.com/golangci/golangci-lint/cmd/golangci-lint

cd "${ROOT_PATH}"

echo "running golangci-lint --fix"

golangci-lint run --fix
