package vrh

// maxScore returns the index and score of the highest-scoring node.
//
// The function assumes nodes and scores have the same length and that
// the slice is non-empty.
func maxScore(nodes, scores []uint64) (owner uint64, best uint64) {
	owner = nodes[0]
	best = scores[0]

	for i := 1; i < len(nodes); i++ {
		if scores[i] > best {
			best = scores[i]
			owner = nodes[i]
		}
	}

	return owner, best
}
