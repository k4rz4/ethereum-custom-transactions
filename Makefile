.PHONY: help test bench build clean deps docker-up docker-down docker-logs docker-rebuild

help:
	@echo "Ethereum Custom Transactions - Optimized"
	@echo ""
	@echo "Commands:"
	@echo "  make deps            - Install dependencies"
	@echo "  make test            - Run tests"
	@echo "  make bench           - Run benchmarks"
	@echo "  make build           - Build all packages"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make run-basic       - Run basic example"
	@echo "  make run-complete    - Run complete example"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-up       - Start all Docker services"
	@echo "  make docker-down     - Stop all Docker services"
	@echo "  make docker-logs     - View all service logs"
	@echo "  make docker-rebuild  - Rebuild and restart all services"
	@echo "  make docker-basic    - Run only basic example"
	@echo "  make docker-manager  - Run only manager example"
	@echo "  make docker-batch    - Run only batch example"
	@echo "  make docker-clean    - Remove all containers and volumes"

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

# Docker commands
docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up

docker-basic:
	docker-compose up --build ganache basic-example

docker-manager:
	docker-compose up --build ganache manager-example

docker-batch:
	docker-compose up --build ganache batch-example

docker-clean:
	docker-compose down -v
	docker system prune -fgo
