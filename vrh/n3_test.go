package vrh

import (
	"math/rand"
	"testing"

	"github.com/dmedovich/VRH/algorithms"
)

func TestN3PreparedOwnerMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(97132))
	for _, n := range []int{1, 2, 3, 16, 17, 31, 64, 65, 256, 1024} {
		nodes := make([]uint64, n)
		for i := range nodes {
			nodes[i] = rng.Uint64()
		}
		ring, err := New(nodes)
		if err != nil {
			t.Fatal(err)
		}
		for range 256 {
			key := rng.Uint64()
			want := nodes[0]
			best := algorithms.VRHScore(key, want)
			for _, node := range nodes[1:] {
				if score := algorithms.VRHScore(key, node); score > best {
					best, want = score, node
				}
			}
			if got := ring.Owner(key); got != want {
				t.Fatalf("n=%d key=%x: got %x want %x", n, key, got, want)
			}
		}
	}
}

func TestN3RejectsOldAVX2Backend(t *testing.T) {
	if _, err := New([]uint64{1, 2}, WithBackend(BackendAVX2)); err == nil {
		t.Fatal("AVX2 backend must not compute a different score mapping")
	}
}
