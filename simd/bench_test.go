package simd

import (
	"math/rand"
	"strconv"
	"testing"
)

// BenchmarkScoreE19Nodes compares the scalar and AVX2 backends across a
// range of node-set sizes so the real speedup on large N is visible via
// `go test -bench=. -benchmem ./simd/...`.
func BenchmarkScoreVRHNodes(b *testing.B) {
	sizes := []int{16, 256, 4096, 65536, 1 << 20}

	backends := []struct {
		name    string
		backend Backend
	}{
		{"scalar", Scalar()},
		{"avx2", AVX2()},
	}

	rng := rand.New(rand.NewSource(7))

	for _, n := range sizes {
		nodes := make([]uint64, n)
		for i := range nodes {
			nodes[i] = rng.Uint64()
		}
		out := make([]uint64, n)
		key := rng.Uint64()

		for _, be := range backends {
			if !be.backend.Available() {
				continue
			}

			b.Run(
				be.name+"/n="+strconv.Itoa(n),
				func(b *testing.B) {
					b.SetBytes(int64(n) * 8)
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						be.backend.ScoreVRHNodes(key, nodes, out)
					}
				},
			)
		}
	}
}

// BenchmarkOwnerVRH compares scalar vs AVX2 owner lookup across sizes.
func BenchmarkOwnerVRH(b *testing.B) {
	sizes := []int{16, 256, 4096, 65536, 1 << 20}

	backends := []struct {
		name    string
		backend Backend
	}{
		{"scalar", Scalar()},
		{"avx2", AVX2()},
	}

	rng := rand.New(rand.NewSource(7))

	for _, n := range sizes {
		nodes := make([]uint64, n)
		for i := range nodes {
			nodes[i] = rng.Uint64()
		}
		key := rng.Uint64()

		for _, be := range backends {
			if !be.backend.Available() {
				continue
			}

			b.Run(
				be.name+"/n="+strconv.Itoa(n),
				func(b *testing.B) {
					b.SetBytes(int64(n) * 8)
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						_ = be.backend.OwnerVRH(key, nodes)
					}
				},
			)
		}
	}
}
