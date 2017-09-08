include Makeroutines.mk

VERSION=$(shell git rev-parse HEAD)
DATE=$(shell date +'%Y-%m-%dT%H:%M%:z')
LDFLAGS=-ldflags '-X github.com/ligato/vpp-agent/vendor/github.com/ligato/cn-infra/core.BuildVersion=$(VERSION) -X github.com/ligato/vpp-agent/vendor/github.com/ligato/cn-infra/core.BuildDate=$(DATE)'
COVER_DIR=/tmp/


# generate go structures from proto files
define generate_sources
	$(call install_generators)
	@echo "# installing generic"
	@cd vendor/github.com/taylorchu/generic/cmd/generic/ && go install -v
	@cd vendor/github.com/ungerik/pkgreflect/ && go install -v
	@echo "# installing gomock"
	@cd vendor/github.com/golang/mock/gomock && go install -v
	@cd vendor/github.com/golang/mock/mockgen && go install -v
	@echo "# generating sources"
	@cd plugins/linuxplugin && go generate
	@cd plugins/defaultplugins/aclplugin && go generate
	@cd plugins/defaultplugins/ifplugin && go generate
	@cd plugins/defaultplugins/l2plugin && go generate
	@cd plugins/defaultplugins/l3plugin && go generate
	@cd plugins/defaultplugins/aclplugin/bin_api/acl && pkgreflect
	@cd plugins/defaultplugins/ifplugin/bin_api/af_packet && pkgreflect
	@cd plugins/defaultplugins/ifplugin/bin_api/bfd && pkgreflect
	@cd plugins/defaultplugins/ifplugin/bin_api/interfaces && pkgreflect
	@cd plugins/defaultplugins/ifplugin/bin_api/ip && pkgreflect
	@cd plugins/defaultplugins/ifplugin/bin_api/memif && pkgreflect
	@cd plugins/defaultplugins/ifplugin/bin_api/tap && pkgreflect
	@cd plugins/defaultplugins/ifplugin/bin_api/vpe && pkgreflect
	@cd plugins/defaultplugins/ifplugin/bin_api/vxlan && pkgreflect
	@cd plugins/defaultplugins/l2plugin/bin_api/l2 && pkgreflect
	@cd plugins/defaultplugins/l2plugin/bin_api/vpe && pkgreflect
	@cd plugins/defaultplugins/l3plugin/bin_api/ip && pkgreflect
	@echo "# done"
endef

# install-only binaries
define install_only
	@echo "# installing vpp-agent"
	@cd cmd/vpp-agent && go install -v ${LDFLAGS}
	@echo "# installing vpp-agent-ctl"
	@cd cmd/vpp-agent-ctl && go install -v
	@echo "# installing agentctl"
    @cd cmd/agentctl && go install -v
	@echo "# done"
endef

# run all tests
define test_only
	@echo "# running unit tests"
	@go test ./cmd/agentctl/utils
	@go test ./idxvpp/nametoidx
	@echo "# done"
endef

# run all tests with coverage
define test_cover_only
	@echo "# running unit tests with coverage analysis"
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit1.out ./cmd/agentctl/utils
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit2.out ./idxvpp/nametoidx
	@echo "# merging coverage results"
    @cd vendor/github.com/wadey/gocovmerge && go install -v
    @gocovmerge ${COVER_DIR}coverage_unit1.out ${COVER_DIR}coverage_unit2.out  > ${COVER_DIR}coverage.out
    @echo "# coverage data generated into ${COVER_DIR}coverage.out"
    @echo "# done"
endef

# run all tests with coverage and display HTML report
define test_cover_html
    $(call test_cover_only)
    @go tool cover -html=${COVER_DIR}coverage.out -o ${COVER_DIR}coverage.html
    @echo "# coverage report generated into ${COVER_DIR}coverage.html"
    @go tool cover -html=${COVER_DIR}coverage.out
endef

# run all tests with coverage and display XML report
define test_cover_xml
	$(call test_cover_only)
    @gocov convert ${COVER_DIR}coverage.out | gocov-xml > ${COVER_DIR}coverage.xml
    @echo "# coverage report generated into ${COVER_DIR}coverage.xml"
endef

# run code analysis
define lint_only
   @echo "# running code analysis"
    @./scripts/golint.sh
    @./scripts/govet.sh
    @echo "# done"
endef

# run code formatter
define format_only
    @echo "# formatting the code"
    @./scripts/gofmt.sh
    @echo "# done"
endef

# run test examples
define test_examples
    @echo "# Testing examples"
    @echo "# done"
endef

# build examples only
define build_examples_only
    @echo "# building examples"
    @cd examples/govpp_call && go build -v -i
    @cd examples/idx_bd_cache && go build -v -i
    @cd examples/idx_iface_cache && go build -v -i
    @cd examples/idx_mapping_lookup && go build -v -i
    @cd examples/idx_mapping_watcher && go build -v -i
    @cd examples/localclient_linux && go build -v -i
    @cd examples/localclient_vpp && go build -v -i
    @echo "# done"
endef

# build vpp agent only
define build_vpp_agent_only
    @echo "# building vpp agent"
    @cd cmd/vpp-agent && go build -v -i ${LDFLAGS}
    @echo "# done"
endef

# verify that links in markdown files are valid
# requires npm install -g markdown-link-check
define check_links_only
    @echo "# checking links"
    @./scripts/check_links.sh
    @echo "# done"
endef

# build vpp-agent-ctl only
define build_vpp_agent_ctl_only
    @echo "# building vpp-agent-ctl"
    @cd cmd/vpp-agent-ctl && go build -v -i
    @echo "# done"
endef

# build-only agentctl
define build_agentctl_only
 	@echo "# building agentctl"
 	@cd cmd/agentctl && go build -v -i
 	@echo "# done"
endef

# clean examples only
define clean_examples_only
    @echo "# cleaning examples"
    @rm -f examples/govpp_call/govpp_call
    @rm -f examples/idx_bd_cache/idx_bd_cache
    @rm -f examples/idx_iface_cache/idx_iface_cache
    @rm -f examples/idx_mapping_lookup/idx_mapping_lookup
    @rm -f examples/idx_mapping_watcher/idx_mapping_watcher
    @rm -f examples/localclient_linux/localclient_linux
    @rm -r examples/localclient_vpp/localclient_vpp
    @echo "# done"
endef

# build all binaries
build:
	$(call build_examples_only)
	$(call build_vpp_agent_only)
	$(call build_vpp_agent_ctl_only)
	$(call build_agentctl_only)

# build vpp-agent
vpp-agent:
	$(call build_vpp_agent_only)

# build vpp-agent-ctl
vpp-agent-ctl:
	$(call build_vpp_agent_ctl_only)

# build agentctl
agentctl:
	$(call build_agentctl_only)

# build examples
example:
	$(call build_examples_only)

# install binaries
install:
	$(call install_only)

# install dependencies
install-dep:
	$(call install_dependencies)

# update dependencies
update-dep:
	$(call update_dependencies)

# generate structures
generate:
	$(call generate_sources)

# run tests
test:
	$(call test_only)

# run smoke tests on examples
test-examples:
	$(call test_examples)

# run tests with coverage report
test-cover:
	$(call test_cover_only)

# run tests with HTML coverage report
test-cover-html:
	$(call test_cover_html)

# run tests with XML coverage report
test-cover-xml:
	$(call test_cover_xml)

# run & print code analysis
lint:
	$(call lint_only)

# format the code
format:
	$(call format_only)

# validate links in markdown files
check_links:
	$(call check_links_only)

# clean
clean:
	$(call clean_examples_only)
	rm -f cmd/vpp-agent/vpp-agent
	rm -f cmd/vpp-agent-ctl/vpp-agent-ctl
	rm -f cmd/agentctl/agentctl
	@echo "# cleanup completed"

# run all targets
all:
	$(call lint_only)
	$(call build_vpp_agent_only)
	$(call build_vpp_agent_ctl_only)
	$(call test_only)
	$(call install_only)

.PHONY: build update-dep install-dep test lint clean
