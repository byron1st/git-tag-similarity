//go:generate mockgen -source=repository.go -destination=../mocks/repository_mock.go -package=mocks
package internal

import (
	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	ErrOpenRepository  = errors.New("failed to open repository")
	ErrFetchTags       = errors.New("failed to fetch tags")
	ErrGetCommit       = errors.New("failed to get commit")
	ErrTraverseCommits = errors.New("failed to traverse commits")
)

// Repository is an interface that abstracts Git operations for testability
type Repository interface {
	FetchAllTags() ([]*plumbing.Reference, error)
	GetCommitSetForTag(ref *plumbing.Reference) (map[plumbing.Hash]struct{}, error)
	GetCommitObject(hash plumbing.Hash) (*object.Commit, error)
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

// GetCommitSetForTag traverses the history of a tag and returns all parent commit hashes
func (gr *GitRepository) GetCommitSetForTag(ref *plumbing.Reference) (map[plumbing.Hash]struct{}, error) {
	commitSet := make(map[plumbing.Hash]struct{})

	// Get the commit object that the tag points to (handles annotated tags automatically)
	commit, err := gr.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, errors.Join(ErrGetCommit, err)
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

// GetCommitObject retrieves a commit object by its hash
func (gr *GitRepository) GetCommitObject(hash plumbing.Hash) (*object.Commit, error) {
	commit, err := gr.repo.CommitObject(hash)
	if err != nil {
		return nil, errors.Join(ErrGetCommit, err)
	}
	return commit, nil
}
