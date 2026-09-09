package verify

import "github.com/dmedovich/VRH/algorithms"

// CheckVRHPermutation checks that, for a fixed key, VRH produces
// distinct scores for all distinct nodes.
//
// This property is important because VRH uses a key-dependent
// permutation of the uint64 score space.
func CheckVRHPermutation(
	key uint64,
	nodes []uint64,
) bool {
	if len(nodes) < 2 {
		return true
	}

	seen := make(map[uint64]struct{}, len(nodes))

	for _, node := range nodes {
		score := algorithms.VRHScore(key, node)

		if _, exists := seen[score]; exists {
			return false
		}

		seen[score] = struct{}{}
	}

	return true
}
