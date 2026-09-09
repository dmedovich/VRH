package metrics

import (
	"math"
	"testing"
)

func TestCalculateUsesCompleteNodeSet(t *testing.T) {
	nodes := []uint64{1, 2, 3, 4}
	owners := []uint64{1, 1, 2, 2, 3, 3}

	got := Calculate(nodes, owners, nil, nil)
	want := math.Sqrt(0.75) / 1.5

	if math.Abs(got.Balance-want) > 1e-12 {
		t.Fatalf("expected balance %.15f, got %.15f", want, got.Balance)
	}
}
