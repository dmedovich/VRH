package verify

// CheckDistinctNodes verifies that the node set contains no duplicates.
func CheckDistinctNodes(nodes []uint64) bool {
	seen := make(map[uint64]struct{}, len(nodes))

	for _, node := range nodes {
		if _, exists := seen[node]; exists {
			return false
		}

		seen[node] = struct{}{}
	}

	return true
}

// CheckOwners verifies a collection of key/owner pairs.
func CheckOwners(
	keys []uint64,
	nodes []uint64,
	owners []uint64,
	score func(key, node uint64) uint64,
) bool {
	if len(keys) != len(owners) || len(nodes) == 0 {
		return false
	}

	for i, key := range keys {
		if !CheckOwner(key, nodes, owners[i], score) {
			return false
		}
	}

	return true
}
