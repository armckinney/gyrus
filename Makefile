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

# verge copied:
# .PHONY: all build build-snapshot release test coverage fmt lint clean

# all: build

# build:
# 	@echo "Building..."
# 	@go build -o verge ./cmd/verge

# build-snapshot:
# 	@echo "Building snapshot..."
# 	@goreleaser build --single-target --snapshot --clean

# release:
# 	@goreleaser release --clean

# test:
# 	@echo "Testing..."
# 	@go test ./... -v

# coverage:
# 	@echo "Running coverage..."
# 	@go test ./... -coverprofile=coverage.out
# 	@go tool cover -func=coverage.out

# fmt:
# 	@echo "Formatting..."
# 	@go fmt ./...

# lint:
# 	@echo "Linting..."
# 	@gofmt -l .
# 	@go vet ./...

# clean:
# 	@echo "Cleaning..."
# 	@rm -rf verge dist/ coverage.out coverage.html
