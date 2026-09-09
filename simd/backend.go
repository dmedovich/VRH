package simd

// Backend calculates E19 ownership for a key across a set of nodes.
type Backend interface {
	Name() string
	Available() bool

	// OwnerE19 returns the node with the highest E19 score for key.
	//
	// The implementation must not allocate a score slice per lookup.
	// The caller guarantees that nodes is non-empty.
	OwnerVRH(key uint64, nodes []uint64) uint64

	// ScoreE19Nodes calculates E19 scores for all nodes.
	//
	// out must have length >= len(nodes).
	// This method is used by batch operations and verification.
	ScoreVRHNodes(key uint64, nodes []uint64, out []uint64)
}

// Auto returns the best backend available on the current CPU.
//
// The actual backend selection is implemented in backend-specific files.
func Auto() Backend {
	return newAutoBackend()
}

// Scalar returns the portable scalar backend.
func Scalar() Backend {
	return scalarBackend{}
}

// AVX2 returns the AVX2 backend.
//
// Availability is reported by Available(). The backend must not be used
// when Available() returns false.
func AVX2() Backend {
	return newAVX2Backend()
}
