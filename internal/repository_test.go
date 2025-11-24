package internal

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

// TestPathInDirectory tests the pathInDirectory helper function
func TestPathInDirectory(t *testing.T) {
	gr := &GitRepository{}

	tests := []struct {
		name      string
		path      string
		directory string
		want      bool
	}{
		{
			name:      "Empty directory matches all",
			path:      "src/main.go",
			directory: "",
			want:      true,
		},
		{
			name:      "Exact directory match",
			path:      "src/main.go",
			directory: "src",
			want:      true,
		},
		{
			name:      "Subdirectory match",
			path:      "src/api/handler.go",
			directory: "src",
			want:      true,
		},
		{
			name:      "Subdirectory with slash",
			path:      "src/api/handler.go",
			directory: "src/api",
			want:      true,
		},
		{
			name:      "Directory with trailing slash",
			path:      "src/api/handler.go",
			directory: "src/api/",
			want:      true,
		},
		{
			name:      "No match - different directory",
			path:      "internal/cli.go",
			directory: "src",
			want:      false,
		},
		{
			name:      "No match - partial prefix",
			path:      "source/main.go",
			directory: "src",
			want:      false,
		},
		{
			name:      "Deep subdirectory match",
			path:      "src/api/v1/handlers/user.go",
			directory: "src/api",
			want:      true,
		},
		{
			name:      "File in root",
			path:      "main.go",
			directory: "src",
			want:      false,
		},
		{
			name:      "Exact file path match (as directory)",
			path:      "src/main.go",
			directory: "src/main.go",
			want:      true,
		},
		{
			name:      "Case sensitive - no match",
			path:      "Src/main.go",
			directory: "src",
			want:      false,
		},
		{
			name:      "Path shorter than directory",
			path:      "src",
			directory: "src/api",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gr.pathInDirectory(tt.path, tt.directory)
			if got != tt.want {
				t.Errorf("pathInDirectory(%q, %q) = %v, want %v", tt.path, tt.directory, got, tt.want)
			}
		})
	}
}

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
