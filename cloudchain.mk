# Copyright IBM Corp All Rights Reserved.
# Copyright London Stock Exchange Group All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

BASE_VERSION = 1.3.1
BASEIMAGE_RELEASE=0.4.13
BASE_DOCKER_LABEL = org.hyperledger.fabric

BUILD_DIR ?= .build
DOCKER_NS ?= cloudchain
BASE_DOCKER_NS = buildpack-deps
BASE_DOCKER_TAG = xenial
ARCH=$(shell go env GOARCH)
EXTRA_VERSION ?= $(shell git rev-parse --short HEAD)
PROJECT_VERSION=$(BASE_VERSION)-snapshot-$(EXTRA_VERSION)
DOCKER_TAG=$(ARCH)-$(PROJECT_VERSION)

PROJECT_NAME = hyperledger/fabric
PKGNAME = github.com/$(PROJECT_NAME)
EXPERIMENTAL ?= false

# defined in common/metadata/metadata.go
METADATA_VAR = Version=$(BASE_VERSION)
METADATA_VAR += CommitSHA=$(EXTRA_VERSION)
METADATA_VAR += BaseVersion=$(BASEIMAGE_RELEASE)
METADATA_VAR += BaseDockerLabel=$(BASE_DOCKER_LABEL)
METADATA_VAR += DockerNamespace=$(DOCKER_NS)
METADATA_VAR += BaseDockerNamespace=$(BASE_DOCKER_NS)
METADATA_VAR += Experimental=$(EXPERIMENTAL)

GO_LDFLAGS = $(patsubst %,-X $(PKGNAME)/common/metadata.%,$(METADATA_VAR))
GO_TAGS ?= pluginsenabled
CGO_FLAGS = CGO_CFLAGS=" "

pkgmap.peer           := $(PKGNAME)/peer
pkgmap.gscc.so        := $(PKGNAME)/plugins/gscc

# No sense rebuilding when non production code is changed
PROJECT_FILES = $(shell git ls-files  | grep -v ^test | grep -v ^unit-test | \
	grep -v ^.git | grep -v ^examples | grep -v ^devenv | grep -v .png$ | \
	grep -v ^LICENSE | grep -v ^vendor )

peer-docker: $(BUILD_DIR)/image/peer

$(BUILD_DIR)/docker/bin/peer: $(PROJECT_FILES)
	@echo "Building $@"
	@mkdir -p $(BUILD_DIR)/docker/bin
	$(CGO_FLAGS) GOBIN=$(abspath $(@D)) go install -tags "$(GO_TAGS)" -ldflags "$(GO_LDFLAGS)" $(pkgmap.$(@F))
	@touch $@

$(BUILD_DIR)/docker/lib/gscc.so: $(PROJECT_FILES)
	@echo "Building $@"
	@mkdir -p $(BUILD_DIR)/docker/lib
	go build -buildmode=plugin -o $(abspath $(@D))/gscc.so $(pkgmap.$(@F))

$(BUILD_DIR)/image/peer/payload:       $(BUILD_DIR)/docker/bin/peer \
				$(BUILD_DIR)/docker/lib/gscc.so \
				$(BUILD_DIR)/sampleconfig.tar.bz2 \
				$(BUILD_DIR)/scripts.tar.bz2

$(BUILD_DIR)/image/%/payload:
	mkdir -p $@
	cp $^ $@

$(BUILD_DIR)/image/peer/Dockerfile: plugins/gscc/images/peer/Dockerfile.in
	mkdir -p $(@D)
	@cat $< \
		| sed -e 's|_BASE_NS_|$(BASE_DOCKER_NS)|g' \
		| sed -e 's|_BASE_TAG_|$(BASE_DOCKER_TAG)|g' \
		> $@

$(BUILD_DIR)/image/peer: cloudchain.mk $(BUILD_DIR)/image/peer/payload $(BUILD_DIR)/image/peer/Dockerfile
	@echo "Building docker peer-image"
	docker build -t $(DOCKER_NS)/fabric-peer $@
	docker tag $(DOCKER_NS)/fabric-peer $(DOCKER_NS)/fabric-peer:$(DOCKER_TAG)
	docker tag $(DOCKER_NS)/fabric-peer $(DOCKER_NS)/fabric-peer:$(ARCH)-latest
	@touch $@

$(BUILD_DIR)/sampleconfig.tar.bz2: $(shell find plugins/gscc/sampleconfig -type f)
	(cd plugins/gscc/sampleconfig && tar -jc *) > $@

$(BUILD_DIR)/scripts.tar.bz2: $(shell find plugins/gscc/scripts -type f)
	(cd plugins/gscc/scripts && tar -jc *) > $@