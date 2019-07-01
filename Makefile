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

build: cmd examples

clean: clean-cmd clean-examples

agent:
	@echo "=> installing agent ${VERSION}"
	@go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent

install:
	@echo "=> installing ${VERSION}"
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent-init
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/agentctl

cmd:
	@echo "=> building ${VERSION}"
	cd cmd/vpp-agent && go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd cmd/vpp-agent-init && go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd cmd/agentctl && go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}

clean-cmd:
	@echo "=> cleaning command binaries"
	rm -f ./cmd/vpp-agent/vpp-agent
	rm -f ./cmd/vpp-agent/vpp-agent-init
	rm -f ./cmd/agentctl/agentctl

examples:
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

clean-examples:
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

# -------------------------------
#  Testing
# -------------------------------

test:
	@echo "=> running unit tests"
	go test -tags="${GO_BUILD_TAGS}" ./...

test-cover:
	@echo "=> running unit tests with coverage"
	go test -tags="${GO_BUILD_TAGS}" -covermode=count -coverprofile=${COVER_DIR}/coverage.out ./...
	@echo "=> coverage data generated into ${COVER_DIR}/coverage.out"

test-cover-html: test-cover
	go tool cover -html=${COVER_DIR}/coverage.out -o ${COVER_DIR}/coverage.html
	@echo "=> coverage report generated into ${COVER_DIR}/coverage.html"

test-cover-xml: test-cover
	gocov convert ${COVER_DIR}/coverage.out | gocov-xml > ${COVER_DIR}/coverage.xml
	@echo "=> coverage report generated into ${COVER_DIR}/coverage.xml"

perf:
	@echo "=> running perf test"
	./tests/perf/grpc-perf/test.sh 1000

perf-all:
	@echo "=> running all perf tests"
	./tests/perf/run_all.sh

# -------------------------------
#  Code generation
# -------------------------------

generate: generate-proto generate-binapi generate-desc-adapters

get-proto-generators:
	@go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogo

generate-proto: get-proto-generators
	@echo "=> generating proto"
	./scripts/genprotos.sh

get-binapi-generators:
	@go install ./vendor/git.fd.io/govpp.git/cmd/binapi-generator

generate-binapi: get-binapi-generators
	@echo "=> generating binapi"
	VPP_BINAPI=$(VPP_BINAPI) ./scripts/genbinapi.sh

verify-binapi:
	@echo "=> verifying binary api"
	docker build -f docker/dev/Dockerfile \
		--build-arg VPP_IMG=${VPP_IMG} \
		--build-arg VPP_BINAPI=${VPP_BINAPI} \
		--target verify-binapi .

get-desc-adapter-generator:
	@go install ./plugins/kvscheduler/descriptor-adapter

generate-desc-adapters: get-desc-adapter-generator
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
	cd plugins/restplugin && go generate

# -------------------------------
#  Dependencies
# -------------------------------

get-dep:
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	dep version

dep-install: get-dep
	@echo "=> installing project's dependencies"
	dep ensure -v

dep-update: get-dep
	@echo "=> updating all dependencies"
	dep ensure -update

dep-check: get-dep
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

lint: get-linters
	@echo "=> running code analysis"
	./scripts/static_analysis.sh golint vet

format:
	@echo "=> formatting the code"
	./scripts/gofmt.sh

MDLINKCHECK := $(shell command -v markdown-link-check 2> /dev/null)

get-linkcheck:
ifndef MDLINKCHECK
	sudo apt-get update && sudo apt-get install -y npm
	npm install -g markdown-link-check@3.6.2
endif

check-links: get-linkcheck
	./scripts/check_links.sh

get-yamllint:
	pip install --user yamllint

yamllint: get-yamllint
	@echo "=> linting the yaml files"
	yamllint -c .yamllint.yml $(shell git ls-files '*.yaml' '*.yml' | grep -v 'vendor/')

# -------------------------------
#  Images
# -------------------------------

images: dev-image prod-image

dev-image:
	@echo "=> building dev image"
	IMAGE_TAG=$(IMAGE_TAG) \
		VPP_IMG=$(VPP_IMG) VPP_BINAPI=$(VPP_BINAPI) \
		VERSION=$(VERSION) COMMIT=$(COMMIT) DATE=$(DATE) \
		./docker/dev/build.sh

prod-image:
	@echo "=> building prod image"
	IMAGE_TAG=$(IMAGE_TAG) \
    	./docker/prod/build.sh

# -------------------------------

travis:
	@echo "=> TRAVIS: $$TRAVIS_BUILD_STAGE_NAME"
	@echo "Build: #$$TRAVIS_BUILD_NUMBER ($$TRAVIS_BUILD_ID)"
	@echo "Job: #$$TRAVIS_JOB_NUMBER ($$TRAVIS_JOB_ID)"
	@echo "AllowFailure: $$TRAVIS_ALLOW_FAILURE TestResult: $$TRAVIS_TEST_RESULT"
	@echo "Type: $$TRAVIS_EVENT_TYPE PullRequest: $$TRAVIS_PULL_REQUEST"
	@echo "Repo: $$TRAVIS_REPO_SLUG Branch: $$TRAVIS_BRANCH"
	@echo "Commit: $$TRAVIS_COMMIT"
	@echo "$$TRAVIS_COMMIT_MESSAGE"
	@echo "Range: $$TRAVIS_COMMIT_RANGE"
	@echo "Files:"
	@echo "$$(git diff --name-only $$TRAVIS_COMMIT_RANGE)"


.PHONY: build clean \
	install cmd examples clean-examples test \
	test-cover test-cover-html test-cover-xml \
	generate genereate-binapi generate-proto get-binapi-generators get-proto-generators \
	get-dep dep-install dep-update dep-check \
	get-linters lint format \
	get-linkcheck check-links \
	get-yamllint yamllint \
	images dev-image prod-image \
	perf perf-all \
	travis
