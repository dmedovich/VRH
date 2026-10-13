package vrh

import (
	"slices"
	"testing"

	"github.com/dmedovich/VRH/algorithms"
	"github.com/dmedovich/VRH/simd"
)

func TestOwnershipChangesOnlyWithMembership(t *testing.T) {
	nodes := []uint64{0, 1, 1 << 63, 17, ^uint64(0), 0x123456789abcdef0, 42}
	backends := []Backend{BackendScalar}
	if simd.AVX2().Available() {
		backends = append(backends, BackendAVX2)
	}
	for _, backend := range backends {
		newRing := func(ns []uint64) *Ring {
			r, err := New(ns, WithBackend(backend))
			if err != nil {
				t.Fatal(err)
			}
			return r
		}
		original := newRing(nodes)
		reversed := slices.Clone(nodes)
		slices.Reverse(reversed)
		reordered := newRing(reversed)
		const added = 999
		expanded := newRing(append(slices.Clone(nodes), added))
		for removed := range nodes {
			reduced := newRing(slices.Delete(slices.Clone(nodes), removed, removed+1))
			for i := uint64(0); i < 4096; i++ {
				key := algorithms.SplitMix64(i)
				owner := original.Owner(key)
				if reordered.Owner(key) != owner {
					t.Fatal("order changed owner")
				}
				if got := expanded.Owner(key); got != owner && got != added {
					t.Fatal("addition moved key between old nodes")
				}
				if got := reduced.Owner(key); owner != nodes[removed] && got != owner {
					t.Fatal("removal moved unaffected key")
				}
			}
		}
	}
}

func TestOwnerDoesNotAllocate(t *testing.T) {
	for _, backend := range []Backend{BackendScalar, BackendAVX2} {
		if backend == BackendAVX2 && !simd.AVX2().Available() {
			continue
		}
		r, err := New([]uint64{0, 1, 2, 3, 1 << 63}, WithBackend(backend))
		if err != nil {
			t.Fatal(err)
		}
		if got := testing.AllocsPerRun(100, func() { _ = r.Owner(42) }); got != 0 {
			t.Fatalf("backend=%v Owner: %v allocations", backend, got)
		}
	}
}
