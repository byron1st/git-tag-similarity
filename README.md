# Git Tag Similarity

A Go application that compares two Git tags and calculates their similarity based on commit history using the Jaccard similarity coefficient.

## Features

- Compare any two Git tags in a repository
- Calculate similarity score based on shared commit history
- Show commits unique to each tag
- Display detailed commit information

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/byron1st/git-tag-similarity.git
cd git-tag-similarity

# Build with version information
make build

# Or install to $GOPATH/bin
make install
```

### Manual Build

```bash
go build -o git-tag-similarity .
```

### Build with Custom Version

```bash
go build -ldflags "-X main.Version=v1.0.0 -X main.Commit=$(git rev-parse --short HEAD) -X main.BuildDate=$(date -u '+%Y-%m-%d_%H:%M:%S')" .
```

## Usage

### Basic Usage

```bash
git-tag-similarity -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0
```

### Show Version

```bash
git-tag-similarity -version
```

### Output Example

```
Comparing tags: v1.0.0 vs v2.0.0
Similarity: 85.50%

Summary:
  Total commits in [v1.0.0]: 150
  Total commits in [v2.0.0]: 180
  Shared commits: 140
  Unique to [v1.0.0]: 10
  Unique to [v2.0.0]: 40

Commits only in [v1.0.0] (10):
  - a1b2c3d : Fix authentication bug
  - e4f5g6h : Update dependencies
  ...

Commits only in [v2.0.0] (40):
  - i7j8k9l : Add new feature
  - m0n1o2p : Refactor code
  ...
```

## Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, but recommended)

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Format code
make fmt

# Generate mocks
make mockgen

# Clean build artifacts
make clean

# Show all available targets
make help
```

### Version Management

This project uses **build-time variable injection** for version management. The version information is set during the build process using `-ldflags`.

#### How It Works

1. **Version variables** are defined in `version.go`:
   - `Version`: Semantic version (e.g., "v1.0.0")
   - `Commit`: Git commit hash
   - `BuildDate`: Build timestamp

2. **Build-time injection** via Makefile:
   ```makefile
   VERSION ?= $(shell git describe --tags --always --dirty)
   COMMIT ?= $(shell git rev-parse --short HEAD)
   BUILD_DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
   ```

3. **Automatic version from Git tags**:
   - If you have git tags, the version will be automatically set from `git describe --tags`
   - Without tags, it defaults to "dev"

#### Creating a Release

```bash
# Tag a new version
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Build with the tagged version
make build

# The version will automatically be set to v1.0.0
./git-tag-similarity -version
```

### Project Structure

```
git-tag-similarity/
├── main.go                    # Main entry point
├── version.go                 # Version information
├── internal/                  # Internal package (not importable)
│   ├── cli.go                # CLI configuration and flag parsing
│   ├── cli_test.go           # CLI tests
│   ├── repository.go         # Repository interface and implementation
│   ├── similarity.go         # Jaccard similarity calculation
│   └── similarity_test.go    # Similarity tests
├── mocks/                    # Generated mocks (go generate)
│   └── repository_mock.go
├── Makefile                  # Build automation
├── go.mod                    # Go module definition
└── README.md                 # This file
```

## Testing

```bash
# Run all tests
make test

# Or use go test directly
go test -v ./...
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

