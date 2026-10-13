package vrh

// OwnerBatch calculates the owner for every key using the ring's
// configured N3 owner scan.
//
// dst must have length at least len(keys). If dst is nil or too small,
// a new slice is allocated.
//
// With a sufficiently long dst, the call does not allocate. Each key uses
// the backend's ownership scan, avoiding an intermediate score buffer and
// a separate maximum search. keys and dst may be the same slice; otherwise
// they must not overlap.
func (r *Ring) OwnerBatch(keys []uint64, dst []uint64) []uint64 {
	if len(dst) < len(keys) {
		dst = make([]uint64, len(keys))
	} else {
		dst = dst[:len(keys)]
	}

	for i, key := range keys {
		dst[i] = r.backend.OwnerVRH(key, r.nodes)
	}

	return dst
}
