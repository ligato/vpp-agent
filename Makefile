VERSION	:= $(shell git describe --always --tags --dirty)
COMMIT	:= $(shell git rev-parse HEAD)
DATE	:= $(shell date +'%Y-%m-%dT%H:%M%:z')

CNINFRA_CORE := github.com/ligato/vpp-agent/vendor/github.com/ligato/cn-infra/core
LDFLAGS	= -ldflags '-X $(CNINFRA_CORE).BuildVersion=$(VERSION) -X $(CNINFRA_CORE).CommitHash=$(COMMIT) -X $(CNINFRA_CORE).BuildDate=$(DATE)'

COVER_DIR ?= /tmp/

# Build all
build: cmd examples

# Clean all
clean: clean-cmd clean-examples

# Install commands
install:
	@echo "# installing commands"
	go install -v ${LDFLAGS} -tags="${GO_BUILD_TAGS}" ./cmd/vpp-agent
	go install -v ${LDFLAGS} -tags="${GO_BUILD_TAGS}" ./cmd/vpp-agent-ctl
	go install -v ${LDFLAGS} -tags="${GO_BUILD_TAGS}" ./cmd/agentctl

# Build commands
cmd:
	@echo "# building commands"
	cd cmd/vpp-agent 		&& go build -v -i ${LDFLAGS} -tags="$(GO_BUILD_TAGS)"
	cd cmd/vpp-agent-ctl	&& go build -v -i ${LDFLAGS} -tags="${GO_BUILD_TAGS}"
	cd cmd/agentctl 		&& go build -v -i ${LDFLAGS} -tags="${GO_BUILD_TAGS}"

# Clean commands
clean-cmd:
	@echo "# cleaning binaries"
	rm -f ./cmd/vpp-agent/vpp-agent
	rm -f ./cmd/vpp-agent-ctl/vpp-agent-ctl
	rm -f ./cmd/agentctl/agentctl

# Build examples
examples:
	@echo "# building examples"
	cd examples/govpp_call 			&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_bd_cache 		&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_iface_cache 	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_mapping_lookup 	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/idx_mapping_watcher && go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_linux 	&& go build -v -i -tags="${GO_BUILD_TAGS}"
	cd examples/localclient_vpp 	&& go build -v -i -tags="${GO_BUILD_TAGS}"

# Clean examples
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
	go test -tags="${GO_BUILD_TAGS}" ./tests/go/itest
	@echo "# running unit tests"
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
	@echo "# running unit tests with coverage analysis"
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_scenario.out -tags="${GO_BUILD_TAGS}" ./tests/go/itest
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
			${COVER_DIR}coverage_l2plugin_bdidx.out \
			${COVER_DIR}coverage_l2plugin_vppcalls.out \
			${COVER_DIR}coverage_l2plugin_vppdump.out  \
			${COVER_DIR}coverage_ifplugin_vppcalls.out \
		> ${COVER_DIR}coverage.out
	@echo "# coverage data generated into ${COVER_DIR}coverage.out"

test-cover-html: test-cover
	go tool cover -html=${COVER_DIR}coverage.out -o ${COVER_DIR}coverage.html
	@echo "# coverage report generated into ${COVER_DIR}coverage.html"

test-cover-xml: test-cover
	gocov convert ${COVER_DIR}coverage.out | gocov-xml > ${COVER_DIR}coverage.xml
	@echo "# coverage report generated into ${COVER_DIR}coverage.xml"

# Get generator tools
get-generators:
	go install -v ./vendor/git.fd.io/govpp.git/cmd/binapi-generator
	go install -v ./vendor/github.com/ungerik/pkgreflect

# Generate sources
generate: get-generators
	@echo "# generating sources"
	cd plugins/linuxplugin && go generate
	cd plugins/defaultplugins/aclplugin && go generate
	cd plugins/defaultplugins/ifplugin && go generate
	cd plugins/defaultplugins/l2plugin && go generate
	cd plugins/defaultplugins/l3plugin && go generate
	cd plugins/defaultplugins/l4plugin && go generate
	cd plugins/linuxplugin/ifplugin && go generate
	cd plugins/linuxplugin/l3plugin && go generate
	cd plugins/defaultplugins/common/bin_api/acl && pkgreflect
	cd plugins/defaultplugins/common/bin_api/af_packet && pkgreflect
	cd plugins/defaultplugins/common/bin_api/bfd && pkgreflect
	cd plugins/defaultplugins/common/bin_api/dhcp && pkgreflect
	cd plugins/defaultplugins/common/bin_api/interfaces && pkgreflect
	cd plugins/defaultplugins/common/bin_api/ip && pkgreflect
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
	@echo "# installing project's dependencies"
	dep ensure

# Update the locked versions of all dependencies
dep-update: get-dep
	@echo "# updating all dependencies"
	dep ensure -update

# Get linter tools
get-linters:
	@echo "# installing linters"
	go get -v github.com/alecthomas/gometalinter
	gometalinter --install

# Run linters
lint: get-linters
	@echo "# running code analysis"
	./scripts/static_analysis.sh golint vet

# Format code
format:
	@echo "# formatting the code"
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
