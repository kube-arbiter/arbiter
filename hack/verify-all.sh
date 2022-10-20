#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_PATH}/hack/lib/init.sh"

go::verify_go_version

bash "$ROOT_PATH/hack/verify-shfmt.sh"
bash "$ROOT_PATH/hack/verify-vendor.sh"
bash "$ROOT_PATH/hack/verify-codegen.sh"
bash "$ROOT_PATH/hack/verify-crdgen.sh"
bash "$ROOT_PATH/hack/verify-golangci-lint.sh"
bash "$ROOT_PATH/hack/verify-copyright.sh"
bash "$ROOT_PATH/hack/verify-helm.sh"
