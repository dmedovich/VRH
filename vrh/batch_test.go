package vrh

import (
	"github.com/dmedovich/VRH/simd"
	"testing"
)

// TestOwnerBatchMatchesOwner verifies that OwnerBatch produces exactly the
// same owners as calling Owner once per key, for both backends.
func TestOwnerBatchMatchesOwner(t *testing.T) {
	nodes := []uint64{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
	}

	keys := make([]uint64, 500)
	for i := range keys {
		keys[i] = uint64(i)*2654435761 + 1
	}

	backends := []Backend{BackendScalar}
	if simd.AVX2().Available() {
		backends = append(backends, BackendAVX2)
	}

	for _, backend := range backends {
		ring, err := New(nodes, WithBackend(backend))
		if err != nil {
			t.Fatalf("backend=%v: %v", backend, err)
		}

		got := ring.OwnerBatch(keys, nil)

		for i, key := range keys {
			want := ring.Owner(key)

			if got[i] != want {
				t.Fatalf(
					"backend=%v key=%d: OwnerBatch=%d Owner=%d",
					backend,
					key,
					got[i],
					want,
				)
			}
		}
	}
}

// TestOwnerBatchReusesDst verifies dst reuse semantics: a large enough dst
// is reused (truncated), and a nil/short dst causes a new allocation.
func TestOwnerBatchReusesDst(t *testing.T) {
	nodes := []uint64{1, 2, 3, 4}
	keys := []uint64{10, 20, 30}

	ring, err := New(nodes, WithBackend(BackendScalar))
	if err != nil {
		t.Fatal(err)
	}

	dst := make([]uint64, 10)
	got := ring.OwnerBatch(keys, dst)

	if len(got) != len(keys) {
		t.Fatalf("expected len %d, got %d", len(keys), len(got))
	}

	if &got[0] != &dst[0] {
		t.Fatal("expected dst buffer to be reused when large enough")
	}
}

// fakeBackend verifies that batch lookup honors the configured backend.
type fakeBackend struct {
	ownerCalls int
}

func (f *fakeBackend) Name() string    { return "fake" }
func (f *fakeBackend) Available() bool { return true }

func (f *fakeBackend) OwnerVRH(key uint64, nodes []uint64) uint64 {
	f.ownerCalls++
	return nodes[int(key%uint64(len(nodes)))]
}

func (f *fakeBackend) ScoreVRHNodes(key uint64, nodes []uint64, out []uint64) {
	panic("batch lookup should not materialize scores")
}

func TestOwnerBatchDelegatesToConfiguredBackend(t *testing.T) {
	fb := &fakeBackend{}

	ring := &Ring{
		nodes:   []uint64{5, 1, 9, 3},
		backend: fb,
	}

	keys := []uint64{100, 200, 300}

	got := ring.OwnerBatch(keys, nil)

	if fb.ownerCalls != len(keys) {
		t.Fatalf(
			"expected OwnerVRH to be called once per key (%d calls), got %d",
			len(keys),
			fb.ownerCalls,
		)
	}

	for i, owner := range got {
		want := ring.nodes[keys[i]%uint64(len(ring.nodes))]
		if owner != want {
			t.Fatalf("key=%d: expected owner=%d from fakeBackend, got %d", keys[i], want, owner)
		}
	}
}

func TestOwnerBatchNoAllocationsAndInPlace(t *testing.T) {
	for _, backend := range []Backend{BackendScalar, BackendAVX2} {
		if backend == BackendAVX2 && !simd.AVX2().Available() {
			continue
		}
		ring, err := New([]uint64{0, 1, 1 << 63, 17, 999}, WithBackend(backend))
		if err != nil {
			t.Fatal(err)
		}
		keys := []uint64{0, 123, 1 << 63, ^uint64(0)}
		dst := make([]uint64, len(keys))
		if n := testing.AllocsPerRun(100, func() { ring.OwnerBatch(keys, dst) }); n != 0 {
			t.Fatalf("backend=%v: %v allocations with reusable dst", backend, n)
		}
		want := ring.OwnerBatch(keys, nil)
		got := ring.OwnerBatch(keys, keys)
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("in-place index %d: got %d want %d", i, got[i], want[i])
			}
		}
		if len(ring.OwnerBatch(nil, dst)) != 0 {
			t.Fatal("empty batch returned owners")
		}
	}
}
