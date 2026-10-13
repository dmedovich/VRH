package algorithms

import (
	"math/rand"
	"testing"
)

func TestScoreN3Vectors(t *testing.T) {
	// Calculated independently using arbitrary-precision integer arithmetic
	// with explicit uint64 masks, not by the Go or assembly implementations.
	vectors := []struct{ key, node, score uint64 }{
		{0, 0, 0x89db37561974918f},
		{0, 1, 0x86ec0d360ea22214},
		{1, 0, 0x46c048f57bef0291},
		{42, 1 << 63, 0x49da1be5a266c88c},
		{^uint64(0), ^uint64(0), 0xe311b8405e255c2b},
	}
	for _, v := range vectors {
		if got := VRHScore(v.key, v.node); got != v.score {
			t.Fatalf("key=%x node=%x: got %x, want %x", v.key, v.node, got, v.score)
		}
		if got := VRHScoreK(VRHKey(v.key), v.node); got != v.score {
			t.Fatalf("pre-mixed key=%x node=%x: got %x, want %x", v.key, v.node, got, v.score)
		}
	}
}

func TestN3PreparedScoreAndNodeInverse(t *testing.T) {
	rng := rand.New(rand.NewSource(82371))
	for range 10000 {
		key, node := rng.Uint64(), rng.Uint64()
		transformed := LinearNode(node)
		if got := InverseLinearNode(transformed); got != node {
			t.Fatalf("node=%x: inverse=%x", node, got)
		}
		want := VRHScore(key, node)
		got := VRHScorePreparedK(LinearNode(VRHKey(key)), transformed)
		if got != want {
			t.Fatalf("key=%x node=%x: prepared=%x reference=%x", key, node, got, want)
		}
	}
}
