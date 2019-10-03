VERSION ?= $(shell git describe --always --tags --dirty)
COMMIT  ?= $(shell git rev-parse HEAD)
DATE    ?= $(shell git log -1 --format="%ct" | xargs -I{} date -d @{} +'%Y-%m-%dT%H:%M%:z')
ARCH    ?= $(shell uname -m)

CNINFRA := github.com/ligato/vpp-agent/vendor/github.com/ligato/cn-infra/agent
LDFLAGS = -X $(CNINFRA).BuildVersion=$(VERSION) -X $(CNINFRA).CommitHash=$(COMMIT) -X $(CNINFRA).BuildDate=$(DATE)

include vpp.env

ifeq ($(VPP_VERSION),)
VPP_VERSION = $(VPP_DEFAULT)
endif
VPP_IMG:=$(value VPP_IMG_$(VPP_VERSION))
ifeq (${ARCH}, aarch64)
VPP_IMG:=$(subst vpp-base,vpp-base-arm64,$(VPP_IMG))
endif
VPP_BINAPI?=$(value VPP_BINAPI_$(VPP_VERSION))
SKIP_CHECK?=

ifeq ($(NOSTRIP),)
LDFLAGS += -w -s
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

build: cmd examples

clean: clean-cmd clean-examples

agent: ## Build agent
	@echo "=> installing agent ${VERSION}"
	@go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent

agentctl: ## Build agentctl
	@echo "=> installing agentctl ${VERSION}"
	@go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/agentctl

install: ## Install commands
	@echo "=> installing ${VERSION}"
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent-init
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/agentctl

cmd: ## Build commands
	@echo "=> building ${VERSION}"
	cd cmd/vpp-agent && go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd cmd/vpp-agent-init && go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd cmd/agentctl && go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}

clean-cmd: ## Clean commands
	@echo "=> cleaning command binaries"
	rm -f ./cmd/vpp-agent/vpp-agent
	rm -f ./cmd/vpp-agent/vpp-agent-init
	rm -f ./cmd/agentctl/agentctl

examples: ## Build examples
	@echo "=> building examples"
	cd examples/custom_model	    	 && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
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
	@echo "=> cleaning examples"
	cd examples/custom_model	    		&& go clean
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

debug-remote: ## Debug remotely
	cd ./cmd/vpp-agent && dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient

# -------------------------------
#  Testing
# -------------------------------

test: ## Run unit tests
	@echo "=> running unit tests"
	go test -tags="${GO_BUILD_TAGS}" ./...

test-cover: ## Run unit tests with coverage
	@echo "=> running unit tests with coverage"
	go test -tags="${GO_BUILD_TAGS}" -covermode=count -coverprofile=${COVER_DIR}/coverage.out ./...
	@echo "=> coverage data generated into ${COVER_DIR}/coverage.out"

test-cover-html: test-cover
	go tool cover -html=${COVER_DIR}/coverage.out -o ${COVER_DIR}/coverage.html
	@echo "=> coverage report generated into ${COVER_DIR}/coverage.html"

test-cover-xml: test-cover
	gocov convert ${COVER_DIR}/coverage.out | gocov-xml > ${COVER_DIR}/coverage.xml
	@echo "=> coverage report generated into ${COVER_DIR}/coverage.xml"

perf: ## Run quick performance test
	@echo "=> running perf test"
	./tests/perf/perf_test.sh grpc-perf 1000

perf-all: ## Run all performance tests
	@echo "=> running all perf tests"
	./tests/perf/run_all.sh

integration-tests: ## Run integration tests
	@echo "=> running integration tests"
	VPP_IMG=$(VPP_IMG) ./tests/integration/vpp_integration.sh

e2e-tests: ## Run end-to-end tests
	@echo "=> running end-to-end tests"
	VPP_IMG=$(VPP_IMG) ./tests/e2e/run_e2e.sh

e2e-tests-cover: ## Run end-to-end tests with coverage
	@echo "=> running end-to-end tests with coverage"
	VPP_IMG=$(VPP_IMG) COVER_DIR=$(COVER_DIR) ./tests/e2e/run_e2e.sh
	@echo "=> coverage report generated into ${COVER_DIR}/e2e-cov.out"

# -------------------------------
#  Code generation
# -------------------------------

generate: generate-proto generate-binapi generate-desc-adapters ## Generate all

get-proto-generators:
	@go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogo

generate-proto: get-proto-generators ## Generate Go code for Protobuf files
	@echo "=> generating proto"
	./scripts/genprotos.sh

get-binapi-generators:
	@go install ./vendor/git.fd.io/govpp.git/cmd/binapi-generator

generate-binapi: get-binapi-generators ## Generate Go code for VPP binary API
	@echo "=> generating binapi"
	VPP_BINAPI=$(VPP_BINAPI) ./scripts/genbinapi.sh

verify-binapi: ## Verify generated VPP binary API
	@echo "=> verifying binary api"
	docker build -f docker/dev/Dockerfile \
		--build-arg VPP_IMG=${VPP_IMG} \
		--build-arg VPP_BINAPI=${VPP_BINAPI} \
		--target verify-binapi .

get-desc-adapter-generator:
	@go install ./plugins/kvscheduler/descriptor-adapter

generate-desc-adapters: get-desc-adapter-generator ## Generate Go code for descriptors
	@echo "=> generating descriptor adapters"
	cd plugins/linux/ifplugin && go generate
	cd plugins/linux/l3plugin && go generate
	cd plugins/linux/iptablesplugin && go generate
	cd plugins/vpp/aclplugin && go generate
	cd plugins/vpp/ifplugin && go generate
	cd plugins/vpp/ipsecplugin && go generate
	cd plugins/vpp/l2plugin && go generate
	cd plugins/vpp/l3plugin && go generate
	cd plugins/vpp/natplugin && go generate
	cd plugins/vpp/puntplugin && go generate
	cd plugins/vpp/stnplugin && go generate
	cd plugins/vpp/puntplugin && go generate
	cd plugins/vpp/srplugin && go generate
	@echo

get-bindata:
	go get -v github.com/jteeuwen/go-bindata/...
	go get -v github.com/elazarl/go-bindata-assetfs/...

bindata: get-bindata
	go generate -x ./plugins/restplugin

# -------------------------------
#  Dependencies
# -------------------------------

get-dep:
	curl -sSfL https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	dep version

dep-install: get-dep
	@echo "=> installing project's dependencies"
	dep ensure -v

dep-update: get-dep
	@echo "=> updating all dependencies"
	dep ensure -update

dep-check: get-dep ## Check Go dependencies
	@echo "=> checking dependencies"
	dep check

# -------------------------------
#  Linters
# -------------------------------

LINTER := $(shell command -v gometalinter 2> /dev/null)

get-linters:
ifndef LINTER
	@echo "=> installing linters"
	go get -v github.com/alecthomas/gometalinter
	gometalinter --install
endif

lint: get-linters ## Lint Go code
	@echo "=> running code analysis"
	./scripts/static_analysis.sh golint vet

format: ## Format Go code
	@echo "=> formatting the code"
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
	@echo "=> linting the yaml files"
	yamllint -c .yamllint.yml $(shell git ls-files '*.yaml' '*.yml' | grep -v 'vendor/')

# -------------------------------
#  Images
# -------------------------------

images: dev-image prod-image ## Build all images

dev-image: ## Build developer image
	@echo "=> building dev image"
	IMAGE_TAG=$(IMAGE_TAG) \
		VPP_IMG=$(VPP_IMG) VPP_BINAPI=$(VPP_BINAPI) \
		VERSION=$(VERSION) COMMIT=$(COMMIT) DATE=$(DATE) \
		./docker/dev/build.sh

prod-image: ## Build production image
	@echo "=> building prod image"
	IMAGE_TAG=$(IMAGE_TAG) \
    	./docker/prod/build.sh


.PHONY: help \
	agent agentctl build clean install \
	cmd examples clean-examples \
	test test-cover test-cover-html test-cover-xml \
	generate genereate-binapi generate-proto get-binapi-generators get-proto-generators \
	get-dep dep-install dep-update dep-check \
	get-linters lint format \
	get-linkcheck check-links \
	get-yamllint yamllint \
	images dev-image prod-image \
	perf perf-all
