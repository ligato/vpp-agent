SHELL := /usr/bin/env bash -o pipefail

PROJECT    := vpp-agent
VERSION	   ?= $(shell git describe --always --tags --dirty --match 'v*')
COMMIT     ?= $(shell git rev-parse HEAD)
BRANCH     ?= $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE ?= $(shell date +%s)
BUILD_HOST ?= $(shell hostname)
BUILD_USER ?= $(shell id -un)

GOPKG := go.ligato.io/vpp-agent/v3
LDFLAGS = -w -s \
	-X $(GOPKG)/pkg/version.app=$(PROJECT) \
	-X $(GOPKG)/pkg/version.version=$(VERSION) \
	-X $(GOPKG)/pkg/version.gitCommit=$(COMMIT) \
	-X $(GOPKG)/pkg/version.gitBranch=$(BRANCH) \
	-X $(GOPKG)/pkg/version.buildDate=$(BUILD_DATE) \
	-X $(GOPKG)/pkg/version.buildUser=$(BUILD_USER) \
	-X $(GOPKG)/pkg/version.buildHost=$(BUILD_HOST)

UNAME_OS   ?= $(shell uname -s)
UNAME_ARCH ?= $(shell uname -m)

ifndef CACHE_BASE
CACHE_BASE := $(HOME)/.cache/$(PROJECT)
endif
CACHE := $(CACHE_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
CACHE_BIN := $(CACHE)/bin
CACHE_INCLUDE := $(CACHE)/include
CACHE_VERSIONS := $(CACHE)/versions

export PATH := $(abspath $(CACHE_BIN)):$(PATH)

ifndef BUILD_DIR
BUILD_DIR := .build
endif

export GO111MODULE=on
export DOCKER_BUILDKIT=1

include vpp.env

ifeq ($(VPP_VERSION),)
VPP_VERSION=$(VPP_DEFAULT)
endif

VPP_IMG?=$(value VPP_$(VPP_VERSION)_IMAGE)
ifeq ($(UNAME_ARCH), aarch64)
VPP_IMG?=$(subst vpp-base,vpp-base-arm64,$(VPP_IMG))
endif
VPP_BINAPI?=$(value VPP_$(VPP_VERSION)_BINAPI)

SKIP_CHECK?=

ifeq ($(NOSTRIP),)
LDFLAGS += -w -s
endif

ifeq ($(NOTRIM),)
GO_BUILD_ARGS += -trimpath
endif

ifeq ($(BUILDPIE),y)
GO_BUILD_ARGS += -buildmode=pie
LDFLAGS += -extldflags=-Wl,-z,now,-z,relro
endif

ifeq ($(V),1)
GO_BUILD_ARGS += -v
endif

COVER_DIR ?= /tmp

help:
	@echo "List of make targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/^[^:]*://g' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT = help

-include scripts/make/buf.make

build: cmd examples

clean: clean-cmd clean-examples

agent: ## Build agent
	@echo "# installing agent ${VERSION}"
	@go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent

agentctl: ## Build agentctl
	@echo "# installing agentctl ${VERSION}"
	@go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/agentctl

install: ## Install commands
	@echo "# installing ${VERSION}"
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent-init
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/agentctl

cmd: ## Build commands
	@echo "# building ${VERSION}"
	cd cmd/vpp-agent && go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd cmd/vpp-agent-init && go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd cmd/agentctl && go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}

clean-cmd: ## Clean commands
	@echo "# cleaning command binaries"
	rm -f ./cmd/vpp-agent/vpp-agent
	rm -f ./cmd/vpp-agent/vpp-agent-init
	rm -f ./cmd/agentctl/agentctl

examples: ## Build examples
	@echo "# building examples"
	cd examples/customize/custom_api_model && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/customize/custom_vpp_plugin && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/govpp_call 		    	 && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/grpc_vpp/remote_client   && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/grpc_vpp/notifications   && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/kvscheduler/acl			 && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/kvscheduler/interconnect && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/kvscheduler/l2           && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/kvscheduler/acl          && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/kvscheduler/nat          && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/kvscheduler/rxplacement  && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/kvscheduler/vpp-l3       && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/kvscheduler/vrf          && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/localclient_linux/tap 	 && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/localclient_linux/veth 	 && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/localclient_vpp/nat      && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/localclient_vpp/plugins	 && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}

clean-examples: ## Clean examples
	@echo "# cleaning examples"
	cd examples/customize/custom_api_model && go clean
	cd examples/customize/custom_vpp_plugin && go clean
	cd examples/govpp_call 		    		&& go clean
	cd examples/grpc_vpp/remote_client 		&& go clean
	cd examples/grpc_vpp/notifications		&& go clean
	cd examples/kvscheduler/acl 			&& go clean
	cd examples/kvscheduler/interconnect 	&& go clean
	cd examples/kvscheduler/l2 				&& go clean
	cd examples/kvscheduler/acl 			&& go clean
	cd examples/kvscheduler/nat 			&& go clean
	cd examples/kvscheduler/vpp-l3 			&& go clean
	cd examples/localclient_linux/tap 	 	&& go clean
	cd examples/localclient_linux/veth 	 	&& go clean
	cd examples/localclient_vpp/nat      	&& go clean
	cd examples/localclient_vpp/plugins	 	&& go clean

purge: ## Purge cached files
	go clean -testcache -cache ./...

debug-remote: ## Debug remotely
	cd ./cmd/vpp-agent && dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient

# -------------------------------
#  Testing
# -------------------------------

test: ## Run unit tests
	@echo "# running unit tests"
	go test -tags="${GO_BUILD_TAGS}" ./...

test-cover: ## Run unit tests with coverage
	@echo "# running unit tests with coverage"
	go test -tags="${GO_BUILD_TAGS}" -covermode=count -coverprofile=${COVER_DIR}/coverage.out ./...
	@echo "# coverage data generated into ${COVER_DIR}/coverage.out"

test-cover-html: test-cover
	go tool cover -html=${COVER_DIR}/coverage.out -o ${COVER_DIR}/coverage.html
	@echo "# coverage report generated into ${COVER_DIR}/coverage.html"

perf: ## Run quick performance test
	@echo "# running perf test"
	./tests/perf/perf_test.sh grpc-perf 1000

perf-all: ## Run all performance tests
	@echo "# running all perf tests"
	./tests/perf/run_all.sh

integration-tests: test-tools ## Run integration tests
	@echo "# running integration tests"
	VPP_IMG=$(VPP_IMG) ./tests/integration/run_integration.sh

e2e-tests: images test-tools ## Run end-to-end tests
	@echo "# running end-to-end tests"
	VPP_AGENT=prod_vpp_agent ./tests/e2e/run_e2e.sh

# -------------------------------
#  Code generation
# -------------------------------

checknodiffgenerated:  ## Check no diff generated
	bash scripts/checknodiffgenerated.sh $(MAKE) generate

generate: generate-proto generate-binapi generate-desc-adapters ## Generate all

generate-proto: protocgengo ## Generate Protobuf files

get-binapi-generators:
	go install -mod=readonly go.fd.io/govpp/cmd/binapi-generator

generate-binapi: get-binapi-generators ## Generate Go code for VPP binary API
	@echo "# generating VPP binapi"
	VPP_BINAPI=$(VPP_BINAPI) ./scripts/genbinapi.sh

verify-binapi: ## Verify generated VPP binary API
	@echo "# verifying generated binapi"
	docker build -f docker/dev/Dockerfile \
		--build-arg VPP_IMG=${VPP_IMG} \
		--build-arg VPP_VERSION=${VPP_VERSION} \
		--target verify-binapi .

get-desc-adapter-generator:
	go install ./plugins/kvscheduler/descriptor-adapter

generate-desc-adapters: get-desc-adapter-generator ## Generate Go code for descriptors
	@echo "# generating descriptor adapters"
	go generate -x -run=descriptor-adapter ./...

get-bindata:
	go get -v github.com/jteeuwen/go-bindata/...
	go get -v github.com/elazarl/go-bindata-assetfs/...

bindata: get-bindata
	@echo "# generating bindata"
	go generate -x -run=go-bindata-assetfs ./...

proto-schema: ## Generate Protobuf schema image
	@echo "# generating proto schema"
	@$(MAKE) --no-print-directory buf-image

# -------------------------------
#  Dependencies
# -------------------------------

dep-install:
	@echo "# downloading project's dependencies"
	go mod download

dep-update:
	@echo "# updating all dependencies"
	@echo go mod tidy -v

dep-check:
	@echo "# checking dependencies"
	@if ! git --no-pager diff go.mod ; then \
		echo >&2 "go.mod has uncommitted changes!"; \
		exit 1; \
	fi
	go mod verify
	go mod tidy -v
	@if ! git --no-pager diff go.mod ; then \
		echo >&2 "go mod tidy check failed!"; \
		exit 1; \
	fi

# -------------------------------
#  Linters
# -------------------------------

gotestsumcmd := $(shell command -v gotestsum 2> /dev/null)

test-tools: ## install test tools
ifndef gotestsumcmd
	go install gotest.tools/gotestsum@v1.8.1
endif
	@env CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/test2json cmd/test2json

LINTER := $(shell command -v gometalinter 2> /dev/null)

get-linters:
ifndef LINTER
	@echo "# installing linters"
	go install github.com/alecthomas/gometalinter@latest
	gometalinter --install
endif

lint: get-linters ## Lint Go code
	@echo "# running code analysis"
	./scripts/static_analysis.sh golint vet

format: ## Format Go code
	@echo "# formatting the code"
	./scripts/gofmt.sh

MDLINKCHECK := $(shell command -v markdown-link-check 2> /dev/null)

get-linkcheck: ## Check links in Markdown files
ifndef MDLINKCHECK
	sudo apt-get update && sudo apt-get install -y npm
	npm install -g markdown-link-check@3.6.2
endif

check-links: get-linkcheck
	./scripts/check_links.sh

get-yamllint:
	pip install --user yamllint

yamllint: get-yamllint ## Lint YAML files
	@echo "# linting the yaml files"
	yamllint -c .yamllint.yml $(shell git ls-files '*.yaml' '*.yml' | grep -v 'vendor/')

lint-proto: ## Lint Protobuf files
	@echo "# linting Protobuf files"
	@$(MAKE) --no-print-directory buf-lint

check-proto: lint-proto ## Check proto files for breaking changes
	@echo "# checking proto files"
	@$(MAKE) --no-print-directory buf-breaking

# -------------------------------
#  Images
# -------------------------------

images: dev-image prod-image ## Build all images

dev-image: ## Build developer image
	@echo "# building dev image"
	IMAGE_TAG=$(IMAGE_TAG) \
		VPP_IMG=$(VPP_IMG) VPP_VERSION=$(VPP_VERSION) VPP_BINAPI=$(VPP_BINAPI) \
		VERSION=$(VERSION) COMMIT=$(COMMIT) BRANCH=$(BRANCH) \
		BUILD_DATE=$(BUILD_DATE) \
	  ./docker/dev/build.sh

prod-image: ## Build production image
	@echo "# building prod image"
	IMAGE_TAG=$(IMAGE_TAG) VPP_VERSION=$(VPP_VERSION) ./docker/prod/build.sh


.PHONY: help \
	agent agentctl build clean install purge \
	cmd examples clean-examples \
	test test-cover test-cover-html \
	generate checknodiffgenerated generate-binapi generate-proto get-binapi-generators \
	get-dep dep-install dep-update dep-check \
	get-linters lint format lint-proto check-proto \
	get-linkcheck check-links \
	get-yamllint yamllint \
	images dev-image prod-image \
	perf perf-all
