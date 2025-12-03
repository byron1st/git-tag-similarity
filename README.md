# Git Tag Similarity

A Go application that compares two Git tags and calculates their similarity based on commit history using the Jaccard similarity coefficient.

## Features

- Compare any two Git tags in a repository
- Calculate similarity score based on shared commit history
- Filter comparisons by specific directories or paths
- Show commits unique to each tag
- Display detailed commit information
- AI-powered markdown report generation (supports Claude, OpenAI, Gemini)
- Automated CI/CD with GitHub Actions

## Installation

### Using go install

```bash
go install github.com/byron1st/git-tag-similarity@latest
```

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

## Usage

The application uses a command-based interface with four commands: `compare`, `config`, `help`, and `version`.

### Compare Two Tags

```bash
# Basic comparison (similarity only)
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0

# Verbose comparison (includes list of different commits)
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -v

# Compare with directory filter (only commits touching specific directory)
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -d src/api

# Generate AI-powered markdown report
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -r report.md

# Combine verbose and directory filter
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -v -d internal
```

### Configure AI Settings

Before using the AI report generation feature, you need to configure your AI provider settings:

```bash
# Configure Claude (default model: claude-sonnet-4-5-20250929)
git-tag-similarity config -provider claude -api-key sk-ant-...

# Configure OpenAI with custom model
git-tag-similarity config -provider openai -api-key sk-... -model gpt-4o

# Configure Gemini (default model: gemini-2.0-flash-001)
git-tag-similarity config -provider gemini -api-key AIza...
```

Configuration is stored in `~/.git-tag-similarity/config.json` and will be used for all subsequent report generations.

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
git-tag-similarity config -h
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

- Go 1.23 or higher
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
  Go version: go1.23.0
  OS/Arch: darwin/arm64
```

### Project Structure

```
git-tag-similarity/
├── main.go                    # Main entry point (minimal, orchestration only)
├── internal/                  # Internal package (all implementation details)
│   ├── cli.go                # Command parsing (compare, config, help, version)
│   ├── compare.go            # Compare command logic and configuration
│   ├── compare_test.go       # Compare logic tests
│   ├── config.go             # AI configuration management
│   ├── help.go               # Usage and help message printing
│   ├── report.go             # AI-powered report generation
│   ├── repository.go         # Repository interface + GitRepository implementation
│   ├── repository_test.go    # Repository unit tests
│   ├── similarity.go         # Jaccard similarity calculation
│   ├── similarity_test.go    # Similarity unit tests
│   └── version.go            # Version info via runtime/debug.ReadBuildInfo()
├── mocks/                    # Generated mocks (go generate)
│   └── repository_mock.go    # Mock Repository (uber-go/mock)
├── .github/                  # GitHub Actions workflows
│   └── workflows/
│       ├── pr-validation.yml # Pull request validation
│       └── release.yml       # Automated release process
├── Makefile                  # Build automation
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── LICENSE                   # MIT License
├── README.md                 # User-facing documentation (this file)
├── CHANGELOG                 # Release history and notes
├── CLAUDE.md                 # Claude Code configuration
└── AGENTS.md                 # Project overview and architecture
```

## Testing

The project includes comprehensive unit tests (33 tests total) covering all major components.

```bash
# Run all tests
make test

# Or use go test directly
go test -v ./...
```

## Architecture

- **Interface-based design**: `Repository` interface allows dependency injection for testing
- **Generated mocks**: Using uber-go/mock for type-safe mocking
- **Automatic VCS stamping**: Version info from `runtime/debug.ReadBuildInfo()`
- **Standard Go project layout**: Code in `internal/` package, entry point in root
- **Separation of concerns**: CLI parsing, compare logic, and help output in separate files
- **AI-powered reports**: Optional AI-generated markdown reports analyzing tag differences
- **Configuration management**: Storage of AI API keys in `~/.git-tag-similarity/config.json`
- **Graceful degradation**: Report generation fails gracefully with warnings if config is missing
- **CI/CD automation**: GitHub Actions for PR validation and automated releases

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

