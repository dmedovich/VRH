package simd
//not implement the N3
import (
	"math/rand"
	"strconv"
	"testing"
)

// BenchmarkScoreVRHNodes measures the scalar N3 score-array path.
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

// BenchmarkOwnerVRH measures complete ownership scans with changing keys.
// Dataset generation is outside timing; the backend's own per-lookup
// N3 key preparation is included in the measurement.
func BenchmarkOwnerVRH(b *testing.B) {
	keys := make([]uint64, 4096)
	rng := rand.New(rand.NewSource(731))
	for i := range keys {
		keys[i] = rng.Uint64()
	}
	backends := []Backend{Scalar(), AVX2()}
	for _, n := range []int{16, 64, 256, 1024, 16384} {
		for _, layout := range []string{"sequential", "scattered"} {
			nodes := make([]uint64, n)
			for i := range nodes {
				nodes[i] = uint64(i)
				if layout == "scattered" {
					nodes[i] = rng.Uint64()
				}
			}
			for _, backend := range backends {
				if !backend.Available() {
					continue
				}
				b.Run(layout+"/n="+strconv.Itoa(n)+"/"+backend.Name(), func(b *testing.B) {
					b.ReportAllocs()
					var owner uint64
					for i := 0; b.Loop(); i++ {
						owner = backend.OwnerVRH(keys[i&4095], nodes)
					}
					ownerSink = owner
				})
			}
		}
	}
}

var ownerSink uint64
