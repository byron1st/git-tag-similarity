# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Linker flags to inject version information
LDFLAGS := -ldflags "\
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildDate=$(BUILD_DATE)"

# Binary name
BINARY_NAME := git-tag-similarity

.PHONY: all build install clean test fmt lint mockgen help

all: build

## build: Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

## install: Install the binary to $GOPATH/bin
install:
	go install $(LDFLAGS) .

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	go clean

## test: Run all tests
test:
	go test -v ./...

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