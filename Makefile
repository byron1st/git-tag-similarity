# Binary name
BINARY_NAME := git-tag-similarity

.PHONY: all build install clean test fmt lint mockgen help

all: build

## build: Build the binary with VCS stamping
build:
	go build -o $(BINARY_NAME) .

## install: Install the binary to $GOPATH/bin
install:
	go install .

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	go clean

## test: Run all tests
test:
	go test ./...

## fmt: Format all Go files
fmt:
	go fmt ./...

## lint: Run linter (requires golangci-lint)
lint:
	golangci-lint run

## mockgen: Generate mocks
mockgen:
	go generate ./...

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'