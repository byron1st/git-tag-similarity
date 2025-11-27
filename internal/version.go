package internal

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// PrintVersion prints the version information retrieved from the binary's build info
func PrintVersion() {
	version := "dev"

	// Read build info from the binary
	if info, ok := debug.ReadBuildInfo(); ok {
		// Get version from module
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}

	// Print version information
	fmt.Printf("git-tag-similarity version %s\n", version)
	fmt.Printf("  Go version: %s\n", runtime.Version())
	fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
