package main

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version (set via -ldflags at build time)
	Version = "dev"
	// Commit is the git commit hash (set via -ldflags at build time)
	Commit = "none"
	// BuildDate is the build date (set via -ldflags at build time)
	BuildDate = "unknown"
)

// PrintVersion prints the version information
func PrintVersion() {
	fmt.Printf("git-tag-similarity version %s\n", Version)
	fmt.Printf("  Commit: %s\n", Commit)
	fmt.Printf("  Built: %s\n", BuildDate)
	fmt.Printf("  Go version: %s\n", runtime.Version())
	fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
