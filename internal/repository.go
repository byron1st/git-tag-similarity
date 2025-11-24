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
	GetCommitSetForTagFilteredByDirectory(ref *plumbing.Reference, directory string) (map[plumbing.Hash]struct{}, error)
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

// GetCommitSetForTagFilteredByDirectory traverses the history of a tag and returns commits
// that touch files in the specified directory
func (gr *GitRepository) GetCommitSetForTagFilteredByDirectory(ref *plumbing.Reference, directory string) (map[plumbing.Hash]struct{}, error) {
	commitSet := make(map[plumbing.Hash]struct{})

	// Get the commit object that the tag points to (handles annotated tags automatically)
	commit, err := gr.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, errors.Join(ErrGetCommit, err)
	}

	// Traverse all parent commits
	cIter, err := gr.repo.Log(&git.LogOptions{From: commit.Hash})
	if err != nil {
		return nil, errors.Join(ErrTraverseCommits, err)
	}
	defer func() { cIter.Close() }()

	// Check each commit to see if it touches the specified directory
	err = cIter.ForEach(func(c *object.Commit) error {
		touchesDirectory, err := gr.commitTouchesDirectory(c, directory)
		if err != nil {
			return err
		}
		if touchesDirectory {
			commitSet[c.Hash] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrTraverseCommits, err)
	}

	return commitSet, nil
}

// commitTouchesDirectory checks if a commit modifies any files in the specified directory
func (gr *GitRepository) commitTouchesDirectory(commit *object.Commit, directory string) (bool, error) {
	// Get the tree for this commit
	tree, err := commit.Tree()
	if err != nil {
		return false, err
	}

	// If the commit has no parents (initial commit), check if any files in the tree are in the directory
	if commit.NumParents() == 0 {
		return gr.treeContainsDirectory(tree, directory)
	}

	// Check each parent (handle merge commits)
	for i := range commit.NumParents() {
		parent, err := commit.Parent(i)
		if err != nil {
			return false, err
		}

		parentTree, err := parent.Tree()
		if err != nil {
			return false, err
		}

		// Get the changes between parent and current commit
		changes, err := parentTree.Diff(tree)
		if err != nil {
			return false, err
		}

		// Check if any changes touch the specified directory
		for _, change := range changes {
			// Check both From and To paths (for renames/moves)
			if gr.pathInDirectory(change.From.Name, directory) || gr.pathInDirectory(change.To.Name, directory) {
				return true, nil
			}
		}
	}

	return false, nil
}

// treeContainsDirectory checks if a tree contains any files in the specified directory
func (gr *GitRepository) treeContainsDirectory(tree *object.Tree, directory string) (bool, error) {
	found := false
	err := tree.Files().ForEach(func(f *object.File) error {
		if gr.pathInDirectory(f.Name, directory) {
			found = true
			return errors.New("found") // Stop iteration
		}
		return nil
	})
	if err != nil && err.Error() != "found" {
		return false, err
	}
	return found, nil
}

// pathInDirectory checks if a file path is within the specified directory
func (gr *GitRepository) pathInDirectory(path string, directory string) bool {
	if directory == "" {
		return true
	}
	// Normalize paths by removing trailing slashes
	dir := directory
	if len(dir) > 0 && dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}
	// Check if path starts with directory/
	if len(path) >= len(dir) {
		if path[:len(dir)] == dir {
			// Exact match or starts with directory/
			if len(path) == len(dir) || path[len(dir)] == '/' {
				return true
			}
		}
	}
	return false
}

// GetCommitObject retrieves a commit object by its hash
func (gr *GitRepository) GetCommitObject(hash plumbing.Hash) (*object.Commit, error) {
	commit, err := gr.repo.CommitObject(hash)
	if err != nil {
		return nil, errors.Join(ErrGetCommit, err)
	}
	return commit, nil
}
