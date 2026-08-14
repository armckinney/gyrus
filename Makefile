.PHONY: all fmt lint test test-unit test-integration clean build run

all: clean fmt lint test build

fmt:
	@echo "Formatting Go code..."
	@go fmt ./...

lint:
	@echo "Linting Go code..."
	@gofmt -l .
	@go vet ./...

test:
	@echo "Running all Gyrus tests..."
	@go test ./... -v

test-unit:
	@echo "Running Unit Tests (internal/...)..."
	@go test ./internal/... -v

test-integration: build
	@echo "Running Integration Tests (tests/...)..."
	@RUN_INTEGRATION_TESTS=1 RUN_DOCKER_TESTS=1 go test ./tests/... -v

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf gyrus bin/ dist/ coverage.out coverage.html

build:
	@echo "Building gyrus binary..."
	@go build -o gyrus cmd/gyrus/main.go

run:
	@go run cmd/gyrus/main.go
