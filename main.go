package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/byron1st/git-tag-similarity/internal"
	"github.com/go-git/go-git/v5/plumbing"
)

func main() {
	// 1. Parse command-line flags
	config, err := internal.ParseFlags(PrintVersion)
	if err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	// Validate basic configuration
	if err := config.Validate(); err != nil {
		flag.Usage()
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "\nError: %v\n", err)
		log.Fatalf("Invalid configuration: %v", err)
	}

	// 2. Open repository
	repo, err := internal.NewGitRepository(config.RepoPath)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}

	// 3. Validate that both tags exist in the repository
	if err := config.ValidateWithRepository(repo); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// 4. Get tag references for both tags
	tag1Ref, err := config.GetTagReference(repo, config.Tag1Name)
	if err != nil {
		log.Fatalf("Failed to get reference for tag1: %v", err)
	}

	tag2Ref, err := config.GetTagReference(repo, config.Tag2Name)
	if err != nil {
		log.Fatalf("Failed to get reference for tag2: %v", err)
	}

	// 5. Get commit sets for both tags
	tag1Commits, err := repo.GetCommitSetForTag(tag1Ref)
	if err != nil {
		log.Fatalf("Failed to get commits for tag '%s': %v", config.Tag1Name, err)
	}

	tag2Commits, err := repo.GetCommitSetForTag(tag2Ref)
	if err != nil {
		log.Fatalf("Failed to get commits for tag '%s': %v", config.Tag2Name, err)
	}

	// 6. Calculate similarity
	similarity := internal.CalculateJaccardSimilarity(tag1Commits, tag2Commits)

	// 7. Calculate shared and unique commits
	sharedCommits := make(map[plumbing.Hash]struct{})
	onlyInTag1 := make(map[plumbing.Hash]struct{})
	onlyInTag2 := make(map[plumbing.Hash]struct{})

	for hash := range tag1Commits {
		if _, ok := tag2Commits[hash]; ok {
			sharedCommits[hash] = struct{}{}
		} else {
			onlyInTag1[hash] = struct{}{}
		}
	}

	for hash := range tag2Commits {
		if _, ok := tag1Commits[hash]; !ok {
			onlyInTag2[hash] = struct{}{}
		}
	}

	// 8. Print results
	fmt.Printf("Comparing tags: %s vs %s\n", config.Tag1Name, config.Tag2Name)
	fmt.Printf("Similarity: %.2f%%\n", similarity*100.0)
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total commits in [%s]: %d\n", config.Tag1Name, len(tag1Commits))
	fmt.Printf("  Total commits in [%s]: %d\n", config.Tag2Name, len(tag2Commits))
	fmt.Printf("  Shared commits: %d\n", len(sharedCommits))
	fmt.Printf("  Unique to [%s]: %d\n", config.Tag1Name, len(onlyInTag1))
	fmt.Printf("  Unique to [%s]: %d\n", config.Tag2Name, len(onlyInTag2))

	// Print unique commits for each tag
	printDiffCommits(repo, config.Tag1Name, onlyInTag1)
	printDiffCommits(repo, config.Tag2Name, onlyInTag2)
}

// printDiffCommits prints the commit messages for commits unique to a tag
func printDiffCommits(repo internal.Repository, tagName string, diffSet map[plumbing.Hash]struct{}) {
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
