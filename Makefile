# Include openshift build-machinery-go libraries
include ./vendor/github.com/openshift/build-machinery-go/make/golang.mk
include ./vendor/github.com/openshift/build-machinery-go/make/targets/openshift/deps.mk

TIMESTAMP :=$(shell date -u +'%Y-%m-%d-%H%M%S')
RELEASE_PRE :=0.4.7-0.microshift
SOURCE_GIT_TAG :=$(shell echo $(RELEASE_PRE)-$(TIMESTAMP))

SRC_ROOT :=$(shell pwd)

BUILD_CFG :=./images/build/Dockerfile
IMAGE_REPO :=quay.io/microshift/microshift
OUTPUT_DIR :=_output
CROSS_BUILD_BINDIR :=$(OUTPUT_DIR)/bin

CTR_CMD :=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))

GO_EXT_LD_FLAGS :=-extldflags '-static'
GO_LD_EXTRAFLAGS :=-X k8s.io/component-base/version.gitMajor=1 \
                   -X k8s.io/component-base/version.gitMinor=20 \
                   -X k8s.io/component-base/version.gitVersion=v1.20.1 \
                   -X k8s.io/component-base/version.gitCommit=5feb30e1bd3620 \
                   -X k8s.io/component-base/version.gitTreeState=clean \
                   -X k8s.io/component-base/version.buildDate=$(TIMESTAMP) \
                   -X k8s.io/client-go/pkg/version.gitMajor=1 \
                   -X k8s.io/client-go/pkg/version.gitMinor=20 \
                   -X k8s.io/client-go/pkg/version.gitVersion=v1.20.1 \
                   -X k8s.io/client-go/pkg/version.gitCommit=5feb30e1bd3620 \
                   -X k8s.io/client-go/pkg/version.gitTreeState=clean \
                   -X k8s.io/client-go/pkg/version.buildDate=$(TIMESTAMP) \
                   $(GO_EXT_LD_FLAGS) \
                   -s -w

debug:
	@echo FLAGS:"$(GO_LD_EXTRAFLAGS)"
	@echo TAG:"$(SOURCE_GIT_TAG)"

# These tags make sure we can statically link and avoid shared dependencies
GO_BUILD_FLAGS :=-tags 'include_gcs include_oss containers_image_openpgp gssapi providerless netgo osusergo'

# targets "all:" and "build:" defined in vendor/github.com/openshift/build-machinery-go/make/targets/golang/build.mk
microshift: build-containerized-cross-build-linux-amd64
.PHONY: microshift

update: update-generated-completions
.PHONY: update

###############################
# host build targets          #
###############################

_build_local:
	@mkdir -p "$(CROSS_BUILD_BINDIR)/$(GOOS)_$(GOARCH)"
	+@GOOS=$(GOOS) GOARCH=$(GOARCH) $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/microshift GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/$(GOOS)_$(GOARCH)

cross-build-linux-amd64:
	+$(MAKE) _build_local GOOS=linux GOARCH=amd64
.PHONY: cross-build-linux-amd64

cross-build-linux-arm64:
	+$(MAKE) _build_local GOOS=linux GOARCH=arm64
.PHONY: cross-build-linux-arm64

cross-build: cross-build-linux-amd64 cross-build-linux-arm64
.PHONY: cross-build

###############################
# containerized build targets #
###############################
_build_containerized:
	$(CTR_CMD) build -t $(IMAGE_REPO):$(SOURCE_GIT_TAG)-linux-$(ARCH) \
		-f "$(SRC_ROOT)"/images/build/Dockerfile \
		--build-arg SOURCE_GIT_TAG=$(SOURCE_GIT_TAG) \
		--build-arg ARCH=$(ARCH) \
		--build-arg MAKE_TARGET="cross-build-linux-$(ARCH)" \
		--platform="linux/$(ARCH)" \
		.
.PHONY: _build_containerized

build-containerized-cross-build-linux-amd64:
	+$(MAKE) _build_containerized ARCH=amd64
.PHONY: build-containerized-cross-build-linux-amd64

build-containerized-cross-build-linux-arm64:
	+$(MAKE) _build_containerized ARCH=arm64
.PHONY: build-containerized-cross-build-linux-arm64

build-containerized-cross-build:
	+$(MAKE) build-containerized-cross-build-linux-amd64
	+$(MAKE) build-containerized-cross-build-linux-arm64
.PHONY: build-containerized-cross-build

vendor:
	./hack/vendoring.sh
.PHONY: vendor

clean-cross-build:
	$(RM) -r '$(CROSS_BUILD_BINDIR)'
	$(RM) -rf $(OUTPUT_DIR)/staging
	if [ -d '$(OUTPUT_DIR)' ]; then rmdir --ignore-fail-on-non-empty '$(OUTPUT_DIR)'; fi
.PHONY: clean-cross-build

clean: clean-cross-build
.PHONY: clean

release:
	./scripts/release.sh --token $(TOKEN) --target $(TARGET) --version $(SOURCE_GIT_TAG)
.PHONY: release