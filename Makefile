.PHONY: all fmt lint test test-integration clean build install

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
	@GYRUS_WORKSPACE="" GYRUS_CONFIG="" go test ./...

test-integration: build
	@echo "Running Integration Tests (tests/...)..."
	@GYRUS_WORKSPACE="" GYRUS_CONFIG="" RUN_INTEGRATION_TESTS=1 RUN_DOCKER_TESTS=1 go test ./tests/...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf gyrus bin/ dist/ coverage.out coverage.html

build:
	@echo "Building gyrus binary..."
	@go build -buildvcs=false -o gyrus cmd/gyrus/main.go

install: build
	@echo "Symlinking gyrus binary to /usr/local/bin/gyrus..."
	@mkdir -p /usr/local/bin
	@ln -sf $(CURDIR)/gyrus /usr/local/bin/gyrus
	@echo "✓ Installed gyrus: /usr/local/bin/gyrus -> $(CURDIR)/gyrus"
