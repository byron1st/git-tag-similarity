package internal

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

var (
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrValidationFailed     = errors.New("validation failed")
	ErrGetTagReference      = errors.New("failed to get tag reference")
	ErrGetCommits           = errors.New("failed to get commits")
)

func PrintCompareResult(result CompareResult) {
	fmt.Printf("Comparing tags: %s vs %s\n", result.Config.Tag1Name, result.Config.Tag2Name)
	fmt.Printf("Similarity: %.2f%%\n", result.Similarity*100.0)
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total commits in [%s]: %d\n", result.Config.Tag1Name, len(result.OnlyInTag1))
	fmt.Printf("  Total commits in [%s]: %d\n", result.Config.Tag2Name, len(result.OnlyInTag2))
	fmt.Printf("  Shared commits: %d\n", len(result.SharedCommits))
	fmt.Printf("  Unique to [%s]: %d\n", result.Config.Tag1Name, len(result.OnlyInTag1))
	fmt.Printf("  Unique to [%s]: %d\n", result.Config.Tag2Name, len(result.OnlyInTag2))

	// Print detailed commit lists if verbose flag is set
	if result.Config.Verbose {
		printDiffCommits(result.Repo, result.Config.Tag1Name, result.OnlyInTag1)
		printDiffCommits(result.Repo, result.Config.Tag2Name, result.OnlyInTag2)
	}
}

func Compare(config CompareConfig) (CompareResult, error) {
	result := CompareResult{Config: config}

	// Validate basic configuration
	if err := config.Validate(); err != nil {
		return result, errors.Join(ErrInvalidConfiguration, err)
	}

	// 2. Open repository
	repo, err := NewGitRepository(config.RepoPath)
	if err != nil {
		return result, errors.Join(ErrOpenRepository, err)
	}

	// 3. Validate that both tags exist in the repository
	if err := config.ValidateWithRepository(repo); err != nil {
		return result, errors.Join(ErrValidationFailed, err)
	}

	// 4. Get tag references for both tags
	tag1Ref, err := config.GetTagReference(repo, config.Tag1Name)
	if err != nil {
		return result, errors.Join(ErrGetTagReference, err)
	}

	tag2Ref, err := config.GetTagReference(repo, config.Tag2Name)
	if err != nil {
		return result, errors.Join(ErrGetTagReference, err)
	}

	// 5. Get commit sets for both tags
	tag1Commits, err := repo.GetCommitSetForTag(tag1Ref)
	if err != nil {
		return result, errors.Join(ErrGetCommits, err)
	}

	tag2Commits, err := repo.GetCommitSetForTag(tag2Ref)
	if err != nil {
		return result, errors.Join(ErrGetCommits, err)
	}

	// 6. Calculate similarity
	result.Similarity = CalculateJaccardSimilarity(tag1Commits, tag2Commits)

	// 7. Calculate shared and unique commits
	result.SharedCommits = make(map[plumbing.Hash]struct{})
	result.OnlyInTag1 = make(map[plumbing.Hash]struct{})
	result.OnlyInTag2 = make(map[plumbing.Hash]struct{})

	for hash := range tag1Commits {
		if _, ok := tag2Commits[hash]; ok {
			result.SharedCommits[hash] = struct{}{}
		} else {
			result.OnlyInTag1[hash] = struct{}{}
		}
	}

	for hash := range tag2Commits {
		if _, ok := tag1Commits[hash]; !ok {
			result.OnlyInTag2[hash] = struct{}{}
		}
	}

	return result, nil
}

// printDiffCommits prints the commit messages for commits unique to a tag
func printDiffCommits(repo Repository, tagName string, diffSet map[plumbing.Hash]struct{}) {
	if len(diffSet) == 0 {
		return
	}

	fmt.Printf("\nCommits only in [%s] (%d):\n", tagName, len(diffSet))
	for hash := range diffSet {
		commit, err := repo.GetCommitObject(hash)
		if err != nil {
			fmt.Printf("  - %s (failed to get message: %v)\n", hash.String(), err)
			continue
		}
		// Get only the first line of the message
		message := strings.Split(commit.Message, "\n")[0]
		fmt.Printf("  - %s : %s\n", hash.String()[:7], message)
	}
}

// CompareConfig holds the application configuration from command-line arguments
type CompareConfig struct {
	Command  Command
	RepoPath string
	Tag1Name string
	Tag2Name string
	Verbose  bool
}

// NewCompareConfig parses the compare command flags
func NewCompareConfig(args []string) (CompareConfig, error) {
	config := CompareConfig{Command: CompareCommand}

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
		return config, err
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *CompareConfig) Validate() error {
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
func (c *CompareConfig) ValidateWithRepository(repo Repository) error {
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
func (c *CompareConfig) GetTagReference(repo Repository, tagName string) (*plumbing.Reference, error) {
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

type CompareResult struct {
	Repo          Repository
	Config        CompareConfig
	Similarity    float64
	SharedCommits map[plumbing.Hash]struct{}
	OnlyInTag1    map[plumbing.Hash]struct{}
	OnlyInTag2    map[plumbing.Hash]struct{}
}
