package internal

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing"
)

var (
	ErrMissingRepo  = errors.New("repository path is required")
	ErrMissingTag1  = errors.New("first tag name is required")
	ErrMissingTag2  = errors.New("second tag name is required")
	ErrInvalidRepo  = errors.New("invalid repository path")
	ErrTag1NotFound = errors.New("first tag not found in repository")
	ErrTag2NotFound = errors.New("second tag not found in repository")
)

// Config holds the application configuration from command-line flags
type Config struct {
	RepoPath string
	Tag1Name string
	Tag2Name string
}

// ParseFlags parses command-line flags and returns the configuration
func ParseFlags(printVersion func()) (*Config, error) {
	config := &Config{}
	showVersion := flag.Bool("version", false, "Print version information and exit")

	flag.StringVar(&config.RepoPath, "repo", "", "Path to the Git repository")
	flag.StringVar(&config.Tag1Name, "tag1", "", "First tag name to compare")
	flag.StringVar(&config.Tag2Name, "tag2", "", "Second tag name to compare")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "This application compares two Git tags and calculates their similarity based on commit history.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -repo /path/to/repo -tag1 v1.0.0 -tag2 v2.0.0\n", os.Args[0])
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	return config, nil
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
