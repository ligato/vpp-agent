include Makeroutines.mk

COVER_DIR=/tmp/

# run all tests
define test_only
	@echo "# running unit tests"
	@go test ./idxvpp/nametoidx
    @echo "# done"
endef

# run all tests with coverage
define test_cover_only
	@echo "# running unit tests with coverage analysis"
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit1.out ./idxvpp/nametoidx
	@echo "# merging coverage results"
    @cd vendor/github.com/wadey/gocovmerge && go install -v
    @gocovmerge ${COVER_DIR}coverage_unit1.out > ${COVER_DIR}coverage.out
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

# build examples only
define build_examples_only
    @echo "# building examples"
    @cd examples/govpp_call && go build
    @cd examples/idx_bd_cache && go build
    @cd examples/idx_iface_cache && go build
    @cd examples/idx_mapping_lookup && go build
    @cd examples/idx_mapping_watcher && go build
    @echo "# done"
endef

# build vpp agent only
define build_vpp_agent_only
    @echo "# building vpp agent"
    @cd cmd/vpp-agent && go build
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
    @echo "# done"
endef

# build all binaries
build:
	$(call build_examples_only)
	$(call build_vpp_agent_only)

# install dependencies
install-dep:
	$(call install_dependencies)

# update dependencies
update-dep:
	$(call update_dependencies)

# run tests
test:
	$(call test_only)

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


# clean
clean:
	@echo "# cleanup completed"
	$(call clean_examples_only)
	rm -f cmd/vpp-agent/vpp-agent

# run all targets
all:
	$(call lint_only)
	$(call test_only)
	$(call install_only)

.PHONY: build update-dep install-dep test lint clean
