# generate go structures from proto files
define generate_sources
	$(if $(shell command -v protoc --gogo_out=. 2> /dev/null),$(info gogo/protobuf is installed),$(error gogo/protobuf missing, please install it with go get github.com/gogo/protobuf))
	@echo "# generating sources"
	@cd airport && go generate -v
endef

all: generate build

generate:
	$(call generate_sources)
	@echo "# done"

clean:
	rm -f simple/simple
	rm -f airport/airport

build:
	cd simple && go build
	cd airport && go build
