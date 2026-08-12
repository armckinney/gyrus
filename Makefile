.PHONY: all fmt lint test test-integration test-e2e clean build run

all: clean fmt lint test build

fmt:
	@echo "Formatting Go code..."
	@go fmt ./...

lint:
	@echo "Linting Go code..."
	@gofmt -l .
	@go vet ./...

test:
	@echo "Running Gyrus tests..."
	@go test ./... -v

test-integration: build
	@echo "Running Live Antigravity Integration Tests..."
	@RUN_INTEGRATION_TESTS=1 go test ./tests/skills/... -v -run TestAntigravityIntegration

test-e2e: test-integration

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf gyrus bin/ dist/ coverage.out coverage.html

build:
	@echo "Building gyrus binary..."
	@go build -o gyrus cmd/gyrus/main.go

run:
	@go run cmd/gyrus/main.go
