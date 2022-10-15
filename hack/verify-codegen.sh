#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

hack/update-codegen.sh

if ! _out="$(git --no-pager diff -I"edited\smanually" --exit-code -- ./pkg/generated)"; then
	echo "Generated output differs" >&2
	echo "${_out}" >&2
	echo "Verification for Code generators failed."
	exit 1
fi

echo "Code Gen for CRD verified."
