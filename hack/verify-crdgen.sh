#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

hack/update-crdgen.sh

if ! _out="$(git --no-pager diff -I"edited\smanually" --exit-code ./manifests)"; then
	echo "Generated output differs" >&2
	echo "${_out}" >&2
	echo "Verification for CRD generators failed."
	exit 1
fi

echo "Gen for CRD verified."
