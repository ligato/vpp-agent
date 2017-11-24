build:
	@cd cmd/binapi-generator && go build -v
	@cd examples/cmd/simple-client && go build -v
	@cd examples/cmd/stats-client && go build -v
	@cd examples/cmd/perf-bench && go build -v

test:
	@cd cmd/binapi-generator && go test -cover .
	@cd api && go test -cover ./...
	@cd core && go test -cover .

install:
	@cd cmd/binapi-generator && go install -v

extras:
	@cd extras/libmemif/examples/raw-data && go build -v
	@cd extras/libmemif/examples/icmp-responder && go build -v

clean:
	@rm -f cmd/binapi-generator/binapi-generator
	@rm -f examples/cmd/simple-client/simple-client
	@rm -f examples/cmd/stats-client/stats-client
	@rm -f examples/cmd/perf-bench/perf-bench
	@rm -f extras/libmemif/examples/raw-data/raw-data
	@rm -f extras/libmemif/examples/icmp-responder/icmp-responder

generate:
	@cd core && go generate ./...
	@cd examples && go generate ./...

lint:
	@golint ./... | grep -v vendor | grep -v bin_api || true

.PHONY: build test install extras clean generate
