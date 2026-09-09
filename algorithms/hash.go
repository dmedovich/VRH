package algorithms

// SplitMix64 is the 64-bit mixing function used by all algorithms
// in this project.
//
// The function is deterministic and operates modulo 2^64.
func SplitMix64(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}
