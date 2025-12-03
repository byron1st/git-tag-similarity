//go:generate mockgen -source=repository.go -destination=../mocks/repository_mock.go -package=mocks
package internal

import (
	"bufio"
	"errors"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	ErrOpenRepository  = errors.New("failed to open repository")
	ErrFetchTags       = errors.New("failed to fetch tags")
	ErrGetCommit       = errors.New("failed to get commit")
	ErrDereferenceTag  = errors.New("failed to dereference tag")
	ErrTraverseCommits = errors.New("failed to traverse commits")
)

// Repository is an interface that abstracts Git operations for testability
type Repository interface {
	FetchAllTags() ([]*plumbing.Reference, error)
	GetCommitSetForTag(ref *plumbing.Reference) (map[plumbing.Hash]struct{}, error)
	GetCommitSetForTagFilteredByDirectory(ref *plumbing.Reference, directory string) (map[plumbing.Hash]struct{}, error)
	GetCommitObject(hash plumbing.Hash) (*object.Commit, error)
	GetDiffBetweenTags(tag1 *plumbing.Reference, tag2 *plumbing.Reference, directory string) (string, error)
}

// GitRepository is a concrete implementation of Repository using go-git
type GitRepository struct {
	path string
	repo *git.Repository
}

// NewGitRepository creates a new GitRepository instance
func NewGitRepository(path string) (*GitRepository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, errors.Join(ErrOpenRepository, err)
	}
	return &GitRepository{
		path: path,
		repo: repo,
	}, nil
}

// resolveTagToCommit resolves a tag reference to its commit object.
// Handles both annotated tags (tag objects) and lightweight tags (direct commit refs).
func (gr *GitRepository) resolveTagToCommit(ref *plumbing.Reference) (*object.Commit, error) {
	// Try to get tag object first (annotated tag)
	tagObj, err := gr.repo.TagObject(ref.Hash())
	if err == nil {
		// Annotated tag - dereference to commit
		commit, err := tagObj.Commit()
		if err != nil {
			return nil, errors.Join(ErrDereferenceTag, err)
		}
		return commit, nil
	}

	// Not a tag object - try commit directly (lightweight tag)
	commit, err := gr.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, errors.Join(ErrDereferenceTag, err)
	}
	return commit, nil
}

// FetchAllTags retrieves all tag references from the repository
func (gr *GitRepository) FetchAllTags() ([]*plumbing.Reference, error) {
	tagRefs, err := gr.repo.Tags()
	if err != nil {
		return nil, errors.Join(ErrFetchTags, err)
	}

	var refs []*plumbing.Reference
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		refs = append(refs, ref)
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrFetchTags, err)
	}

	return refs, nil
}

// GetCommitSetForTag traverses the history of a tag and returns all parent commit hashes.
// Handles both annotated tags (tag objects) and lightweight tags (direct commit refs).
func (gr *GitRepository) GetCommitSetForTag(ref *plumbing.Reference) (map[plumbing.Hash]struct{}, error) {
	commitSet := make(map[plumbing.Hash]struct{})

	// Resolve tag to commit (handles both annotated and lightweight tags)
	commit, err := gr.resolveTagToCommit(ref)
	if err != nil {
		return nil, err // Error already wrapped by helper
	}

	// Traverse all parent commits (similar to git log)
	cIter, err := gr.repo.Log(&git.LogOptions{From: commit.Hash})
	if err != nil {
		return nil, errors.Join(ErrTraverseCommits, err)
	}
	defer func() { cIter.Close() }()

	// Add all parent commits to the set
	err = cIter.ForEach(func(c *object.Commit) error {
		commitSet[c.Hash] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrTraverseCommits, err)
	}

	return commitSet, nil
}

// GetCommitSetForTagFilteredByDirectory traverses the history of a tag and returns commits
// that touch files in the specified directory.
// Handles both annotated tags (tag objects) and lightweight tags (direct commit refs).
// Uses native git log command for performance (go-git's PathFilter is extremely slow).
func (gr *GitRepository) GetCommitSetForTagFilteredByDirectory(ref *plumbing.Reference, directory string) (map[plumbing.Hash]struct{}, error) {
	commitSet := make(map[plumbing.Hash]struct{})

	// Resolve tag to commit (handles both annotated and lightweight tags)
	commit, err := gr.resolveTagToCommit(ref)
	if err != nil {
		return nil, err // Error already wrapped by helper
	}

	// Use native git log with path filtering (orders of magnitude faster than go-git's PathFilter)
	// Command: git log <commit> --format=%H -- <directory>
	cmd := exec.Command("git", "log", commit.Hash.String(), "--format=%H", "--", directory)
	cmd.Dir = gr.path

	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Join(ErrTraverseCommits, err)
	}

	// Parse commit hashes from output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		hash := plumbing.NewHash(line)
		commitSet[hash] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Join(ErrTraverseCommits, err)
	}

	return commitSet, nil
}

// GetCommitObject retrieves a commit object by its hash
func (gr *GitRepository) GetCommitObject(hash plumbing.Hash) (*object.Commit, error) {
	commit, err := gr.repo.CommitObject(hash)
	if err != nil {
		return nil, errors.Join(ErrGetCommit, err)
	}
	return commit, nil
}

// GetDiffBetweenTags returns the diff between two tags.
// Handles both annotated tags (tag objects) and lightweight tags (direct commit refs).
// If directory is specified, only shows diff for files in that directory.
func (gr *GitRepository) GetDiffBetweenTags(tag1 *plumbing.Reference, tag2 *plumbing.Reference, directory string) (string, error) {
	// Resolve tags to commits (handles both annotated and lightweight tags)
	commit1, err := gr.resolveTagToCommit(tag1)
	if err != nil {
		return "", err // Error already wrapped by helper
	}

	commit2, err := gr.resolveTagToCommit(tag2)
	if err != nil {
		return "", err // Error already wrapped by helper
	}

	// Use git diff command with stat for summary
	// Command: git diff <commit1> <commit2> [-- <directory>]
	args := []string{"diff", "--stat", "--stat-width=120", commit1.Hash.String(), commit2.Hash.String()}
	if directory != "" {
		args = append(args, "--", directory)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = gr.path

	output, err := cmd.Output()
	if err != nil {
		return "", errors.Join(ErrTraverseCommits, err)
	}

	return string(output), nil
}
