#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

hack/update-vendor.sh

if ! _out="$(git --no-pager diff -I"edited\smanually" --exit-code ./vendor/* go.mod go.sum)"; then
	echo "Generated output differs" >&2
	echo "${_out}" >&2
	echo "Verification for Vendor failed."
	exit 1
fi

echo "Vendor verified."
