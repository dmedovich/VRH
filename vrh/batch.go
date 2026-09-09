package vrh

// OwnerBatch calculates the owner for every key using the ring's
// configured backend (scalar or AVX2).
//
// dst must have length at least len(keys). If dst is nil or too small,
// a new slice is allocated.
//
// A single score buffer is reused for all keys to avoid allocating
// once per key. Scoring is delegated to r.backend.ScoreVRHNodes, which is
// exactly the entry point the AVX2 backend implements for batch use — this
// is what makes OwnerBatch benefit from AVX2 when the ring was constructed
// with WithBackend(BackendAVX2) (or BackendAuto selects it). A prior
// version of this method called algorithms.VRHScore directly in a manual
// loop, which silently always ran the scalar path regardless of the
// configured backend.
func (r *Ring) OwnerBatch(keys []uint64, dst []uint64) []uint64 {
	if len(dst) < len(keys) {
		dst = make([]uint64, len(keys))
	} else {
		dst = dst[:len(keys)]
	}

	scores := make([]uint64, len(r.nodes))

	for i, key := range keys {
		r.backend.ScoreVRHNodes(key, r.nodes, scores)

		owner, _ := maxScore(r.nodes, scores)
		dst[i] = owner
	}

	return dst
}
