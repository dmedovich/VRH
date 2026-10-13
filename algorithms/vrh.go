package algorithms

import "math/bits"

// ScoreVersion identifies the ownership mapping. Version 3 uses N3 and
// remaps keys relative to versions 1 and 2.
const ScoreVersion = 3

// VRHScore returns the N3 rendezvous score for a key and node ID.
// For a fixed key, the node-to-score mapping is a permutation of uint64.
// This prevents ties; it does not prove uniform ownership for every node set.
func VRHScore(key, node uint64) uint64 {
	return VRHScoreK(VRHKey(key), node)
}

// VRHKey prepares a key once for a scan over multiple nodes.
func VRHKey(key uint64) uint64 {
	key += 0x9e3779b97f4a7c15
	key ^= key >> 32
	key *= 0xbf58476d1ce4e5b9
	return key ^ key>>27
}

// VRHScoreK scores a node using a key prepared by VRHKey.
func VRHScoreK(k, node uint64) uint64 {
	return n3Tail(LinearNode(k ^ node))
}

// LinearNode is N3's invertible linear transform. Preparing nodes with it
// at ring construction avoids repeating this transform for every lookup.
func LinearNode(x uint64) uint64 {
	return x ^ bits.RotateLeft64(x, 17) ^ bits.RotateLeft64(x, 43)
}

// InverseLinearNode reverses LinearNode.
func InverseLinearNode(x uint64) uint64 {
	for i := range 6 {
		x ^= bits.RotateLeft64(x, (17<<i)%64) ^ bits.RotateLeft64(x, (43<<i)%64)
	}
	return x
}

// VRHScorePreparedK scores a transformed node using a transformed prepared key.
func VRHScorePreparedK(transformedKey, transformedNode uint64) uint64 {
	return n3Tail(transformedKey ^ transformedNode)
}

func n3Tail(x uint64) uint64 {
	x *= 0x2cb0f69f4abea221
	x ^= x >> 27
	x = bits.RotateLeft64(x, 23)
	x *= 0xb19e04cb9dce3147
	x ^= x >> 31
	x *= 0xbf58476d1ce4e5b9
	return x ^ x>>33
}
