VERSION	:= $(shell git describe --tags --dirty)
COMMIT	:= $(shell git rev-parse HEAD)
DATE	:= $(shell date +'%Y-%m-%dT%H:%M%:z')

CNINFRA_CORE := github.com/ligato/vpp-agent/vendor/github.com/ligato/cn-infra/core
LDFLAGS	= -ldflags '-X $(CNINFRA_CORE).BuildVersion=$(VERSION) -X $(CNINFRA_CORE).CommitHash=$(COMMIT) -X $(CNINFRA_CORE).BuildDate=$(DATE)'

COVER_DIR ?= /tmp/

default: examples build

all: lint test install

clean-all: clean clean-examples

build:
	@echo "# building commands"
	cd cmd/vpp-agent && go build -v -i ${LDFLAGS} -tags="$(GO_BUILD_TAGS)"
	cd cmd/vpp-agent-ctl && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd cmd/agentctl && go build -v -i -tags="${GO_BUILD_TAGS}"

install:
	@echo "# installing commands"
	cd cmd/vpp-agent && go install -v ${LDFLAGS} -tags="${GO_BUILD_TAGS}"
	cd cmd/vpp-agent-ctl && go install -v -tags="${GO_BUILD_TAGS}"
	cd cmd/agentctl && go install -v -tags="${GO_BUILD_TAGS}"

clean:
	@echo "# cleaning binaries"
	rm -f cmd/vpp-agent/vpp-agent
	rm -f cmd/vpp-agent-ctl/vpp-agent-ctl
	rm -f cmd/agentctl/agentctl

examples:
	@echo "# building examples"
	cd examples/govpp_call && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_bd_cache && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_iface_cache && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_mapping_lookup && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_mapping_watcher && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_linux && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_vpp && go build -v -i -tags="${GO_BUILD_TAGS}"

clean-examples:
	@echo "# cleaning examples"
	rm -f examples/govpp_call/govpp_call
	rm -f examples/idx_bd_cache/idx_bd_cache
	rm -f examples/idx_iface_cache/idx_iface_cache
	rm -f examples/idx_mapping_lookup/idx_mapping_lookup
	rm -f examples/idx_mapping_watcher/idx_mapping_watcher
	rm -f examples/localclient_linux/localclient_linux
	rm -r examples/localclient_vpp/localclient_vpp

# Run tests
test:
	@echo "# running scenario tests"
	go test -tags mockvpp ./tests/go/itest
	@echo "# running unit tests"
	go test ./cmd/agentctl/utils
	go test ./idxvpp/nametoidx
	go test ./plugins/defaultplugins/l2plugin/bdidx
	go test ./plugins/defaultplugins/l2plugin/vppcalls
	go test ./plugins/defaultplugins/l2plugin/vppdump
	go test ./plugins/defaultplugins/ifplugin/vppcalls

# Run tests with coverage report
test-cover:
	@echo "# running unit tests with coverage analysis"
	go test -tags mockvpp -covermode=count -coverprofile=${COVER_DIR}coverage_scenario.out ./tests/go/itest
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit1.out ./cmd/agentctl/utils
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit2.out ./idxvpp/nametoidx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l2plugin_bdidx.out ./plugins/defaultplugins/l2plugin/bdidx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l2plugin_vppcalls.out ./plugins/defaultplugins/l2plugin/vppcalls
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l2plugin_vppdump.out ./plugins/defaultplugins/l2plugin/vppdump
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_ifplugin_vppcalls.out ./plugins/defaultplugins/ifplugin/vppcalls
	@echo "# merging coverage results"
	gocovmerge \
			${COVER_DIR}coverage_scenario.out \
			${COVER_DIR}coverage_unit1.out \
			${COVER_DIR}coverage_unit2.out \
			${COVER_DIR}coverage_l2plugin_bdidx.out    > ${COVER_DIR}coverage.out \
			${COVER_DIR}coverage_l2plugin_vppcalls.out > ${COVER_DIR}coverage.out \
			${COVER_DIR}coverage_l2plugin_vppdump.out  > ${COVER_DIR}coverage.out \
			${COVER_DIR}coverage_ifplugin_vppcalls.out > ${COVER_DIR}coverage.out
	@echo "# coverage data generated into ${COVER_DIR}coverage.out"

test-cover-html: test-cover
	go tool cover -html=${COVER_DIR}coverage.out -o ${COVER_DIR}coverage.html
	@echo "# coverage report generated into ${COVER_DIR}coverage.html"

test-cover-xml: test-cover
	gocov convert ${COVER_DIR}coverage.out | gocov-xml > ${COVER_DIR}coverage.xml
	@echo "# coverage report generated into ${COVER_DIR}coverage.xml"

# Get protobuf and binapi-generator
get-generators: dep-install
	cd vendor/git.fd.io/govpp.git/cmd/binapi-generator && go install -v
	cd vendor/github.com/ungerik/pkgreflect && go install -v

# Generate sources
generate: get-generators
	@echo "# generating sources"
	cd plugins/linuxplugin && go generate
	cd plugins/defaultplugins/aclplugin && go generate
	cd plugins/defaultplugins/ifplugin && go generate
	cd plugins/defaultplugins/l2plugin && go generate
	cd plugins/defaultplugins/l3plugin && go generate
	cd plugins/defaultplugins/l4plugin && go generate
	cd plugins/defaultplugins/common/bin_api/acl && pkgreflect
	cd plugins/defaultplugins/common/bin_api/af_packet && pkgreflect
	cd plugins/defaultplugins/common/bin_api/bfd && pkgreflect
	cd plugins/defaultplugins/common/bin_api/interfaces && pkgreflect
	cd plugins/defaultplugins/common/bin_api/ip && pkgreflect
	cd plugins/defaultplugins/common/bin_api/l2 && pkgreflect
	cd plugins/defaultplugins/common/bin_api/memif && pkgreflect
	cd plugins/defaultplugins/common/bin_api/session && pkgreflect
	cd plugins/defaultplugins/common/bin_api/stats && pkgreflect
	cd plugins/defaultplugins/common/bin_api/tap && pkgreflect
	cd plugins/defaultplugins/common/bin_api/tapv2 && pkgreflect
	cd plugins/defaultplugins/common/bin_api/vpe && pkgreflect
	cd plugins/defaultplugins/common/bin_api/vxlan && pkgreflect

get-dep:
	go get github.com/golang/dep/cmd/dep

# Install the project's dependencies
dep-install: get-dep
	dep ensure

# Update the locked versions of all dependencies
dep-update: get-dep
	dep ensure -update

# Validate links in markdown files
check-links:
	npm install -g markdown-link-check
	./scripts/check_links.sh

get-linters:
	@echo "# installing linters"
	go get github.com/alecthomas/gometalinter
	go get github.com/golang/lint/golint
	go get github.com/golang/go/src/cmd/vet

lint: get-linters
	@echo "# running code analysis"
	./scripts/static_analysis.sh golint vet

format:
	@echo "# formatting the code"
	./scripts/gofmt.sh

.PHONY: default all build install clean examples clean-examples test lint format get-generators test-cover test-cover-html test-cover-xml get-dep dep-install dep-update check_links
