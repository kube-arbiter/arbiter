#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_PATH}/hack/lib/init.sh"

go::verify_go_version

bash "$ROOT_PATH/hack/update-vendor.sh"
bash "$ROOT_PATH/hack/update-codegen.sh"
bash "$ROOT_PATH/hack/update-crdgen.sh"
