package vrh

import "github.com/dmedovich/VRH/algorithms"

// preparedBackend keeps the raw IDs for the public Nodes method and a
// transformed copy for the native large-ring scan. It is immutable after New.
type preparedBackend struct {
	raw         []uint64
	transformed []uint64
}

func newPreparedBackend(nodes []uint64) *preparedBackend {
	b := &preparedBackend{raw: nodes}
	if len(nodes) > 16 {
		b.transformed = make([]uint64, len(nodes))
		for i, node := range nodes {
			b.transformed[i] = algorithms.LinearNode(node)
		}
	}
	return b
}

func (*preparedBackend) Name() string    { return "scalar" }
func (*preparedBackend) Available() bool { return true }

func (b *preparedBackend) OwnerVRH(key uint64, _ []uint64) uint64 {
	if len(b.raw) == 1 {
		return b.raw[0]
	}
	preparedKey := algorithms.VRHKey(key)
	if len(b.transformed) == 0 {
		bestNode := b.raw[0]
		best := algorithms.VRHScoreK(preparedKey, bestNode)
		for _, node := range b.raw[1:] {
			if score := algorithms.VRHScoreK(preparedKey, node); score > best {
				best, bestNode = score, node
			}
		}
		return bestNode
	}
	transformedKey := algorithms.LinearNode(preparedKey)
	return algorithms.InverseLinearNode(ownerPrepared(transformedKey, b.transformed))
}

func (*preparedBackend) ScoreVRHNodes(key uint64, nodes, out []uint64) {
	if len(out) < len(nodes) {
		panic("vrh: output buffer is too small")
	}
	k := algorithms.VRHKey(key)
	for i, node := range nodes {
		out[i] = algorithms.VRHScoreK(k, node)
	}
}
