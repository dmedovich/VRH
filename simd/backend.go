package simd

// Backend calculates VRH ownership for a key across a set of nodes.
type Backend interface {
	Name() string
	Available() bool

	// OwnerVRH returns the node with the highest VRH score for key.
	//
	// The implementation must not allocate a score slice per lookup.
	// The caller guarantees that nodes is non-empty.
	OwnerVRH(key uint64, nodes []uint64) uint64

	// ScoreVRHNodes calculates VRH scores for all nodes.
	//
	// out must have length >= len(nodes).
	// This method is for callers needing every score and for verification.
	ScoreVRHNodes(key uint64, nodes []uint64, out []uint64)
}

// Auto returns the scalar N3 backend. The old AVX2 code used a different
// score mapping and cannot be selected for N3.
//
// The actual backend selection is implemented in backend-specific files.
func Auto() Backend {
	return newAutoBackend()
}

// Scalar returns the portable scalar N3 backend.
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
