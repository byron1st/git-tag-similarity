# Overview

This application compares two Git tags in a repository and calculates their similarity based on commit history using the Jaccard similarity coefficient.

# Coding Style

*   Follow standard Go conventions and formatting (`make lint` to verify).
*   Wrap returned errors with `errors.Join` using a distinct `Err...` variable.
*   Use `defer func() { _ = closer.Close() }()` when closing resources.
*   Always specify parameter types in function signatures.
*   Use "range over integers" for `for` loops where appropriate.

# Commands

*   `make mockgen`: Generate mocks.
*   `make fmt`: Check code quality.
*   `make test`: Run all tests.
*   `make build`: Build the binary.
*   `make help`: Show all available make targets.

# Final Project Structure

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
├── README.md                 # User-facing documentation
├── CHANGELOG                 # Release history and notes
├── PLANS.md                  # Development plan
├── CLAUDE.md                 # Claude Code configuration
└── AGENTS.md                 # This file
```

# Command-Based CLI

The application uses a command-based interface (like git, docker, kubectl):

**Commands:**
- `compare`: Compare two Git tags (requires: `-repo`, `-tag1`, `-tag2`; optional: `-v`, `-d`, `-r`/`--report`)
- `config`: Configure AI settings for report generation (requires: `-provider`, `-api-key`)
- `help`: Show usage information
- `version`: Show version info (using embedded VCS data)

**Examples:**
```bash
# Compare two tags (basic)
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0

# Compare with verbose output (includes commit lists)
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -v

# Compare with directory filter (only commits touching specific directory)
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -d src/api

# Compare with AI-generated markdown report
git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -r report.md

# Configure AI settings (Claude)
git-tag-similarity config -provider claude -api-key sk-ant-...

# Show help
git-tag-similarity help

# Show version
git-tag-similarity version
```

# Architecture Highlights

1. **Interface-based design**: `Repository` interface allows dependency injection for testing
2. **Generated mocks**: Using uber-go/mock for type-safe mocking
3. **Automatic VCS stamping**: Version info from `runtime/debug.ReadBuildInfo()`
4. **Standard Go project layout**: Code in `internal/` package, entry point in root
5. **Comprehensive testing**: Unit tests for all major components (33 tests total)
6. **Separation of concerns**: CLI parsing, compare logic, and help output are in separate files
7. **CI/CD automation**: GitHub Actions for PR validation and automated releases
8. **Directory filtering**: Optional directory filter for comparing tags based on specific paths
9. **AI-powered reports**: Optional AI-generated markdown reports analyzing tag differences
10. **Configuration management**: Storage of AI API keys in `~/.git-tag-similarity/config.json`
11. **Graceful degradation**: Report generation fails gracefully with warnings if config is missing