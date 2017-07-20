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

clean:
	@rm -f cmd/binapi-generator/binapi-generator
	@rm -f examples/cmd/simple-client/simple-client
	@rm -f examples/cmd/stats-client/stats-client
	@rm -f examples/cmd/perf-bench/perf-bench

generate:
	@cd core && go generate ./...
	@cd examples && go generate ./...

lint:
	@golint ./... | grep -v vendor | grep -v bin_api || true

.PHONY: build test install clean generate
