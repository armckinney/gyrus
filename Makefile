# Gyrus Makefile

all: build

build:
	@echo "Building gyrus binary..."
	@go build -o gyrus cmd/gyrus/main.go

test:
	@echo "Running Gyrus tests..."
	@go test ./... -v

fmt:
	@echo "Formatting Go code..."
	@go fmt ./...

lint:
	@echo "Linting Go code..."
	@gofmt -l .
	@go vet ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf gyrus bin/ dist/ coverage.out coverage.html

run:
	@go run cmd/gyrus/main.go

.PHONY: all build test fmt lint clean run
