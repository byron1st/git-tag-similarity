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

### Direct Build (without Make)

```bash
go build -o git-tag-similarity .
```

## Usage

The application uses a command-based interface with three commands: `compare`, `help`, and `version`.

### Compare Two Tags

```bash
# Basic comparison (similarity only)
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0

# Verbose comparison (includes list of different commits)
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -v
```

### Show Help

```bash
git-tag-similarity help
# Or
git-tag-similarity
```

### Show Version

```bash
git-tag-similarity version
```

### Get Help for a Specific Command

```bash
git-tag-similarity compare -h
```

### Output Examples

#### Basic Output (without -v flag)
```
Comparing tags: v1.0.0 vs v2.0.0
Similarity: 85.50%

Summary:
  Total commits in [v1.0.0]: 150
  Total commits in [v2.0.0]: 180
  Shared commits: 140
  Unique to [v1.0.0]: 10
  Unique to [v2.0.0]: 40
```

#### Verbose Output (with -v flag)
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

This project uses **Go's embedded build information** (`runtime/debug.ReadBuildInfo()`) for version management. The version information is automatically embedded in the binary during the build process.

#### How It Works

1. **VCS Information**: Go automatically embeds version control information when building:
   - **Commit hash**: From `git rev-parse HEAD`
   - **Commit time**: From `git log -1 --format=%cI`
   - **Modified status**: Whether there are uncommitted changes
   - **Module version**: From `go.mod` or git tags

2. **Automatic Embedding**: When you build with `go build`, Go automatically includes:
   - VCS type (git)
   - Revision (commit hash)
   - Commit timestamp
   - Dirty flag (uncommitted changes)

3. **Version from Git Tags**:
   - Use semantic versioning tags in your repository
   - The version shown will be the module version or "dev" if not tagged

#### Creating a Release

```bash
# Tag a new version
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Update go.mod with the version (if needed)
go mod edit -require=github.com/byron1st/git-tag-similarity@v1.0.0

# Build the binary
make build

# Check version (will show v1.0.0)
./git-tag-similarity version
```

#### Version Output Example

```
git-tag-similarity version v1.0.0+dirty
  Commit: cfd009b (modified)
  Commit time: 2025-10-24 07:18:08 UTC
  Go version: go1.25.2
  OS/Arch: darwin/arm64
```

### Project Structure

```
git-tag-similarity/
├── main.go                    # Main entry point
├── internal/                  # Internal package (not importable)
│   ├── cli.go                # CLI configuration and flag parsing
│   ├── cli_test.go           # CLI tests
│   ├── repository.go         # Repository interface and implementation
│   ├── similarity.go         # Jaccard similarity calculation
│   ├── similarity_test.go    # Similarity tests
│   └── version.go            # Version information
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

