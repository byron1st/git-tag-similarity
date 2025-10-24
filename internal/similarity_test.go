package internal

import (
	"math"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

// TestCalculateJaccardSimilarity tests the Jaccard similarity calculation function
func TestCalculateJaccardSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		setA     map[plumbing.Hash]struct{}
		setB     map[plumbing.Hash]struct{}
		expected float64
	}{
		{
			name:     "Both empty sets",
			setA:     map[plumbing.Hash]struct{}{},
			setB:     map[plumbing.Hash]struct{}{},
			expected: 1.0,
		},
		{
			name: "Identical sets",
			setA: map[plumbing.Hash]struct{}{
				hashFromString("commit1"): {},
				hashFromString("commit2"): {},
				hashFromString("commit3"): {},
			},
			setB: map[plumbing.Hash]struct{}{
				hashFromString("commit1"): {},
				hashFromString("commit2"): {},
				hashFromString("commit3"): {},
			},
			expected: 1.0,
		},
		{
			name: "Completely disjoint sets",
			setA: map[plumbing.Hash]struct{}{
				hashFromString("commit1"): {},
				hashFromString("commit2"): {},
			},
			setB: map[plumbing.Hash]struct{}{
				hashFromString("commit3"): {},
				hashFromString("commit4"): {},
			},
			expected: 0.0,
		},
		{
			name: "Partially overlapping sets (50% overlap)",
			setA: map[plumbing.Hash]struct{}{
				hashFromString("commit1"): {},
				hashFromString("commit2"): {},
			},
			setB: map[plumbing.Hash]struct{}{
				hashFromString("commit2"): {},
				hashFromString("commit3"): {},
			},
			expected: 1.0 / 3.0, // 1 common / 3 total
		},
		{
			name: "One empty, one non-empty",
			setA: map[plumbing.Hash]struct{}{
				hashFromString("commit1"): {},
			},
			setB:     map[plumbing.Hash]struct{}{},
			expected: 0.0,
		},
		{
			name: "Empty first, non-empty second",
			setA: map[plumbing.Hash]struct{}{},
			setB: map[plumbing.Hash]struct{}{
				hashFromString("commit1"): {},
			},
			expected: 0.0,
		},
		{
			name: "Subset relationship (A is subset of B)",
			setA: map[plumbing.Hash]struct{}{
				hashFromString("commit1"): {},
				hashFromString("commit2"): {},
			},
			setB: map[plumbing.Hash]struct{}{
				hashFromString("commit1"): {},
				hashFromString("commit2"): {},
				hashFromString("commit3"): {},
				hashFromString("commit4"): {},
			},
			expected: 2.0 / 4.0, // 2 common / 4 total
		},
		{
			name: "One commit overlap in larger sets",
			setA: map[plumbing.Hash]struct{}{
				hashFromString("commit1"): {},
				hashFromString("commit2"): {},
				hashFromString("commit3"): {},
			},
			setB: map[plumbing.Hash]struct{}{
				hashFromString("commit3"): {},
				hashFromString("commit4"): {},
				hashFromString("commit5"): {},
			},
			expected: 1.0 / 5.0, // 1 common / 5 total
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateJaccardSimilarity(tt.setA, tt.setB)
			if math.Abs(result-tt.expected) > 0.0001 { // To handle the inherent imprecision of floating-point arithmetic
				t.Errorf("calculateJaccardSimilarity() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateJaccardSimilaritySymmetry ensures that similarity is symmetric
func TestCalculateJaccardSimilaritySymmetry(t *testing.T) {
	setA := map[plumbing.Hash]struct{}{
		hashFromString("commit1"): {},
		hashFromString("commit2"): {},
	}
	setB := map[plumbing.Hash]struct{}{
		hashFromString("commit2"): {},
		hashFromString("commit3"): {},
	}

	resultAB := CalculateJaccardSimilarity(setA, setB)
	resultBA := CalculateJaccardSimilarity(setB, setA)

	if math.Abs(resultAB-resultBA) > 0.0001 {
		t.Errorf("Jaccard similarity is not symmetric: AB=%v, BA=%v", resultAB, resultBA)
	}
}

// Helper function to create a commit hash from a string
func hashFromString(s string) plumbing.Hash {
	var h plumbing.Hash
	copy(h[:], s)
	return h
}
