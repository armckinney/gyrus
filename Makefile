.PHONY: all fmt lint test test-unit test-integration clean build run install

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
	@GYRUS_WORKSPACE="" GYRUS_CONFIG="" go test ./... -v

test-unit:
	@echo "Running Unit Tests (internal/...)..."
	@GYRUS_WORKSPACE="" GYRUS_CONFIG="" go test ./internal/... -v

test-integration: build
	@echo "Running Integration Tests (tests/...)..."
	@GYRUS_WORKSPACE="" GYRUS_CONFIG="" RUN_INTEGRATION_TESTS=1 RUN_DOCKER_TESTS=1 go test ./tests/... -v

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf gyrus bin/ dist/ coverage.out coverage.html

build:
	@echo "Building gyrus binary..."
	@go build -buildvcs=false -o gyrus cmd/gyrus/main.go

install: build
	@echo "Installing gyrus binary to /usr/local/bin..."
	@install -m 755 gyrus /usr/local/bin/gyrus

run:
	@go run cmd/gyrus/main.go
