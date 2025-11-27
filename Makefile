# Binary name
BINARY_NAME := git-tag-similarity

# Version information from git
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
COMMIT_TIME ?= $(shell git log -1 --format=%cI 2>/dev/null || echo "unknown")

# Ldflags to inject version information
LDFLAGS := -ldflags "\
	-X 'github.com/byron1st/git-tag-similarity/internal.Version=$(VERSION)' \
	-X 'github.com/byron1st/git-tag-similarity/internal.Commit=$(COMMIT)' \
	-X 'github.com/byron1st/git-tag-similarity/internal.CommitTime=$(COMMIT_TIME)'"

.PHONY: all build install clean test fmt lint mockgen help

all: build

## build: Build the binary with VCS stamping
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