#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/..
helm lint ${ROOT_PATH}/charts/arbiter
if [ $? -ne 0 ]; then
	exit 1
fi

echo "helm verified"
