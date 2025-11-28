package internal

import (
	"fmt"
	"os"
)

// PrintUsage prints the main usage information
func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage: git-tag-similarity <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "A tool to compare two Git tags and calculate their similarity based on commit history.\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  compare    Compare two Git tags\n")
	fmt.Fprintf(os.Stderr, "  config     Configure AI settings for report generation\n")
	fmt.Fprintf(os.Stderr, "  help       Show this help message\n")
	fmt.Fprintf(os.Stderr, "  version    Show version information\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -r report.md\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity config -provider claude -api-key sk-ant-...\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity config -provider openai -api-key sk-...\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity help\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity version\n")
	fmt.Fprintf(os.Stderr, "\nFor more information on a command, use:\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity <command> -h\n")
}
