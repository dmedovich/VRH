package simd

import (
	"github.com/dmedovich/VRH/algorithms"
	"math/rand"
	"testing"
)

func TestAVX2OwnerMatchesScalar(t *testing.T) {
	if !newAVX2Backend().Available() {
		t.Skip("AVX2 is not available on this CPU")
	}

	rng := rand.New(rand.NewSource(42))
	backend := newAVX2Backend()

	for nodesCount := 1; nodesCount <= 1024; nodesCount++ {
		nodes := make([]uint64, nodesCount)
		scores := make([]uint64, nodesCount)

		for i := range nodes {
			nodes[i] = rng.Uint64()
		}

		for range 16 {
			key := rng.Uint64()

			backend.ScoreVRHNodes(key, nodes, scores)

			simdOwner := nodes[0]
			simdBest := scores[0]

			for i := 1; i < len(nodes); i++ {
				if scores[i] > simdBest {
					simdBest = scores[i]
					simdOwner = nodes[i]
				}
			}

			scalarOwner := nodes[0]
			scalarBest := algorithms.VRHScore(key, nodes[0])

			for i := 1; i < len(nodes); i++ {
				score := algorithms.VRHScore(key, nodes[i])

				if score > scalarBest {
					scalarBest = score
					scalarOwner = nodes[i]
				}
			}

			if simdOwner != scalarOwner {
				t.Fatalf(
					"nodes=%d key=%016x: SIMD owner=%016x score=%016x, scalar owner=%016x score=%016x",
					nodesCount,
					key,
					simdOwner,
					simdBest,
					scalarOwner,
					scalarBest,
				)
			}
		}
	}
}

func TestAVX2OwnerUnsignedOrdering(t *testing.T) {
	if !newAVX2Backend().Available() {
		t.Skip("AVX2 is not available on this CPU")
	}

	backend := newAVX2Backend()

	nodes := []uint64{
		0,
		1,
		0x7fffffffffffffff,
		0x8000000000000000,
		0xfffffffffffffffe,
		0xffffffffffffffff,
	}

	for key := range uint64(1000) {
		got := backend.OwnerVRH(key, nodes)

		want := nodes[0]
		best := algorithms.VRHScore(key, nodes[0])

		for _, node := range nodes[1:] {
			score := algorithms.VRHScore(key, node)

			if score > best {
				best = score
				want = node
			}
		}

		if got != want {
			t.Fatalf(
				"key=%016x: AVX2 owner=%016x scalar owner=%016x",
				key,
				got,
				want,
			)
		}
	}
}
