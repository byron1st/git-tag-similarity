package internal

import "github.com/go-git/go-git/v5/plumbing"

// CalculateJaccardSimilarity computes the Jaccard similarity coefficient between two commit sets
// Returns a value between 0.0 and 1.0, where 1.0 means identical sets
func CalculateJaccardSimilarity(setA map[plumbing.Hash]struct{}, setB map[plumbing.Hash]struct{}) float64 {
	if len(setA) == 0 && len(setB) == 0 {
		return 1.0 // Both empty sets are considered identical
	}

	// Calculate union
	union := make(map[plumbing.Hash]struct{})
	for hash := range setA {
		union[hash] = struct{}{}
	}
	for hash := range setB {
		union[hash] = struct{}{}
	}

	if len(union) == 0 {
		return 0.0
	}

	// Calculate intersection
	intersection := make(map[plumbing.Hash]struct{})
	for hash := range setA {
		if _, ok := setB[hash]; ok {
			intersection[hash] = struct{}{}
		}
	}

	return float64(len(intersection)) / float64(len(union))
}
