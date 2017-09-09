include Makeroutines.mk

VERSION=$(shell git rev-parse HEAD)
DATE=$(shell date +'%Y-%m-%dT%H:%M%:z')
COVER_DIR=/tmp/
LDFLAGS=-ldflags '-X github.com/ligato/cn-infra/core.BuildVersion=$(VERSION) -X github.com/ligato/cn-infra/core.BuildDate=$(DATE)'

# run all tests
define test_only
	@echo "# running unit tests"
	@go test ./logging/logrus
	@go test ./db/keyval/etcdv3
	@go test ./messaging/kafka/client
    @go test ./messaging/kafka/mux
    @go test ./utils/addrs
    @go test ./core
    @go test ./db/keyval/redis
    @go test ./db/sql/cassandra
    @go test ./idxmap/mem
    @echo "# done"
endef

# run all tests with coverage
define test_cover_only
	@echo "# running unit tests with coverage analysis"
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit1.out ./logging/logrus
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit2.out ./db/keyval/etcdv3
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit3.out ./messaging/kafka/client
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit4.out ./messaging/kafka/mux
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit5.out ./utils/addrs
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit6.out ./core
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit7.out ./db/keyval/redis
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit8.out ./db/sql/cassandra
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit9.out ./idxmap/mem
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit10.out ./tests/go/itest
    @echo "# merging coverage results"
    @cd vendor/github.com/wadey/gocovmerge && go install -v
    @gocovmerge ${COVER_DIR}coverage_unit1.out ${COVER_DIR}coverage_unit2.out ${COVER_DIR}coverage_unit3.out ${COVER_DIR}coverage_unit4.out ${COVER_DIR}coverage_unit5.out ${COVER_DIR}coverage_unit6.out ${COVER_DIR}coverage_unit7.out  ${COVER_DIR}coverage_unit8.out ${COVER_DIR}coverage_unit9.out ${COVER_DIR}coverage_unit10.out > ${COVER_DIR}coverage.out
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

# verify that links in markdown files are valid
# requires npm install -g markdown-link-check
define check_links_only
    @echo "# checking links"
    @./scripts/check_links.sh
    @echo "# done"
endef

# build plugin example only
define build_plugin_examples_only
    @echo "# building plugin examples"
    @cd examples/datasync_plugin && go build -v ${LDFLAGS}
    @cd examples/flags_plugin && go build -v ${LDFLAGS}
    @cd examples/kafka_plugin/non_clustered && go build -v ${LDFLAGS}
    @cd examples/logs_plugin && go build -v ${LDFLAGS}
    @cd examples/simple-agent && go build -v ${LDFLAGS}
    @echo "# done"
endef


# build examples only
define build_examples_only
    @echo "# building examples"
    @cd examples/cassandra_lib && go build
    @cd examples/etcdv3_lib && make build
    @cd examples/kafka_lib && make build
    @cd examples/logs_lib && make build
    @cd examples/redis_lib && make build
    @echo "# done"
endef

# clean examples only
define clean_examples_only
    @echo "# cleaning examples"
    @cd examples/cassandra_lib && rm -f simple
    @cd examples/etcdv3_lib && make clean
    @cd examples/kafka_lib && make clean
    @cd examples/logs_lib && make clean
    @cd examples/redis_lib && make clean
    @echo "# done"
endef

# clean plugin examples only
define clean_plugin_examples_only
    @echo "# cleaning plugin examples"
    @rm -f examples/simple-agent/simple-agent
    @rm -f examples/datasync_plugin/datasync_plugin
    @rm -f examples/flags_plugin/flags_plugin
    @rm -f examples/kafka_plugin/non_clustered/non_clustered
    @rm -f examples/logs_plugin/logs_plugin
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
    @./scripts/test_examples.sh
    @echo "# done"
endef

# build all binaries
build:
	$(call build_examples_only)
	$(call build_plugin_examples_only)

# install dependencies
install-dep:
	$(call install_dependencies)

# update dependencies
update-dep:
	$(call update_dependencies)

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

# validate links in markdown files
check_links:
	$(call check_links_only)

# format the code
format:
	$(call format_only)


# clean
clean:
	@echo "# cleanup completed"
	$(call clean_examples_only)
	$(call clean_plugin_examples_only)
# run all targets
all:
	$(call lint_only)
	$(call test_only)
	$(call install_only)

.PHONY: build update-dep install-dep test lint check_links clean
