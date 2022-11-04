#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_PATH}/hack/lib/init.sh"

go::setup_env
cd "${ROOT_PATH}"

echo "codegen start."
echo "tools install..."

GO111MODULE=on go install k8s.io/code-generator/cmd/deepcopy-gen@v0.23.10
GO111MODULE=on go install k8s.io/code-generator/cmd/register-gen@v0.23.10
GO111MODULE=on go install k8s.io/code-generator/cmd/client-gen@v0.23.10
GO111MODULE=on go install k8s.io/code-generator/cmd/lister-gen@v0.23.10
GO111MODULE=on go install k8s.io/code-generator/cmd/informer-gen@v0.23.10
GO111MODULE=on go install k8s.io/code-generator/cmd/conversion-gen@v0.23.10
GO111MODULE=on go install k8s.io/code-generator/cmd/defaulter-gen@v0.23.10

echo "Generating with deepcopy-gen..."
deepcopy-gen \
	--go-header-file hack/boilerplate/boilerplate.generatego.txt \
	--input-dirs=github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1 \
	--output-package=github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1 \
	--output-file-base=zz_generated.deepcopy

echo "Generating with register-gen..."
register-gen \
	--go-header-file hack/boilerplate/boilerplate.generatego.txt \
	--input-dirs=github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1 \
	--output-package=github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1 \
	--output-file-base=zz_generated.register

echo "Generating with client-gen..."
client-gen \
	--go-header-file hack/boilerplate/boilerplate.generatego.txt \
	--input-base="" \
	--input=github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1 \
	--output-package=github.com/kube-arbiter/arbiter/pkg/generated/clientset \
	--clientset-name=versioned

echo "Generating with lister-gen..."
lister-gen \
	--go-header-file hack/boilerplate/boilerplate.generatego.txt \
	--input-dirs=github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1 \
	--output-package=github.com/kube-arbiter/arbiter/pkg/generated/listers

echo "Generating with informer-gen..."
informer-gen \
	--go-header-file hack/boilerplate/boilerplate.generatego.txt \
	--input-dirs=github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1 \
	--versioned-clientset-package=github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned \
	--listers-package=github.com/kube-arbiter/arbiter/pkg/generated/listers \
	--output-package=github.com/kube-arbiter/arbiter/pkg/generated/informers

echo "codegen done."
