package simd

import (
	"github.com/dmedovich/VRH/algorithms"
	"math/bits"
	"math/rand"
	"testing"
)

func TestAVX2MatchesScalar(t *testing.T) {
	if !newAVX2Backend().Available() {
		t.Skip("AVX2 is not available on this CPU")
	}

	const (
		keysCount  = 128
		nodesCount = 257
	)

	rng := rand.New(rand.NewSource(1))

	nodes := make([]uint64, nodesCount)
	for i := range nodes {
		nodes[i] = rng.Uint64()
	}

	out := make([]uint64, nodesCount)

	backend := newAVX2Backend()

	for range keysCount {
		key := rng.Uint64()

		backend.ScoreVRHNodes(key, nodes, out)

		for i, node := range nodes {
			want := algorithms.VRHScore(key, node)
			got := out[i]

			if got != want {
				t.Fatalf(
					"key=%d nodeIndex=%d node=%016x: SIMD=%016x scalar=%016x",
					key,
					i,
					node,
					got,
					want,
				)
			}
		}
	}
}

func TestAVX2MatchesScalarForSmallTails(t *testing.T) {
	if !newAVX2Backend().Available() {
		t.Skip("AVX2 is not available on this CPU")
	}

	backend := newAVX2Backend()

	key := uint64(0x123456789abcdef0)

	for nodesCount := 1; nodesCount <= 15; nodesCount++ {
		nodes := make([]uint64, nodesCount)
		out := make([]uint64, nodesCount)

		for i := range nodes {
			nodes[i] = uint64(i)*0x9e3779b97f4a7c15 + 0x1234
		}

		backend.ScoreVRHNodes(key, nodes, out)

		for i, node := range nodes {
			want := algorithms.VRHScore(key, node)

			if out[i] != want {
				t.Fatalf(
					"nodes=%d index=%d: SIMD=%016x scalar=%016x",
					nodesCount,
					i,
					out[i],
					want,
				)
			}
		}
	}
}

// TestAVX2ScoreAdversarialValues checks ScoreVRHNodes against known
// tricky uint64 values (0, max uint64, the sign-bit-only value, and a
// mix of small/large keys) across a range of node-count tails so the
// AVX2 vector path and its scalar remainder both get exercised.
func TestAVX2ScoreAdversarialValues(t *testing.T) {
	if !newAVX2Backend().Available() {
		t.Skip("AVX2 is not available on this CPU")
	}

	adversarialNodes := []uint64{
		0,
		1,
		0x7fffffffffffffff,
		0x8000000000000000,
		0x8000000000000001,
		0xfffffffffffffffe,
		0xffffffffffffffff,
	}

	adversarialKeys := []uint64{
		0,
		1,
		0x7fffffffffffffff,
		0x8000000000000000,
		0xffffffffffffffff,
		0x123456789abcdef0,
	}

	backend := newAVX2Backend()

	// Exercise counts on both sides of the 4-lane boundary (pure tail,
	// exactly one vector block, one block + tail, multiple blocks).
	for _, n := range []int{1, 2, 3, 4, 5, 7, 8, 9, len(adversarialNodes)} {
		nodes := make([]uint64, n)
		for i := range nodes {
			nodes[i] = adversarialNodes[i%len(adversarialNodes)]
		}

		out := make([]uint64, n)

		for _, key := range adversarialKeys {
			backend.ScoreVRHNodes(key, nodes, out)

			for i, node := range nodes {
				want := algorithms.VRHScore(key, node)

				if out[i] != want {
					t.Fatalf(
						"n=%d key=%016x index=%d node=%016x: AVX2=%016x scalar=%016x",
						n, key, i, node, out[i], want,
					)
				}
			}
		}
	}
}

// TestAVX2SingleBlockRegression pins the exact key/nodes vector from the
// register-aliasing bug report (X1/Y1 clobbered by the scratch register
// reused to build Y2's broadcast): a single 4-lane AVX2 block must match
// both algorithms.E19Score() and the previously-verified scalar constants.
//
// This guards specifically against the aliasing bug regressing: Xn is
// the low 128 bits of Yn, so writing to Xn as scratch for a *different*
// Yn silently zeroes/overwrites the Yn that already aliased it.
func TestAVX2SingleBlockRegression(t *testing.T) {
	if !newAVX2Backend().Available() {
		t.Skip("AVX2 is not available on this CPU")
	}

	key := uint64(0xfa5ee246c455bd05)
	nodes := []uint64{
		0x4d65822107fcfd52,
		1,
		2,
		3,
	}

	wantK := uint64(0x82007e6d0ad12448)
	if k := algorithms.SplitMix64(key); k != wantK {
		t.Fatalf("SplitMix64(key)=%016x, want %016x", k, wantK)
	}

	out := make([]uint64, len(nodes))
	avx2Backend{}.ScoreVRHNodes(key, nodes, out)

	k := algorithms.SplitMix64(key)

	for i, node := range nodes {
		want := algorithms.VRHScore(key, node)

		x := node ^ k
		x += bits.RotateLeft64(k, 23)
		x ^= x >> 29
		x += bits.RotateLeft64(k, 41)
		x ^= x << 17
		x ^= x >> 32

		if x != want {
			t.Fatalf("test harness bug: manual computation %016x != algorithms.VRHScore %016x", x, want)
		}

		if out[i] != want {
			t.Fatalf(
				"lane%d node=%016x: AVX2=%016x scalar=%016x",
				i, node, out[i], want,
			)
		}
	}
}
