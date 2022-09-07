ifndef CACHE_BIN
$(error CACHE_BIN is not set)
endif
ifndef UNAME_OS
$(error UNAME_OS is not set)
endif
ifndef UNAME_ARCH
$(error UNAME_ARCH is not set)
endif

REMOTE_GIT := https://github.com/ligato/vpp-agent.git
CHECK_BREAKING_BRANCH := master

# https://github.com/bufbuild/buf/releases 20200724
BUF_VERSION := 0.51.1
# https://github.com/protocolbuffers/protobuf-go 20200624
PROTOC_GEN_GO_VERSION ?= v1.27.1
# https://github.com/grpc/grpc-go 20200730
PROTOC_GEN_GO_GRPC_VERSION ?= v1.38.0
# https://github.com/protocolbuffers/protobuf/releases 20200729
PROTOC_VERSION ?= 3.17.3

GO_BINS := $(GO_BINS) \
	buf \
	protoc-gen-buf-check-breaking \
	protoc-gen-buf-check-lint

PROTO_PATH := proto
PROTOC_GEN_GO_OUT := proto

PROTOC_GEN_GO_PARAMETER ?= paths=source_relative

ifeq ($(UNAME_OS),Darwin)
PROTOC_OS := osx
endif
ifeq ($(UNAME_OS),Linux)
PROTOC_OS = linux
endif
PROTOC_ARCH := $(UNAME_ARCH)

IMAGE_DIR=$(BUILD_DIR)/image

.PHONY: buf-image
buf-image: $(BUF)
	@echo "# Building buf image"
	mkdir -p $(IMAGE_DIR)/$(VERSION)
	buf image build -o $(IMAGE_DIR)/$(VERSION)/image.bin
	buf image build -o $(IMAGE_DIR)/$(VERSION)/image.json

# BUF points to the marker file for the installed version.
#
# If BUF_VERSION is changed, the binary will be re-downloaded.
BUF := $(CACHE_VERSIONS)/buf/$(BUF_VERSION)
$(BUF):
	@rm -f $(CACHE_BIN)/buf
	@mkdir -p $(CACHE_BIN)
	curl -sSL \
		"https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(UNAME_OS)-$(UNAME_ARCH)" \
		-o "$(CACHE_BIN)/buf"
	chmod +x "$(CACHE_BIN)/buf"
	@rm -rf $(dir $(BUF))
	@mkdir -p $(dir $(BUF))
	@touch $(BUF)

PROTOC := $(CACHE_VERSIONS)/protoc/$(PROTOC_VERSION)
$(PROTOC):
	@if ! command -v curl >/dev/null 2>/dev/null; then echo "error: curl must be installed"  >&2; exit 1; fi
	@if ! command -v unzip >/dev/null 2>/dev/null; then echo "error: unzip must be installed"  >&2; exit 1; fi
	@rm -f $(CACHE_BIN)/protoc
	@rm -rf $(CACHE_INCLUDE)/google
	@mkdir -p $(CACHE_BIN) $(CACHE_INCLUDE)
	$(eval PROTOC_TMP := $(shell mktemp -d))
	cd $(PROTOC_TMP); curl -sSL https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip -o protoc.zip
	cd $(PROTOC_TMP); unzip protoc.zip && mv bin/protoc $(CACHE_BIN)/protoc && mv include/google $(CACHE_INCLUDE)/google
	@rm -rf $(PROTOC_TMP)
	@rm -rf $(dir $(PROTOC))
	@mkdir -p $(dir $(PROTOC))
	@touch $(PROTOC)

PROTOC_GEN_GO := $(CACHE_VERSIONS)/protoc-gen-go/$(PROTOC_GEN_GO_VERSION)
$(PROTOC_GEN_GO):
	@rm -f $(GOBIN)/protoc-gen-go
	$(eval PROTOC_GEN_GO_TMP := $(shell mktemp -d))
	cd $(PROTOC_GEN_GO_TMP); go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@rm -rf $(PROTOC_GEN_GO_TMP)
	@rm -rf $(dir $(PROTOC_GEN_GO))
	@mkdir -p $(dir $(PROTOC_GEN_GO))
	@touch $(PROTOC_GEN_GO)

PROTOC_GEN_GO_GRPC := $(CACHE_VERSIONS)/protoc-gen-go-grpc/$(PROTOC_GEN_GO_GRPC_VERSION)
$(PROTOC_GEN_GO_GRPC):
	@if ! command -v git >/dev/null 2>/dev/null; then echo "error: git must be installed"  >&2; exit 1; fi
	@rm -f $(GOBIN)/protoc-gen-go-grpc
	$(eval PROTOC_GEN_GO_GRPC_TMP := $(shell mktemp -d))
	#cd $(PROTOC_GEN_GO_GRPC_TMP); go get -u -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	cd $(PROTOC_GEN_GO_GRPC_TMP); git clone -b $(PROTOC_GEN_GO_GRPC_VERSION) https://github.com/grpc/grpc-go
	cd $(PROTOC_GEN_GO_GRPC_TMP); cd grpc-go/cmd/protoc-gen-go-grpc && go install .
	@rm -rf $(PROTOC_GEN_GO_GRPC_TMP)
	@rm -rf $(dir $(PROTOC_GEN_GO_GRPC))
	@mkdir -p $(dir $(PROTOC_GEN_GO_GRPC))
	@touch $(PROTOC_GEN_GO_GRPC)

.PHONY: buf-ls-packages
buf-ls-packages:
	buf image build --exclude-imports --exclude-source-info \
		-o -#format=json | jq '.file[] | .package' | sort | uniq

.PHONY: buf-ls-files
buf-ls-files:
	buf image build --exclude-imports --exclude-source-info \
		-o -#format=json | jq '.file[] | .name' | sort | uniq

# buf-deps allows us to install deps without running any checks.

.PHONY: buf-deps
buf-deps: $(BUF)

# buf-lint is what we run when developing
# this does linting for proto files

.PHONY: buf-lint
buf-lint: $(BUF)
	buf check lint

# buf-breaking-local is what we run when testing locally
# this does breaking change detection against our local git repository

.PHONY: buf-breaking-local
buf-breaking-local: $(BUF)
	buf check breaking --against-input '.git#branch=$(CHECK_BREAKING_BRANCH)'

# buf-breaking is what we run when testing in most CI providers
# this does breaking change detection against our remote git repository

.PHONY: buf-breaking
buf-breaking: $(BUF)
	buf check breaking --timeout 60s \
	 --against-input "$(REMOTE_GIT)#branch=$(CHECK_BREAKING_BRANCH)"


.PHONY: protocgengoclean
protocgengoclean:
	find "$(PROTOC_GEN_GO_OUT)" -name "*.pb.go" -exec rm -drfv '{}' \;

.PHONY: protocgengo
protocgengo: protocgengoclean $(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)
	bash scripts/protoc_gen_go.sh "$(PROTO_PATH)" "$(PROTOC_GEN_GO_OUT)" "$(PROTOC_GEN_GO_PARAMETER)"
