.PHONY: help test bench build clean deps

help:
	@echo "Ethereum Custom Transactions - Optimized"
	@echo ""
	@echo "Commands:"
	@echo "  make deps     - Install dependencies"
	@echo "  make test     - Run tests"
	@echo "  make bench    - Run benchmarks"
	@echo "  make build    - Build all packages"
	@echo "  make clean    - Clean build artifacts"

deps:
	go mod download
	go mod tidy

test:
	go test -v ./...

bench:
	go test -bench=. -benchmem ./pkg/merkle

build:
	go build ./pkg/...
	go build ./internal/...

clean:
	go clean -cache -testcache
	rm -rf bin/

run-basic:
	go run examples/basic/main.go

run-complete:
	go run examples/complete/main.go
