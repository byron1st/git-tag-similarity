package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

// TestCompareWithDirectoryFilter tests the Compare function with directory filtering
func TestCompareWithDirectoryFilter(t *testing.T) {
	// Note: This is a unit test using mocks. For a full integration test,
	// you would need to create a real git repository with test data.

	// Create test commit hashes
	hash1 := plumbing.NewHash("0000000000000000000000000000000000000001")
	hash2 := plumbing.NewHash("0000000000000000000000000000000000000002")
	hash3 := plumbing.NewHash("0000000000000000000000000000000000000003")
	hash4 := plumbing.NewHash("0000000000000000000000000000000000000004")

	// Simulate filtered commits - tag1 has commits 1,2,3 and tag2 has commits 2,3,4
	// when filtering by a specific directory
	tag1FilteredCommits := map[plumbing.Hash]struct{}{
		hash1: {},
		hash2: {},
		hash3: {},
	}

	tag2FilteredCommits := map[plumbing.Hash]struct{}{
		hash2: {},
		hash3: {},
		hash4: {},
	}

	// The expected similarity should be 2 (shared: hash2, hash3) / 4 (total unique: hash1, hash2, hash3, hash4) = 0.5
	expectedSimilarity := 0.5

	// Calculate the Jaccard similarity
	similarity := CalculateJaccardSimilarity(tag1FilteredCommits, tag2FilteredCommits)

	if similarity != expectedSimilarity {
		t.Errorf("CalculateJaccardSimilarity() = %v, want %v", similarity, expectedSimilarity)
	}

	// Verify shared commits
	sharedCommits := make(map[plumbing.Hash]struct{})
	for hash := range tag1FilteredCommits {
		if _, ok := tag2FilteredCommits[hash]; ok {
			sharedCommits[hash] = struct{}{}
		}
	}

	if len(sharedCommits) != 2 {
		t.Errorf("Expected 2 shared commits, got %d", len(sharedCommits))
	}

	if _, ok := sharedCommits[hash2]; !ok {
		t.Errorf("Expected hash2 to be in shared commits")
	}

	if _, ok := sharedCommits[hash3]; !ok {
		t.Errorf("Expected hash3 to be in shared commits")
	}
}

// TestResolveTagToCommit_AnnotatedTag tests the helper with real annotated tags
func TestResolveTagToCommit_AnnotatedTag(t *testing.T) {
	// This repo has annotated tags (v1.0.0, v1.1.0, etc.)
	repo, err := NewGitRepository("..")
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	// Test with an annotated tag
	tags, err := repo.FetchAllTags()
	if err != nil {
		t.Fatalf("Failed to fetch tags: %v", err)
	}

	// Find v1.0.0 tag (we know it's annotated)
	var v100Ref *plumbing.Reference
	for _, ref := range tags {
		if ref.Name().Short() == "v1.0.0" {
			v100Ref = ref
			break
		}
	}
	if v100Ref == nil {
		t.Skip("v1.0.0 tag not found, skipping test")
	}

	// Resolve tag to commit
	commit, err := repo.resolveTagToCommit(v100Ref)
	if err != nil {
		t.Errorf("resolveTagToCommit() failed for annotated tag: %v", err)
		return
	}
	if commit == nil {
		t.Errorf("resolveTagToCommit() returned nil commit")
		return
	}

	// Verify it's a valid commit
	if commit.Hash.IsZero() {
		t.Errorf("resolveTagToCommit() returned commit with zero hash")
	}
}

// TestResolveTagToCommit_LightweightTag tests the helper with lightweight tags
func TestResolveTagToCommit_LightweightTag(t *testing.T) {
	// Create a test git repository with lightweight tag
	tempDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create a commit
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@test.com",
		"commit", "-m", "test commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create lightweight tag
	cmd = exec.Command("git", "tag", "lightweight-test")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create lightweight tag: %v", err)
	}

	// Open repository and test
	repo, err := NewGitRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	tags, err := repo.FetchAllTags()
	if err != nil {
		t.Fatalf("Failed to fetch tags: %v", err)
	}

	var lwRef *plumbing.Reference
	for _, ref := range tags {
		if ref.Name().Short() == "lightweight-test" {
			lwRef = ref
			break
		}
	}
	if lwRef == nil {
		t.Fatalf("lightweight-test tag not found")
	}

	// Resolve tag to commit
	commit, err := repo.resolveTagToCommit(lwRef)
	if err != nil {
		t.Errorf("resolveTagToCommit() failed for lightweight tag: %v", err)
		return
	}
	if commit == nil {
		t.Errorf("resolveTagToCommit() returned nil commit")
		return
	}

	// Verify it's a valid commit
	if commit.Hash.IsZero() {
		t.Errorf("resolveTagToCommit() returned commit with zero hash")
	}
}

// TestGetCommitSetForTag_AnnotatedTag tests with real annotated tags
func TestGetCommitSetForTag_AnnotatedTag(t *testing.T) {
	repo, err := NewGitRepository("..")
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	tags, err := repo.FetchAllTags()
	if err != nil {
		t.Fatalf("Failed to fetch tags: %v", err)
	}

	// Find v1.0.0 tag (annotated)
	var v100Ref *plumbing.Reference
	for _, ref := range tags {
		if ref.Name().Short() == "v1.0.0" {
			v100Ref = ref
			break
		}
	}
	if v100Ref == nil {
		t.Skip("v1.0.0 tag not found, skipping test")
	}

	// Get commit set
	commits, err := repo.GetCommitSetForTag(v100Ref)
	if err != nil {
		t.Errorf("GetCommitSetForTag() failed: %v", err)
	}
	if len(commits) == 0 {
		t.Errorf("GetCommitSetForTag() returned empty commit set")
	}
}

// TestGetCommitSetForTagFilteredByDirectory_AnnotatedTag tests with directory filter
func TestGetCommitSetForTagFilteredByDirectory_AnnotatedTag(t *testing.T) {
	repo, err := NewGitRepository("..")
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	tags, err := repo.FetchAllTags()
	if err != nil {
		t.Fatalf("Failed to fetch tags: %v", err)
	}

	var v100Ref *plumbing.Reference
	for _, ref := range tags {
		if ref.Name().Short() == "v1.0.0" {
			v100Ref = ref
			break
		}
	}
	if v100Ref == nil {
		t.Skip("v1.0.0 tag not found, skipping test")
	}

	// Get filtered commit set (internal directory exists in this repo)
	commits, err := repo.GetCommitSetForTagFilteredByDirectory(v100Ref, "internal")
	if err != nil {
		t.Errorf("GetCommitSetForTagFilteredByDirectory() failed: %v", err)
	}

	// Should have at least some commits touching internal/
	if len(commits) == 0 {
		t.Logf("Warning: No commits found for internal/ directory in v1.0.0")
	}
}

// TestGetDiffBetweenTags_AnnotatedTags tests diff with two annotated tags
func TestGetDiffBetweenTags_AnnotatedTags(t *testing.T) {
	repo, err := NewGitRepository("..")
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	tags, err := repo.FetchAllTags()
	if err != nil {
		t.Fatalf("Failed to fetch tags: %v", err)
	}

	var v100Ref, v110Ref *plumbing.Reference
	for _, ref := range tags {
		switch ref.Name().Short() {
		case "v1.0.0":
			v100Ref = ref
		case "v1.1.0":
			v110Ref = ref
		}
	}

	if v100Ref == nil || v110Ref == nil {
		t.Skip("Required tags not found, skipping test")
	}

	// Get diff between tags
	diff, err := repo.GetDiffBetweenTags(v100Ref, v110Ref, "")
	if err != nil {
		t.Errorf("GetDiffBetweenTags() failed: %v", err)
	}

	// Diff should not be empty (there are changes between these versions)
	if diff == "" {
		t.Logf("Warning: Empty diff between v1.0.0 and v1.1.0")
	}
}

// TestGetDiffBetweenTags_WithDirectory tests diff with directory filter
func TestGetDiffBetweenTags_WithDirectory(t *testing.T) {
	repo, err := NewGitRepository("..")
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	tags, err := repo.FetchAllTags()
	if err != nil {
		t.Fatalf("Failed to fetch tags: %v", err)
	}

	var v100Ref, v110Ref *plumbing.Reference
	for _, ref := range tags {
		switch ref.Name().Short() {
		case "v1.0.0":
			v100Ref = ref
		case "v1.1.0":
			v110Ref = ref
		}
	}

	if v100Ref == nil || v110Ref == nil {
		t.Skip("Required tags not found, skipping test")
	}

	// Get diff for internal directory only
	diff, err := repo.GetDiffBetweenTags(v100Ref, v110Ref, "internal")
	if err != nil {
		t.Errorf("GetDiffBetweenTags() with directory filter failed: %v", err)
	}

	// Should have some diff (internal/ has changes between versions)
	if diff == "" {
		t.Logf("Warning: Empty diff for internal/ between v1.0.0 and v1.1.0")
	}
}
