package verify

import "github.com/dmedovich/VRH/algorithms"

// CheckOwner verifies that the supplied owner is the actual
// highest-scoring node according to the given scoring function.
func CheckOwner(
	key uint64,
	nodes []uint64,
	owner uint64,
	score func(key, node uint64) uint64,
) bool {
	if len(nodes) == 0 {
		return false
	}

	bestNode := nodes[0]
	bestScore := score(key, bestNode)

	for _, node := range nodes[1:] {
		s := score(key, node)

		if s > bestScore {
			bestScore = s
			bestNode = node
		}
	}

	return owner == bestNode
}

// CheckVRHOwner verifies an VRH owner.
func CheckVRHOwner(
	key uint64,
	nodes []uint64,
	owner uint64,
) bool {
	return CheckOwner(
		key,
		nodes,
		owner,
		algorithms.VRHScore,
	)
}
