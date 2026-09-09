package metrics

import (
	"math"
	"testing"
)

func TestBalancePerfect(t *testing.T) {
	nodes := []uint64{1, 2, 3, 4}
	owners := []uint64{
		1, 2, 3, 4,
		1, 2, 3, 4,
	}

	got := Balance(nodes, owners)

	if got != 0 {
		t.Fatalf("expected perfect balance, got %v", got)
	}
}

func TestBalanceImperfect(t *testing.T) {
	nodes := []uint64{1, 2, 3, 4}
	owners := []uint64{
		1, 1, 1, 1,
		2, 2,
		3,
		4,
	}

	got := Balance(nodes, owners)

	if got <= 0 {
		t.Fatalf("expected non-zero imbalance, got %v", got)
	}
}

func TestBalanceIncludesDeadNode(t *testing.T) {
	nodes := []uint64{1, 2, 3, 4}
	owners := []uint64{1, 1, 2, 2, 3, 3}

	// Counts are [2, 2, 2, 0]. The mean is 1.5, so the coefficient of
	// variation is sqrt(((0.5)^2*3 + (-1.5)^2)/4) / 1.5 = sqrt(0.75)/1.5.
	want := math.Sqrt(0.75) / 1.5
	got := Balance(nodes, owners)

	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("expected %.15f, got %.15f", want, got)
	}
}

func TestBalanceCounts(t *testing.T) {
	nodes := []uint64{10, 20, 30, 40}
	owners := []uint64{
		10, 10, 20, 20, 20, 30,
	}

	got := BalanceCounts(nodes, owners)

	want := map[uint64]int{
		10: 2,
		20: 3,
		30: 1,
		40: 0,
	}

	for node, expected := range want {
		if got[node] != expected {
			t.Fatalf(
				"node %d: expected %d, got %d",
				node,
				expected,
				got[node],
			)
		}
	}
}

func TestBalanceCountsUnknownOwnerPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for owner outside node set")
		}
	}()

	BalanceCounts([]uint64{10, 20}, []uint64{10, 30})
}

func TestBalanceNumerically(t *testing.T) {
	nodes := []uint64{1, 2, 3, 4}
	owners := []uint64{
		1, 1,
		2, 2,
		3, 3,
		4, 4,
	}

	got := Balance(nodes, owners)

	if math.Abs(got) > 1e-12 {
		t.Fatalf("expected zero imbalance, got %.15f", got)
	}
}
