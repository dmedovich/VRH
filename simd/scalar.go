package simd

import "github.com/dmedovich/VRH/algorithms"

type scalarBackend struct{}

func (scalarBackend) Name() string {
	return "scalar"
}

func (scalarBackend) Available() bool {
	return true
}

func (scalarBackend) OwnerVRH(key uint64, nodes []uint64) uint64 {
	if len(nodes) == 0 {
		panic("e19/simd: node list is empty")
	}

	k := algorithms.SplitMix64(key)

	owner := nodes[0]
	best := algorithms.VRHScoreK(k, owner)

	for i := 1; i < len(nodes); i++ {
		node := nodes[i]
		score := algorithms.VRHScoreK(k, node)

		if score > best {
			best = score
			owner = node
		}
	}

	return owner
}

func (scalarBackend) ScoreVRHNodes(
	key uint64,
	nodes []uint64,
	out []uint64,
) {
	if len(out) < len(nodes) {
		panic("vrh/simd: output buffer is too small")
	}

	k := algorithms.SplitMix64(key)

	for i, node := range nodes {
		out[i] = algorithms.VRHScoreK(k, node)
	}
}
