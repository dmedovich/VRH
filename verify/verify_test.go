package verify

import (
	"github.com/dmedovich/VRH/algorithms"
	"testing"
)

func TestCheckOwner(t *testing.T) {
	nodes := []uint64{1, 2, 3, 4}

	for key := range uint64(1000) {
		want := nodes[0]
		best := algorithms.VRHScore(key, nodes[0])

		for _, node := range nodes[1:] {
			score := algorithms.VRHScore(key, node)

			if score > best {
				best = score
				want = node
			}
		}

		if !CheckVRHOwner(key, nodes, want) {
			t.Fatalf("key=%d: expected owner %d to be valid", key, want)
		}
	}
}

func TestCheckVRHPermutation(t *testing.T) {
	nodes := make([]uint64, 1024)

	for i := range nodes {
		nodes[i] = uint64(i)
	}

	for key := range uint64(100) {
		if !CheckVRHPermutation(key, nodes) {
			t.Fatalf("key=%d: VRH scores are not unique", key)
		}
	}
}

func TestCheckDistinctNodes(t *testing.T) {
	if !CheckDistinctNodes([]uint64{1, 2, 3}) {
		t.Fatal("expected distinct nodes to pass")
	}

	if CheckDistinctNodes([]uint64{1, 2, 2, 3}) {
		t.Fatal("expected duplicate nodes to fail")
	}
}
