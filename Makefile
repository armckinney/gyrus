# Build the application
all: build

test:
	@echo "Testing..."
	@./tests/db/setup.sh
	@go test ./... -v

fmt:
	@echo "Formatting..."
	@go fmt ./...

lint:
	@echo "Linting..."
	@gofmt -l .
	@go vet ./...

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

clean:
	@echo "Cleaning..."
	@rm -rf main tmp

run:
	@go run cmd/api/main.go

watch:
	@echo "Access app on http://localhost:8080"
	@go tool air

.PHONY: all build run test clean watch
