include vpp.env

VERSION ?= $(shell git describe --always --tags --dirty)
COMMIT  ?= $(shell git rev-parse HEAD)
DATE    ?= $(shell git log -1 --format="%ct" | xargs -I{} date -d @{} +'%Y-%m-%dT%H:%M%:z')

CNINFRA := github.com/ligato/vpp-agent/vendor/github.com/ligato/cn-infra/agent
LDFLAGS = -X $(CNINFRA).BuildVersion=$(VERSION) -X $(CNINFRA).CommitHash=$(COMMIT) -X $(CNINFRA).BuildDate=$(DATE)

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

# Build all
build: cmd examples

# Clean all
clean: clean-cmd clean-examples

# Install commands
agent:
	@echo "=> installing agent ${VERSION}"
	@go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent

# Install commands
install:
	@echo "=> installing commands ${VERSION}"
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/vpp-agent-ctl
	go install -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS} ./cmd/agentctl

# Build commands
cmd:
	@echo "=> building commands ${VERSION}"
	cd cmd/vpp-agent 		&& go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd cmd/vpp-agent-ctl	&& go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd cmd/agentctl 		&& go build -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}

# Clean commands
clean-cmd:
	@echo "=> cleaning command binaries"
	rm -f ./cmd/vpp-agent/vpp-agent
	rm -f ./cmd/vpp-agent-ctl/vpp-agent-ctl
	rm -f ./cmd/agentctl/agentctl

# Build examples
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
	cd examples/kvscheduler/vpp-l3       && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/localclient_linux/tap 	 && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/localclient_linux/veth 	 && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/localclient_vpp/nat      && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}
	cd examples/localclient_vpp/plugins	 && go build -tags="${GO_BUILD_TAGS}" ${GO_BUILD_ARGS}

# Clean examples
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

# Run tests
test:
	@echo "=> running unit tests"
	go test -tags="${GO_BUILD_TAGS}" ./...

# Run coverage report
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

# Code generation
generate: generate-proto generate-binapi generate-desc-adapters

# Get generator tools
get-proto-generators:
	@go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogo

# Generate proto models
generate-proto: get-proto-generators
	@echo "=> generating proto"
	./scripts/genprotos.sh
	@echo "=> generating docs"
	@cd api/models/docgen && go build && ./docgen ../../../docs

# Get generator tools
get-binapi-generators:
	@go install ./vendor/git.fd.io/govpp.git/cmd/binapi-generator

# Generate binary api
generate-binapi: get-binapi-generators
	@echo "=> generating binapi"
	cd plugins/vpp/binapi && go generate
	@echo "=> applying fix patches"
	find plugins/vpp/binapi -maxdepth 1 -type f -name '*.patch' -exec patch --no-backup-if-mismatch -p1 -i {} \;
	@echo

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

verify-binapi:
	@echo "=> verifying binary api"
	docker build -f docker/dev/Dockerfile \
		--build-arg VPP_REPO_URL=${VPP_REPO_URL} \
		--build-arg VPP_COMMIT=${VPP_COMMIT} \
		--target verify-stage .

get-bindata:
	go get -v github.com/jteeuwen/go-bindata/...
	go get -v github.com/elazarl/go-bindata-assetfs/...

bindata: get-bindata
	cd plugins/restplugin && go generate

# Get dependency manager tool
get-dep:
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	dep version

# Install the project's dependencies
dep-install: get-dep
	@echo "=> installing project's dependencies"
	dep ensure -v

# Update the locked versions of all dependencies
dep-update: get-dep
	@echo "=> updating all dependencies"
	dep ensure -update

# Check state of dependencies
dep-check: get-dep
	@echo "=> checking dependencies"
	dep check

LINTER := $(shell command -v gometalinter 2> /dev/null)

# Get linter tools
get-linters:
ifndef LINTER
	@echo "=> installing linters"
	go get -v github.com/alecthomas/gometalinter
	gometalinter --install
endif

# Run linters
lint: get-linters
	@echo "=> running code analysis"
	./scripts/static_analysis.sh golint vet

# Format code
format:
	@echo "=> formatting the code"
	./scripts/gofmt.sh

MDLINKCHECK := $(shell command -v markdown-link-check 2> /dev/null)

# Get link check tool
get-linkcheck:
ifndef MDLINKCHECK
	sudo apt-get update && sudo apt-get install -y npm
	npm install -g markdown-link-check@3.6.2
endif

# Validate links in markdown files
check-links: get-linkcheck
	./scripts/check_links.sh

# Travis
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

# Install yamllint
get-yamllint:
	pip install --user yamllint

# Lint the yaml files
yamllint: get-yamllint
	@echo "=> linting the yaml files"
	yamllint -c .yamllint.yml $(shell git ls-files '*.yaml' '*.yml' | grep -v 'vendor/')

images: dev-image prod-image

dev-image:
	./docker/dev/build.sh

prod-image:
	./docker/prod/build.sh

perf:
	@echo "=> running perf test"
	./tests/perf/grpc-perf/test.sh 1000

perf-all:
	@echo "=> running all perf tests"
	./tests/perf/run_all.sh

.PHONY: build clean \
	install cmd examples clean-examples test \
	test-cover test-cover-html test-cover-xml \
	generate genereate-binapi generate-proto get-binapi-generators get-proto-generators \
	get-dep dep-install dep-update dep-check \
	get-linters lint format \
	get-linkcheck check-links \
	travis \
	get-yamllint yamllint \
	images dev-image prod-image \
	perf perf-all
