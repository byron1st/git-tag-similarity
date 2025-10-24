package internal

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing"
)

var (
	ErrMissingRepo    = errors.New("repository path is required")
	ErrMissingTag1    = errors.New("first tag name is required")
	ErrMissingTag2    = errors.New("second tag name is required")
	ErrInvalidRepo    = errors.New("invalid repository path")
	ErrTag1NotFound   = errors.New("first tag not found in repository")
	ErrTag2NotFound   = errors.New("second tag not found in repository")
	ErrInvalidCommand = errors.New("invalid command")
	ErrNoCommand      = errors.New("no command specified")
)

// Command represents the CLI command type
type Command string

const (
	CompareCommand Command = "compare"
	HelpCommand    Command = "help"
	VersionCommand Command = "version"
)

// Config holds the application configuration from command-line arguments
type Config struct {
	Command  Command
	RepoPath string
	Tag1Name string
	Tag2Name string
	Verbose  bool
}

// ParseCommand parses command-line arguments and returns the configuration
func ParseCommand(args []string) (*Config, error) {
	if len(args) < 1 {
		PrintUsage()
		return nil, ErrNoCommand
	}

	command := args[0]
	config := &Config{}

	switch command {
	case "compare":
		return parseCompareCommand(args[1:])
	case "help":
		PrintUsage()
		os.Exit(0)
	case "version":
		PrintVersion()
		os.Exit(0)
	default:
		PrintUsage()
		return nil, errors.Join(ErrInvalidCommand, fmt.Errorf("unknown command: %s", command))
	}

	return config, nil
}

// parseCompareCommand parses the compare command flags
func parseCompareCommand(args []string) (*Config, error) {
	config := &Config{Command: CompareCommand}

	compareCmd := flag.NewFlagSet("compare", flag.ExitOnError)
	compareCmd.StringVar(&config.RepoPath, "repo", "", "Path to the Git repository")
	compareCmd.StringVar(&config.Tag1Name, "tag1", "", "First tag name to compare")
	compareCmd.StringVar(&config.Tag2Name, "tag2", "", "Second tag name to compare")
	compareCmd.BoolVar(&config.Verbose, "v", false, "Verbose output (show list of different commits)")

	compareCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: git-tag-similarity compare [options]\n\n")
		fmt.Fprintf(os.Stderr, "Compare two Git tags and calculate their similarity.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		compareCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0\n")
		fmt.Fprintf(os.Stderr, "  git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0 -v\n")
	}

	if err := compareCmd.Parse(args); err != nil {
		return nil, err
	}

	return config, nil
}

// PrintUsage prints the main usage information
func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage: git-tag-similarity <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "A tool to compare two Git tags and calculate their similarity based on commit history.\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  compare    Compare two Git tags\n")
	fmt.Fprintf(os.Stderr, "  help       Show this help message\n")
	fmt.Fprintf(os.Stderr, "  version    Show version information\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity compare -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity help\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity version\n")
	fmt.Fprintf(os.Stderr, "\nFor more information on a command, use:\n")
	fmt.Fprintf(os.Stderr, "  git-tag-similarity <command> -h\n")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check required flags
	if c.RepoPath == "" {
		return ErrMissingRepo
	}

	if c.Tag1Name == "" {
		return ErrMissingTag1
	}

	if c.Tag2Name == "" {
		return ErrMissingTag2
	}

	// Check if repository path exists and is accessible
	if _, err := os.Stat(c.RepoPath); os.IsNotExist(err) {
		return errors.Join(ErrInvalidRepo, fmt.Errorf("path does not exist: %s", c.RepoPath))
	}

	return nil
}

// ValidateWithRepository checks if both tags exist in the repository
func (c *Config) ValidateWithRepository(repo Repository) error {
	// First validate basic configuration
	if err := c.Validate(); err != nil {
		return err
	}

	// Fetch all tags to check if the specified tags exist
	tagRefs, err := repo.FetchAllTags()
	if err != nil {
		return err
	}

	// Build a map of tag names for quick lookup
	tagMap := make(map[string]bool)
	for _, ref := range tagRefs {
		tagMap[ref.Name().Short()] = true
	}

	// Check if both tags exist
	tag1Found := tagMap[c.Tag1Name]
	tag2Found := tagMap[c.Tag2Name]

	if !tag1Found {
		return errors.Join(ErrTag1NotFound, fmt.Errorf("tag '%s' not found in repository", c.Tag1Name))
	}

	if !tag2Found {
		return errors.Join(ErrTag2NotFound, fmt.Errorf("tag '%s' not found in repository", c.Tag2Name))
	}

	return nil
}

// GetTagReference finds and returns the reference for a specific tag name
func (c *Config) GetTagReference(repo Repository, tagName string) (*plumbing.Reference, error) {
	tagRefs, err := repo.FetchAllTags()
	if err != nil {
		return nil, err
	}

	for _, ref := range tagRefs {
		if ref.Name().Short() == tagName {
			return ref, nil
		}
	}

	return nil, fmt.Errorf("tag '%s' not found", tagName)
}
