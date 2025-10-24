-include dev.env

## Set all the environment variables here
# Docker Registry
DOCKER_REGISTRY ?= docker.io/rocm

# Build Container environment
DOCKER_BUILDER_TAG ?= v1.0
BUILD_BASE_IMAGE ?= ubuntu:22.04
BUILD_CONTAINER ?= $(DOCKER_REGISTRY)/container-toolkit-build:$(DOCKER_BUILDER_TAG)
RPM_BUILD_CONTAINER ?= $(DOCKER_REGISTRY)/container-toolkit-rpm-build:$(DOCKER_BUILDER_TAG)
BIN_DIRECTORY_SUFFIX ?= deb


# export environment variables used across project
export DOCKER_REGISTRY
export BUILD_CONTAINER
export RPM_BUILD_CONTAINER
export BUILD_BASE_IMAGE

CUR_USER:=$(shell whoami)
CUR_TIME:=$(shell date +%Y-%m-%d_%H.%M.%S)
CONTAINER_NAME:=${CUR_USER}_container_toolkit-bld
RPM_CONTAINER_NAME:=${CUR_USER}_container_toolkit-rpm-bld
CONTAINER_WORKDIR ?= /usr/src/github.com/ROCm/container-toolkit

TOP_DIR := $(PWD)
GOINSECURE='github.com, google.golang.org, golang.org'
GOFLAGS ='-buildvcs=false'
BUILD_DATE ?= $(shell date   +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT ?= $(shell git rev-list -1 HEAD --abbrev-commit)
VERSION ?= $(shell git describe --tags --always --dirty)

export ${GOROOT}
export ${GOPATH}
export ${TOP_DIR}
export ${GOFLAGS}
export ${GOINSECURE}
export ${BUILD_VER_ENV}

# 22.04 - jammy
# 24.04 - noble
UBUNTU_VERSION ?= jammy
UBUNTU_VERSION_NUMBER = 22.04
ifeq (${UBUNTU_VERSION}, noble)
UBUNTU_VERSION_NUMBER = 24.04
endif

DEBIAN_VERSION := "1.2.0"

DEBIAN_CONTROL = ${TOP_DIR}/build/debian/DEBIAN/control
DEBIAN_PRERM = ${TOP_DIR}/build/debian/DEBIAN/prerm
BUILD_VER_ENV = ${DEBIAN_VERSION}~$(UBUNTU_VERSION_NUMBER)
PKG_PATH := ${TOP_DIR}/build/debian/usr/local/bin

##################
# Makefile targets
#
##@ QuickStart
.PHONY: default
default: build-dev-container-deb ## Quick start to build everything from docker shell container

# create development build container only if there is changes done on
# tools/deb/base-image/Dockerfile
.PHONY: build-dev-container-deb
build-dev-container-deb:
	${MAKE} -C tools/deb/base-image all INSECURE_REGISTRY=$(INSECURE_REGISTRY)
	$(MAKE) docker-compile-deb

.PHONY: docker-compile-deb
docker-compile-deb:
	docker run --rm -it --privileged \
		--name ${CONTAINER_NAME} \
		-e "USER_NAME=$(shell whoami)" \
		-e "USER_UID=$(shell id -u)" \
		-e "USER_GID=$(shell id -g)" \
		-e "GIT_COMMIT=${GIT_COMMIT}" \
		-e "GIT_VERSION=${GIT_VERSION}" \
		-e "BUILD_DATE=${BUILD_DATE}" \
		-v $(CURDIR):$(CONTAINER_WORKDIR) \
		-v $(HOME)/.ssh:/home/$(shell whoami)/.ssh \
		-w $(CONTAINER_WORKDIR) \
		$(BUILD_CONTAINER) \
		bash -c "cd $(CONTAINER_WORKDIR) && source ~/.bashrc && git config --global --add safe.directory $(CONTAINER_WORKDIR) && make all"

# create rpmbuild development build container only if there is changes done on
# tools/rpmbuild/base-image/Dockerfile
.PHONY: build-dev-container-rpm
build-dev-container-rpm:
	${MAKE} -C tools/rpmbuild/base-image all INSECURE_REGISTRY=$(INSECURE_REGISTRY)
	$(MAKE) docker-compile-rpm

.PHONY: docker-compile-rpm
docker-compile-rpm:
	docker run --rm -it --privileged \
		--name ${RPM_CONTAINER_NAME} \
		-e "USER_NAME=$(shell whoami)" \
		-e "USER_UID=$(shell id -u)" \
		-e "USER_GID=$(shell id -g)" \
		-e "GIT_COMMIT=${GIT_COMMIT}" \
		-e "GIT_VERSION=${GIT_VERSION}" \
		-e "BUILD_DATE=${BUILD_DATE}" \
		-v $(CURDIR):$(CONTAINER_WORKDIR) \
		-v $(HOME)/.ssh:/home/$(shell whoami)/.ssh \
		-w $(CONTAINER_WORKDIR) \
		$(RPM_BUILD_CONTAINER) \
		bash -c "cd $(CONTAINER_WORKDIR) && source ~/.bashrc && git config --global --add safe.directory $(CONTAINER_WORKDIR) && make BIN_DIRECTORY_SUFFIX=rpmbuild all "

.PHONY: all
all:
	${MAKE} gen checks container-toolkit container-toolkit-ctk

.PHONY: pkg-deb pkg-deb-clean
pkg-deb-clean:
	rm -rf ${TOP_DIR}/deb/bin/*.deb

pkg-deb: pkg-deb-clean
	docker run --rm -it --privileged \
		--name ${CONTAINER_NAME} \
		-e "USER_NAME=$(shell whoami)" \
		-e "USER_UID=$(shell id -u)" \
		-e "USER_GID=$(shell id -g)" \
		-e "GIT_COMMIT=${GIT_COMMIT}" \
		-e "GIT_VERSION=${GIT_VERSION}" \
		-e "BUILD_DATE=${BUILD_DATE}" \
		-v $(CURDIR):$(CONTAINER_WORKDIR) \
		-v $(HOME)/.ssh:/home/$(shell whoami)/.ssh \
		-w $(CONTAINER_WORKDIR) \
		$(BUILD_CONTAINER) \
		bash -c "cd $(CONTAINER_WORKDIR) && source ~/.bashrc && git config --global --add safe.directory $(CONTAINER_WORKDIR) && make deb-pkg-build UBUNTU_VERSION=$(UBUNTU_VERSION)"

.PHONY: deb-pkg-build
deb-pkg-build: all
	@echo "Building debian for $(BUILD_VER_ENV)"

	# copy and strip files
	mkdir -p ${PKG_PATH}
	cp -vf $(CURDIR)/bin/deb/amd-container-runtime ${PKG_PATH}/
	cp -vf $(CURDIR)/bin/deb/amd-ctk ${PKG_PATH}/
	cp -vf $(CURDIR)/build/cleanup.sh $(DEBIAN_PRERM)
	chmod 0755 $(DEBIAN_PRERM)

	cd ${TOP_DIR}
	sed -i "s/BUILD_VER_ENV/$(BUILD_VER_ENV)/g" $(DEBIAN_CONTROL)
	dpkg-deb -Zxz --build build/debian ${TOP_DIR}/bin

	# revert the dynamic version set file
	git checkout $(DEBIAN_CONTROL)
	rm -rf $(DEBIAN_PRERM)

	# rename for internal build
	mv -vf ${TOP_DIR}/bin/amd-container-toolkit*~${UBUNTU_VERSION_NUMBER}_amd64.deb ${TOP_DIR}/bin/amd-container-toolkit_${UBUNTU_VERSION_NUMBER}_amd64.deb

.PHONY: rpm-pkg-build
rpm-pkg-build: all
	CONTAINER_WORKDIR=$(CONTAINER_WORKDIR) rpmbuild -bb $(CURDIR)/build/rpmbuild.spec
	cp $(HOME)/rpmbuild/RPMS/x86_64/*.rpm $(CURDIR)/bin/

.PHONY: pkg-rpm pkg-rpm-clean
pkg-rpm-clean:
	rm -rf ${CURDIR}/bin/*.rpm

pkg-rpm: pkg-rpm-clean
	docker run --rm -it --privileged \
		--name ${RPM_CONTAINER_NAME} \
		-e "USER_NAME=$(shell whoami)" \
		-e "USER_UID=$(shell id -u)" \
		-e "USER_GID=$(shell id -g)" \
		-e "GIT_COMMIT=${GIT_COMMIT}" \
		-e "GIT_VERSION=${GIT_VERSION}" \
		-e "BUILD_DATE=${BUILD_DATE}" \
		-v $(CURDIR):$(CONTAINER_WORKDIR) \
		-v $(HOME)/.ssh:/home/$(shell whoami)/.ssh \
		-w $(CONTAINER_WORKDIR) \
		$(RPM_BUILD_CONTAINER) \
		bash -c "cd $(CONTAINER_WORKDIR) && source ~/.bashrc && git config --global --add safe.directory $(CONTAINER_WORKDIR) && make BIN_DIRECTORY_SUFFIX=rpmbuild rpm-pkg-build"



.PHONY: gen
gen: gopkglist

.PHONY: gopkglist
gopkglist:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.1
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY:checks
checks: vet fmt

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: ## Run go test against code.
	AMD_CTK_PATH=$(CURDIR)/bin/$(BIN_DIRECTORY_SUFFIX)/amd-ctk go test -v ./...

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
.PHONY: golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	$(call go-get-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.1)


GOFILES_NO_VENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
.PHONY: lint
lint: golangci-lint ## Run golangci-lint against code.
	@if [ `gofmt -l $(GOFILES_NO_VENDOR) | wc -l` -ne 0 ]; then \
		echo There are some malformed files, please make sure to run \'make fmt\'; \
		gofmt -l $(GOFILES_NO_VENDOR); \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run -v --timeout 5m0s

# go-get-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
}
endef

container-toolkit:
	@echo "building amd container toolkit"
	CGO_ENABLED=0 go build  -C cmd/container-runtime -ldflags "-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE} -X main.Publish=${DISABLE_DEBUG}" -o $(CURDIR)/bin/$(BIN_DIRECTORY_SUFFIX)/amd-container-runtime

container-toolkit-ctk:
	@echo "building amd container toolkit ctk"
	CGO_ENABLED=0 go build  -C cmd/amd-ctk -ldflags "-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE} -X main.Publish=${DISABLE_DEBUG}" -o $(CURDIR)/bin/$(BIN_DIRECTORY_SUFFIX)/amd-ctk
