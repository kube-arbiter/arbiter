#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_PATH}/hack/lib/init.sh"

go::setup_env

echo "installing shfmt"
go install mvdan.cc/sh/v3/cmd/shfmt@latest

cd "${ROOT_PATH}"

echo "running shfmt"

if shfmt -f . | grep -v vendor/ | grep -v _output/ | xargs shfmt -l -w -s -d; then
	echo "shfmt verified."
else
	echo "shfmt failed."
	exit 1
fi
