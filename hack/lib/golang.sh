#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function go::build_binary() {
	local -r targets=($@)

	IFS="," read -ra platforms <<<"${BUILD_PLATFORMS:-}"
	if [[ ${#platforms[@]} -eq 0 ]]; then
		read -ra platforms <<<$(go::host_platform)
	fi

	echo "Go version: $(go version)"

	LDFLAGS+=" $(version::ldflags)"

	for platform in "${platforms[@]}"; do
		for target in "${targets[@]}"; do
			echo "Start building ${target} for ${platform}:"
			go::build_binary_for_platform "${target}" "${platform}"
		done
	done
}

function go::build_binary_for_platform() {
	local -r target=$1
	local -r platform=$2
	local -r os=${platform%/*}
	local -r arch=${platform##*/}

	set -x
	GO111MODULE=on CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -v \
		-ldflags "${LDFLAGS:-}" \
		-o "_output/bin/${platform}/$target" \
		"cmd/$target/main.go"
	set +x
}

function go::host_platform() {
	echo "$(go env GOHOSTOS)/$(go env GOHOSTARCH)"
}

function go::verify_go_version() {
	if [[ -z "$(command -v go)" ]]; then
		echo "go is not installed"
		exit 1
	fi

	local go_version
	IFS=" " read -ra go_version <<<"$(GOFLAGS='' go version)"
	local minimum_go_version
	minimum_go_version=go1.18.3
	if [[ ${minimum_go_version} != $(echo -e "${minimum_go_version}\n${go_version[2]}" | sort -s -t. -k 1,1 -k 2,2n -k 3,3n | head -n1) && ${go_version[2]} != "devel" ]]; then
		echo "go version ${go_version[*]} is too old. Minimum required version is ${minimum_go_version}"
		exit 1
	fi
}

function go::create_gopath_tree() {
	local go_pkg_dir="${BUILD_GOPATH}/src/${PACKAGE_NAME}"
	local go_pkg_basedir
	go_pkg_basedir=$(dirname "${go_pkg_dir}")

	mkdir -p "${go_pkg_basedir}"

	if [[ ! -e ${go_pkg_dir} || "$(readlink "${go_pkg_dir}")" != "${ROOT_PATH}" ]]; then
		ln -snf "${ROOT_PATH}" "${go_pkg_dir}"
	fi
}

function go::setup_env() {
	go::verify_go_version

	go::create_gopath_tree

	export GOPATH="${BUILD_GOPATH}"
	export GOCACHE="${BUILD_GOPATH}/cache"

	export PATH="${BUILD_GOPATH}/bin:${PATH}"

	GOROOT=$(go env GOROOT)
	export GOROOT

	unset GOBIN

	cd "$BUILD_GOPATH/src/$PACKAGE_NAME"
}
