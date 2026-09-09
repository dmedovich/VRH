package vrh

import (
	"github.com/dmedovich/VRH/algorithms"
	"testing"
)

func TestRingOwnerMatchesScalar(t *testing.T) {
	nodes := []uint64{
		1,
		2,
		3,
		4,
		5,
		6,
		7,
		8,
	}

	ring, err := New(nodes, WithBackend(BackendScalar))
	if err != nil {
		t.Fatal(err)
	}

	keys := []uint64{
		0,
		1,
		2,
		3,
		42,
		123456789,
		0xffffffffffffffff,
	}

	for _, key := range keys {
		got := ring.Owner(key)
		want := ring.OwnerScalar(key)

		if got != want {
			t.Fatalf(
				"key=%016x: Owner=%016x OwnerScalar=%016x",
				key,
				got,
				want,
			)
		}
	}
}

func TestRingRejectsEmptyNodes(t *testing.T) {
	_, err := New(nil)

	if err == nil {
		t.Fatal("expected error for empty node list")
	}
}

func TestRingRejectsDuplicateNodes(t *testing.T) {
	_, err := New([]uint64{1, 2, 2, 3})

	if err == nil {
		t.Fatal("expected error for duplicate node ID")
	}
}

func TestRingNodesReturnsCopy(t *testing.T) {
	nodes := []uint64{10, 20, 30}

	ring, err := New(nodes)
	if err != nil {
		t.Fatal(err)
	}

	got := ring.Nodes()

	got[0] = 999

	if ring.Nodes()[0] != 10 {
		t.Fatal("Nodes returned internal slice")
	}
}

func TestVRHScoreMatchesRingScalar(t *testing.T) {
	nodes := []uint64{
		11,
		22,
		33,
		44,
	}

	ring, err := New(nodes, WithBackend(BackendScalar))
	if err != nil {
		t.Fatal(err)
	}

	for key := range uint64(1000) {
		got := ring.OwnerScalar(key)

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
				"key=%d: OwnerScalar=%d expected=%d",
				key,
				got,
				want,
			)
		}
	}
}
