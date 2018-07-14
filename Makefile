include vpp.env

VERSION	?= $(shell git describe --always --tags --dirty)
COMMIT	?= $(shell git rev-parse HEAD)
DATE	:= $(shell date +'%Y-%m-%dT%H:%M%:z')

CNINFRA_CORE := github.com/ligato/vpp-agent/vendor/github.com/ligato/cn-infra/core
LDFLAGS	= -X $(CNINFRA_CORE).BuildVersion=$(VERSION) -X $(CNINFRA_CORE).CommitHash=$(COMMIT) -X $(CNINFRA_CORE).BuildDate=$(DATE)

ifeq ($(NOSTRIP),)
LDFLAGS += -w -s
endif

COVER_DIR ?= /tmp

# Build all
build: cmd examples

# Clean all
clean: clean-cmd clean-examples

# Install commands
install:
	@echo "=> installing commands ${VERSION}"
	go install -v -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ./cmd/vpp-agent
	go install -v -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ./cmd/vpp-agent-ctl
	go install -v -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ./cmd/agentctl

# Build commands
cmd:
	@echo "=> building commands ${VERSION}"
	cd cmd/vpp-agent 		&& go build -v -i -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}"
	cd cmd/vpp-agent-ctl	&& go build -v -i -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}"
	cd cmd/agentctl 		&& go build -v -i -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}"

# Clean commands
clean-cmd:
	@echo "=> cleaning binaries"
	rm -f ./cmd/vpp-agent/vpp-agent
	rm -f ./cmd/vpp-agent-ctl/vpp-agent-ctl
	rm -f ./cmd/agentctl/agentctl

# Build examples
examples:
	@echo "=> building examples"
	cd examples/govpp_call 		    	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_bd_cache 	    	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_iface_cache     	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_mapping_lookup  	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_mapping_watcher     && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_linux/veth 	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_linux/tap 	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_vpp/plugins && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_vpp/nat     && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/grpc_vpp/remote_client  && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/grpc_vpp/notifications  && go build -v -i -tags="${GO_BUILD_TAGS}"

# Clean examples
clean-examples:
	@echo "=> cleaning examples"
	rm -f examples/govpp_call/govpp_call
	rm -f examples/idx_bd_cache/idx_bd_cache
	rm -f examples/idx_iface_cache/idx_iface_cache
	rm -f examples/idx_mapping_lookup/idx_mapping_lookup
	rm -f examples/idx_mapping_watcher/idx_mapping_watcher
	rm -f examples/localclient_linux/veth/veth
	rm -f examples/localclient_linux/tap/tap
	rm -r examples/localclient_vpp/localclient_vpp

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
generate: generate-proto generate-binapi

# Get generator tools
get-proto-generators:
	go install -v ./vendor/github.com/gogo/protobuf/protoc-gen-gogo

# Generate proto models
generate-proto: get-proto-generators
	@echo "=> generating proto"
	cd plugins/linux/ifplugin && go generate
	cd plugins/linux/l3plugin && go generate
	cd plugins/vpp/aclplugin && go generate
	cd plugins/vpp/ifplugin && go generate
	cd plugins/vpp/ipsecplugin && go generate
	cd plugins/vpp/l2plugin && go generate
	cd plugins/vpp/l3plugin && go generate
	cd plugins/vpp/l4plugin && go generate
	cd plugins/vpp/rpc && go generate
	cd plugins/vpp/srplugin && go generate

# Get generator tools
get-binapi-generators:
	go install -v ./vendor/git.fd.io/govpp.git/cmd/binapi-generator
	go install -v ./vendor/github.com/ungerik/pkgreflect

# Generate binary api
generate-binapi: get-binapi-generators
	@echo "=> generating binapi"
	cd plugins/vpp/binapi && go generate
	cd plugins/vpp/binapi/acl && pkgreflect
	cd plugins/vpp/binapi/af_packet && pkgreflect
	cd plugins/vpp/binapi/bfd && pkgreflect
	cd plugins/vpp/binapi/dhcp && pkgreflect
	cd plugins/vpp/binapi/interfaces && pkgreflect
	cd plugins/vpp/binapi/ip && pkgreflect
	cd plugins/vpp/binapi/ipsec && pkgreflect
	cd plugins/vpp/binapi/l2 && pkgreflect
	cd plugins/vpp/binapi/memif && pkgreflect
	cd plugins/vpp/binapi/nat && pkgreflect
	cd plugins/vpp/binapi/session && pkgreflect
	cd plugins/vpp/binapi/sr && pkgreflect
	cd plugins/vpp/binapi/stats && pkgreflect
	cd plugins/vpp/binapi/stn && pkgreflect
	cd plugins/vpp/binapi/tap && pkgreflect
	cd plugins/vpp/binapi/tapv2 && pkgreflect
	cd plugins/vpp/binapi/vpe && pkgreflect
	cd plugins/vpp/binapi/vxlan && pkgreflect
	@echo "=> applying fix patch"
	patch -p1 -i plugins/vpp/binapi/fixapi.patch

verify-binapi:
	@echo "=> verifying binary api"
	cd docker/dev && docker build --file Dockerfile --target verify \
			--build-arg VPP_REPO_URL=${VPP_REPO_URL} \
			--build-arg VPP_COMMIT=${VPP_COMMIT} \
	 	../..

get-bindata:
	go get -v github.com/jteeuwen/go-bindata/...
	go get -v github.com/elazarl/go-bindata-assetfs/...

bindata: get-bindata
	cd plugins/restplugin && go generate

# Get dependency manager tool
get-dep:
	go get -v github.com/golang/dep/cmd/dep

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
	dep ensure -dry-run -no-vendor

# Get linter tools
get-linters:
	@echo "=> installing linters"
	go get -v github.com/alecthomas/gometalinter
	gometalinter --install

# Run linters
lint: get-linters
	@echo "=> running code analysis"
	./scripts/static_analysis.sh golint vet

# Format code
format:
	@echo "=> formatting the code"
	./scripts/gofmt.sh

# Get link check tool
get-linkcheck:
	sudo apt-get install npm
	npm install -g markdown-link-check

# Validate links in markdown files
check-links: get-linkcheck
	./scripts/check_links.sh

# Travis
travis:
	@echo "=> TRAVIS BUILD: #$$TRAVIS_BUILD_NUMBER ($$TRAVIS_BUILD_ID) $$TRAVIS_BUILD_STAGE_NAME"
	@echo "Job: #$$TRAVIS_JOB_NUMBER ($$TRAVIS_JOB_ID) AllowFailure: $$TRAVIS_ALLOW_FAILURE ($$TRAVIS_TEST_RESULT)"
	@echo "Repo: $$TRAVIS_REPO_SLUG"
	@echo "Type: $$TRAVIS_EVENT_TYPE PullRequest: $$TRAVIS_PULL_REQUEST"
	@echo "Branch: $$TRAVIS_BRANCH"
	@echo "Commit: $$TRAVIS_COMMIT"
	@echo
	@echo "Range: $$TRAVIS_COMMIT_RANGE"
	@echo "$$TRAVIS_COMMIT_MESSAGE"


.PHONY: build clean \
	install cmd examples clean-examples test \
	test-cover test-cover-html test-cover-xml \
	generate genereate-binapi generate-proto get-binapi-generators get-proto-generators \
	get-dep dep-install dep-update dep-check \
	get-linters lint format \
	get-linkcheck check-links \
	travis
