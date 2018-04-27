VERSION	:= $(shell git describe --always --tags --dirty)
COMMIT	:= $(shell git rev-parse HEAD)
DATE	:= $(shell date +'%Y-%m-%dT%H:%M%:z')

CNINFRA_CORE := github.com/ligato/vpp-agent/vendor/github.com/ligato/cn-infra/core
LDFLAGS	= -X $(CNINFRA_CORE).BuildVersion=$(VERSION) -X $(CNINFRA_CORE).CommitHash=$(COMMIT) -X $(CNINFRA_CORE).BuildDate=$(DATE)

ifeq ($(STRIP), y)
LDFLAGS += -w -s
endif

COVER_DIR ?= /tmp/

# Build all
build: cmd examples

# Clean all
clean: clean-cmd clean-examples

# Install commands
install:
	@echo " => installing commands"
	go install -v -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ./cmd/vpp-agent
	go install -v -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ./cmd/vpp-agent-grpc
	go install -v -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ./cmd/vpp-agent-ctl
	go install -v -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}" ./cmd/agentctl

# Build commands
cmd:
	@echo " => building commands"
	cd cmd/vpp-agent 		&& go build -v -i -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}"
	cd cmd/vpp-agent-grpc	&& go build -v -i -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}"
	cd cmd/vpp-agent-ctl	&& go build -v -i -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}"
	cd cmd/agentctl 		&& go build -v -i -ldflags "${LDFLAGS}" -tags="${GO_BUILD_TAGS}"

# Clean commands
clean-cmd:
	@echo " => cleaning binaries"
	rm -f ./cmd/vpp-agent/vpp-agent
	rm -f ./cmd/vpp-agent-grpc/vpp-agent-grpc
	rm -f ./cmd/vpp-agent-ctl/vpp-agent-ctl
	rm -f ./cmd/agentctl/agentctl

# Build examples
examples:
	@echo " => building examples"
	cd examples/govpp_call 		    	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_bd_cache 	    	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_iface_cache     	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_mapping_lookup  	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_mapping_watcher     && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_linux/veth 	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_linux/tap 	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_vpp/plugins && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_vpp/nat     && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/remoteclient_grpc_vpp   && go build -v -i -tags="${GO_BUILD_TAGS}"

# Clean examples
clean-examples:
	@echo " => cleaning examples"
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
	@echo " => running scenario tests"
	go test -tags="${GO_BUILD_TAGS}" ./tests/go/itest
	@echo " => running unit tests"
	go test ./cmd/agentctl/utils
	go test ./idxvpp/nametoidx
	go test ./plugins/defaultplugins/l2plugin/bdidx
	go test ./plugins/defaultplugins/l2plugin/vppcalls
	go test ./plugins/defaultplugins/l2plugin/vppdump
	go test ./plugins/defaultplugins/ifplugin/vppcalls

# Get coverage report tools
get-covtools:
	go get -v github.com/wadey/gocovmerge

# Run coverage report
test-cover: get-covtools
	@echo " => running unit tests with coverage analysis"
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_scenario.out -tags="${GO_BUILD_TAGS}" ./tests/go/itest
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit1.out ./cmd/agentctl/utils
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit2.out ./idxvpp/nametoidx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_aclplugin.out -tags=mockvpp ./plugins/defaultplugins/aclplugin
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_aclplugin_aclidx.out -tags=mockvpp ./plugins/defaultplugins/aclplugin/aclidx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_aclplugin_vppcalls.out -tags=mockvpp ./plugins/defaultplugins/aclplugin/vppcalls
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_aclplugin_vppdump.out -tags=mockvpp ./plugins/defaultplugins/aclplugin/vppdump
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_ifplugin.out -tags=mockvpp ./plugins/defaultplugins/ifplugin
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_ifplugin_ifaceidx.out ./plugins/defaultplugins/ifplugin/ifaceidx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_ifplugin_vppcalls.out ./plugins/defaultplugins/ifplugin/vppcalls
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_ifplugin_vppdump.out ./plugins/defaultplugins/ifplugin/vppdump
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_ipsecplugin.out -tags=mockvpp ./plugins/defaultplugins/ipsecplugin
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_ipsecplugin_ipsecidx.out ./plugins/defaultplugins/ipsecplugin/ipsecidx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_ipsecplugin_vppcalls.out ./plugins/defaultplugins/ipsecplugin/vppcalls
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l2plugin.out -tags=mockvpp ./plugins/defaultplugins/l2plugin
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l2plugin_bdidx.out ./plugins/defaultplugins/l2plugin/bdidx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l2plugin_vppcalls.out ./plugins/defaultplugins/l2plugin/vppcalls
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l2plugin_vppdump.out ./plugins/defaultplugins/l2plugin/vppdump
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l3plugin.out -tags=mockvpp ./plugins/defaultplugins/l3plugin
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l3plugin_l3idx.out ./plugins/defaultplugins/l3plugin/l3idx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l3plugin_vppcalls.out ./plugins/defaultplugins/l3plugin/vppcalls
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l3plugin_vppdump.out ./plugins/defaultplugins/l3plugin/vppdump
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l4plugin.out -tags=mockvpp ./plugins/defaultplugins/l4plugin
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l4plugin_nsidx.out ./plugins/defaultplugins/l4plugin/nsidx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_l4plugin_vppcalls.out ./plugins/defaultplugins/l4plugin/vppcalls
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_linux_ifplugin.out -tags=mockvpp ./plugins/linuxplugin/ifplugin
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_linux_ifplugin_ifaceidx.out ./plugins/linuxplugin/ifplugin/ifaceidx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_linux_ifplugin_linuxcalls.out ./plugins/linuxplugin/ifplugin/linuxcalls
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_linux_l3plugin.out -tags=mockvpp ./plugins/linuxplugin/l3plugin
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_linux_l3plugin_l3idx.out ./plugins/linuxplugin/l3plugin/l3idx
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_linux_l3plugin_linuxcalls.out ./plugins/linuxplugin/l3plugin/linuxcalls
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_linux_nsplugin.out ./plugins/linuxplugin/nsplugin
	@echo " => merging coverage results"
	gocovmerge \
			${COVER_DIR}coverage_scenario.out \
			${COVER_DIR}coverage_unit1.out \
			${COVER_DIR}coverage_unit2.out \
			${COVER_DIR}coverage_aclplugin.out \
			${COVER_DIR}coverage_aclplugin_aclidx.out \
			${COVER_DIR}coverage_aclplugin_vppcalls.out \
			${COVER_DIR}coverage_aclplugin_vppdump.out \
			${COVER_DIR}coverage_ifplugin.out \
			${COVER_DIR}coverage_ifplugin_ifaceidx.out \
			${COVER_DIR}coverage_ifplugin_vppcalls.out \
			${COVER_DIR}coverage_ifplugin_vppdump.out \
			${COVER_DIR}coverage_l2plugin.out \
			${COVER_DIR}coverage_l2plugin_bdidx.out \
			${COVER_DIR}coverage_l2plugin_vppcalls.out \
			${COVER_DIR}coverage_l2plugin_vppdump.out \
			${COVER_DIR}coverage_l3plugin.out \
			${COVER_DIR}coverage_l3plugin_l3idx.out \
			${COVER_DIR}coverage_l3plugin_vppcalls.out \
			${COVER_DIR}coverage_l3plugin_vppdump.out \
			${COVER_DIR}coverage_l4plugin.out \
			${COVER_DIR}coverage_l4plugin_nsidx.out \
			${COVER_DIR}coverage_l4plugin_vppcalls.out \
			${COVER_DIR}coverage_linux_ifplugin.out \
			${COVER_DIR}coverage_linux_ifplugin_ifaceidx.out \
			${COVER_DIR}coverage_linux_ifplugin_linuxcalls.out \
			${COVER_DIR}coverage_linux_l3plugin.out \
			${COVER_DIR}coverage_linux_l3plugin_l3idx.out \
			${COVER_DIR}coverage_linux_l3plugin_linuxcalls.out \
			${COVER_DIR}coverage_linux_nsplugin.out \
		> ${COVER_DIR}coverage.out
	@echo " => coverage data generated into ${COVER_DIR}coverage.out"

test-cover-html: test-cover
	go tool cover -html=${COVER_DIR}coverage.out -o ${COVER_DIR}coverage.html
	@echo " => coverage report generated into ${COVER_DIR}coverage.html"

test-cover-xml: test-cover
	gocov convert ${COVER_DIR}coverage.out | gocov-xml > ${COVER_DIR}coverage.xml
	@echo " => coverage report generated into ${COVER_DIR}coverage.xml"

# Get generator tools
get-generators:
	go install -v ./vendor/github.com/gogo/protobuf/protoc-gen-gogo
	go install -v ./vendor/git.fd.io/govpp.git/cmd/binapi-generator
	go install -v ./vendor/github.com/ungerik/pkgreflect

# Generate sources
generate: get-generators
	@echo " => generating sources"
	cd plugins/linuxplugin && go generate
	cd plugins/defaultplugins/aclplugin && go generate
	cd plugins/defaultplugins/ifplugin && go generate
	cd plugins/defaultplugins/ipsecplugin && go generate
	cd plugins/defaultplugins/l2plugin && go generate
	cd plugins/defaultplugins/l3plugin && go generate
	cd plugins/defaultplugins/l4plugin && go generate
	cd plugins/defaultplugins/rpc && go generate
	cd plugins/linuxplugin/ifplugin && go generate
	cd plugins/linuxplugin/l3plugin && go generate
	cd plugins/defaultplugins/common/bin_api/acl && pkgreflect
	cd plugins/defaultplugins/common/bin_api/af_packet && pkgreflect
	cd plugins/defaultplugins/common/bin_api/bfd && pkgreflect
	cd plugins/defaultplugins/common/bin_api/dhcp && pkgreflect
	cd plugins/defaultplugins/common/bin_api/interfaces && pkgreflect
	cd plugins/defaultplugins/common/bin_api/ip && pkgreflect
	cd plugins/defaultplugins/common/bin_api/ipsec && pkgreflect
	cd plugins/defaultplugins/common/bin_api/l2 && pkgreflect
	cd plugins/defaultplugins/common/bin_api/memif && pkgreflect
	cd plugins/defaultplugins/common/bin_api/nat && pkgreflect
	cd plugins/defaultplugins/common/bin_api/session && pkgreflect
	cd plugins/defaultplugins/common/bin_api/stats && pkgreflect
	cd plugins/defaultplugins/common/bin_api/stn && pkgreflect
	cd plugins/defaultplugins/common/bin_api/tap && pkgreflect
	cd plugins/defaultplugins/common/bin_api/tapv2 && pkgreflect
	cd plugins/defaultplugins/common/bin_api/vpe && pkgreflect
	cd plugins/defaultplugins/common/bin_api/vxlan && pkgreflect

# Get dependency manager tool
get-dep:
	go get -v github.com/golang/dep/cmd/dep

# Install the project's dependencies
dep-install: get-dep
	@echo " => installing project's dependencies"
	dep ensure

# Update the locked versions of all dependencies
dep-update: get-dep
	@echo " => updating all dependencies"
	dep ensure -update

# Get linter tools
get-linters:
	@echo " => installing linters"
	go get -v github.com/alecthomas/gometalinter
	gometalinter --install

# Run linters
lint: get-linters
	@echo " => running code analysis"
	./scripts/static_analysis.sh golint vet

# Format code
format:
	@echo " => formatting the code"
	./scripts/gofmt.sh

# Get link check tool
get-linkcheck:
	sudo apt-get install npm
	npm install -g markdown-link-check

# Validate links in markdown files
check-links: get-linkcheck
	./scripts/check_links.sh

.PHONY: build clean \
	install cmd examples clean-examples test \
	get-covtools test-cover test-cover-html test-cover-xml \
	get-generators generate \
	get-dep dep-install dep-update \
	get-linters lint format \
	get-linkcheck check-links
