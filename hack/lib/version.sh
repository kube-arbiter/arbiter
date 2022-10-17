#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

version::ldflags() {
	version::git_version
	GIT_COMMIT=$(git rev-parse HEAD)
	if GIT_STATUS=$(git status --porcelain 2>/dev/null) && [[ -z ${GIT_STATUS} ]]; then
		GIT_TREE_STATE="clean"
	else
		GIT_TREE_STATE="dirty"
	fi
	BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

	local -a ldflags
	function add_ld_flag() {
		local key=${1}
		local val=${2}
		ldflags+=(
			"-X 'k8s.io/component-base/version.${key}=${val}'"
		)
	}

	if [[ -n ${BUILD_DATE-} ]]; then
		add_ld_flag "buildDate" "${BUILD_DATE}"
	fi

	if [[ -n ${GIT_COMMIT-} ]]; then
		add_ld_flag "gitCommit" "${GIT_COMMIT}"
		add_ld_flag "gitTreeState" "${GIT_TREE_STATE}"
	fi

	if [[ -n ${GIT_VERSION-} ]]; then
		add_ld_flag "gitVersion" "${GIT_VERSION}"
	fi

	if [[ -n ${GIT_MAJOR-} && -n ${GIT_MINOR-} ]]; then
		add_ld_flag "gitMajor" "${GIT_MAJOR}"
		add_ld_flag "gitMinor" "${GIT_MINOR}"
	fi

	echo "${ldflags[*]-}"
}

version::git_version() {
	GIT_VERSION=$(git describe --tags --dirty --match "v*" --abbrev=14 2>/dev/null || echo "v0.1.0")
}
