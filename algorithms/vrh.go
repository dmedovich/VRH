package algorithms

// E19Score calculates the VRH score.
//
// Every transformation of node is bijective modulo 2^64:
//
//	k  = SplitMix64(key)
//	x  = node ^ k
//	x += ROTL(k, 23)
//	x ^= x >> 29
//	x += ROTL(k, 41)
//	x ^= x << 17
//	x ^= x >> 32
//
// The construction deliberately avoids 64-bit multiplication in the
// node-scoring path so it can be implemented efficiently with AVX2.
// algorithms/vrh.go

func VRHScore(key, node uint64) uint64 {
	k := SplitMix64(key)
	return VRHScoreK(k, node)
}

func VRHScoreK(k, node uint64) uint64 {
	x := node ^ k
	x += bitsRotateLeft64(k, 23)
	x ^= x >> 29
	x += bitsRotateLeft64(k, 41)
	x ^= x << 17
	x ^= x >> 32
	return x
}

// bitsRotateLeft64 performs a 64-bit rotate-left.
func bitsRotateLeft64(x uint64, n uint) uint64 {
	return (x << n) | (x >> (64 - n))
}
