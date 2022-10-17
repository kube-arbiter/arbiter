GOOS ?= linux
GOARCH ?= $(shell go env GOARCH)
GOFLAGS ?=""
SOURCES := $(shell find . -type f  -name '*.go')

REGISTRY ?= "docker.io/kubearbiter"

DOCKER_TARGETS := observer \
                  executor \
                  scheduler

TARGETS := ${DOCKER_TARGETS} \
           abctl

WHAT ?= $(TARGETS)

# Build binary
#
# Args:
#   WHAT:   Target to build.
#   GOOS:   OS to build.
#   GOARCH: Arch to build.
#
# Example:
#   make all
#   make all WHAT=observer
#   make all WHAT=observer GOOS=linux GOARCH=amd64
#   make binary WHAT=cli GOARCH=amd64 GOOS=darwin

.PHONY: all
all: binary

binary:
	@GOFLAGS=$(GOFLAGS) BUILD_PLATFORMS=$(GOOS)/$(GOARCH) hack/build.sh $(WHAT)

.PHONY: clean
clean:
	rm -rf _output

.PHONY: update
update:
	@hack/update-all.sh

.PHONY: verify
verify:
	@hack/verify-all.sh

# Build image.
#
# Args:
#   WHAT:        Target to build.
#   GOOS:        OS to build.
#   GOARCH:      Arch to build.
#   OUTPUT_TYPE: Destination to save image(docker/registry).
#
# Example:
#   make images
#   make image WHAT=observer
#   make image WHAT=observer GOARCH=arm64
#   make image WHAT=observer GOARCH=amd64 OUTPUT_TYPE=registry
.PHONY: images
image:
	@REGISTRY=$(REGISTRY) OUTPUT_TYPE=$(OUTPUT_TYPE) BUILD_PLATFORMS=$(GOOS)/$(GOARCH) hack/build-image.sh $(filter ${DOCKER_TARGETS}, ${WHAT})
