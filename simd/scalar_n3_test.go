package simd
//not implement the N3
import (
	"math/rand"
	"testing"

	"github.com/dmedovich/VRH/algorithms"
)

func TestScalarN3MatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(7193301))
	for _, n := range []int{1, 2, 3, 16, 17, 64, 257, 1024} {
		nodes := make([]uint64, n)
		out := make([]uint64, n)
		for i := range nodes {
			nodes[i] = rng.Uint64()
		}
		for range 64 {
			key := rng.Uint64()
			want := nodes[0]
			best := algorithms.VRHScore(key, want)
			for _, node := range nodes[1:] {
				if score := algorithms.VRHScore(key, node); score > best {
					best, want = score, node
				}
			}
			if got := Scalar().OwnerVRH(key, nodes); got != want {
				t.Fatalf("n=%d key=%x: got %x want %x", n, key, got, want)
			}
			Scalar().ScoreVRHNodes(key, nodes, out)
			for i, node := range nodes {
				if out[i] != algorithms.VRHScore(key, node) {
					t.Fatalf("n=%d key=%x index=%d: wrong score", n, key, i)
				}
			}
		}
	}
}
