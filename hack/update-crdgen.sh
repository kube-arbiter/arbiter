#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_PATH}/hack/lib/init.sh"

echo "Update CRD starting..."
echo "tools installing..."

go::setup_env

GO111MODULE=on go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0

echo "Generating with controller-gen..."

controller-gen crd paths=./pkg/apis/... output:crd:dir=./manifests/crds

echo "Update CRD done."
