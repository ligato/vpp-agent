# generate go structures from proto files
define generate_sources
	$(if $(shell command -v protoc --gogo_out=. 2> /dev/null),$(info gogo/protobuf is installed),$(error gogo/protobuf missing, please install it with go get github.com/gogo/protobuf))
	@echo "# generating sources"
	@cd phonebook/view && go generate -v
endef

all: generate build

generate:
	$(call generate_sources)
	@echo "# done"

clean:
	rm -f phonebook/editor/editor
	rm -f phonebook/view/view
	rm -f phonebook/watcher/watcher

build:
	cd phonebook/view && go build
	cd phonebook/editor && go build
	cd phonebook/watcher && go build
