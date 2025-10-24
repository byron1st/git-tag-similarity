package internal

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

// PrintVersion prints the version information retrieved from the binary's build info
func PrintVersion() {
	version := "dev"
	commit := "unknown"
	commitTime := "unknown"
	modified := false

	// Read build info from the binary
	if info, ok := debug.ReadBuildInfo(); ok {
		// Get version from module
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}

		// Extract VCS information from build settings
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) >= 7 {
					commit = setting.Value[:7] // Short hash (7 characters)
				} else {
					commit = setting.Value
				}
			case "vcs.time":
				if t, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					commitTime = t.Format("2006-01-02 15:04:05")
				}
			case "vcs.modified":
				modified = setting.Value == "true"
			}
		}
	}

	// Print version information
	fmt.Printf("git-tag-similarity version %s\n", version)
	if modified {
		fmt.Printf("  Commit: %s (modified)\n", commit)
	} else {
		fmt.Printf("  Commit: %s\n", commit)
	}
	fmt.Printf("  Commit time: %s\n", commitTime)
	fmt.Printf("  Go version: %s\n", runtime.Version())
	fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
