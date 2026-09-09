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

// fakeBackend is a minimal simd.Backend used to prove that OwnerBatch
// actually delegates scoring to the ring's configured backend, instead of
// silently recomputing scores with a hardcoded scalar path. A regression
// here means OwnerBatch stopped honoring WithBackend, which previously
// caused it to always run at scalar speed even when AVX2 was requested.
type fakeBackend struct {
	scoreCalls int
}

func (f *fakeBackend) Name() string    { return "fake" }
func (f *fakeBackend) Available() bool { return true }

func (f *fakeBackend) OwnerVRH(key uint64, nodes []uint64) uint64 {
	panic("fakeBackend.OwnerVRH should not be called by OwnerBatch")
}

func (f *fakeBackend) ScoreVRHNodes(key uint64, nodes []uint64, out []uint64) {
	f.scoreCalls++

	for i, node := range nodes {
		out[i] = node // deterministic, order-preserving stand-in score
	}
}

func TestOwnerBatchDelegatesToConfiguredBackend(t *testing.T) {
	fb := &fakeBackend{}

	ring := &Ring{
		nodes:   []uint64{5, 1, 9, 3},
		backend: fb,
	}

	keys := []uint64{100, 200, 300}

	got := ring.OwnerBatch(keys, nil)

	if fb.scoreCalls != len(keys) {
		t.Fatalf(
			"expected ScoreVRHNodes to be called once per key (%d calls), got %d",
			len(keys),
			fb.scoreCalls,
		)
	}

	// fakeBackend scores each node as itself, so the max-scoring node is
	// always the largest node ID (9), independent of the key.
	for i, owner := range got {
		if owner != 9 {
			t.Fatalf("key=%d: expected owner=9 from fakeBackend, got %d", keys[i], owner)
		}
	}
}
