package metrics

// Balance calculates how evenly keys are distributed across nodes.
//
// nodes is the complete node set. owners contains the owner node for every
// key. Nodes with zero assigned keys are included in the calculation.
//
// The returned value is the coefficient of variation:
//
// \tstandard deviation / mean
//
// 0 means perfectly balanced. Lower values are better.
func Balance(nodes, owners []uint64) float64 {
	if len(nodes) == 0 {
		return 0
	}

	counts := BalanceCounts(nodes, owners)

	mean := float64(len(owners)) / float64(len(nodes))
	if mean == 0 {
		return 0
	}

	var variance float64

	for _, count := range counts {
		delta := float64(count) - mean
		variance += delta * delta
	}

	variance /= float64(len(nodes))

	return sqrt(variance) / mean
}

// BalanceCounts returns the number of keys assigned to every node.
//
// Every node from nodes is present in the result, including nodes with zero
// assignments. An owner ID outside nodes is rejected because it would make
// the distribution metric ill-defined for the supplied node set.
func BalanceCounts(nodes, owners []uint64) map[uint64]int {
	counts := make(map[uint64]int, len(nodes))

	for _, node := range nodes {
		counts[node] = 0
	}

	for _, owner := range owners {
		if _, ok := counts[owner]; !ok {
			panic("metrics: owner is not present in node set")
		}

		counts[owner]++
	}

	return counts
}

// sqrt is kept local so this package has a very small dependency surface.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}

	z := x

	for i := 0; i < 32; i++ {
		z = 0.5 * (z + x/z)
	}

	return z
}
