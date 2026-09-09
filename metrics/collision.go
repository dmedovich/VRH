package metrics

// CollisionCount counts duplicate scores.
//
// scores are expected to be generated for distinct nodes under
// the same key.
func CollisionCount(scores []uint64) int {
	seen := make(map[uint64]struct{}, len(scores))
	collisions := 0

	for _, score := range scores {
		if _, exists := seen[score]; exists {
			collisions++
			continue
		}

		seen[score] = struct{}{}
	}

	return collisions
}

// HasCollision reports whether scores contain at least one duplicate.
func HasCollision(scores []uint64) bool {
	return CollisionCount(scores) != 0
}
