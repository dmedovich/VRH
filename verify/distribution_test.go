package verify

import (
	"math"
	"testing"

	"github.com/dmedovich/VRH/algorithms"
)

func TestN3OwnershipDistribution(t *testing.T) {
	sets := map[string][]uint64{
		"legacy-counterexample": {0, 1, 1 << 63},
		"v2-counterexample":     {0, 1 << 29, 1 << 59},
		"dense":                 {0, 1, 2},
		"adaptive":              {0, 73184597756084224, 1152921521787764800},
	}
	const samples = 1 << 20
	for name, nodes := range sets {
		t.Run(name, func(t *testing.T) {
			counts := make([]int, len(nodes))
			for i := range uint64(samples) {
				key := algorithms.SplitMix64(i ^ 0x529af)
				owner := nodes[0]
				best := algorithms.VRHScore(key, owner)
				for _, node := range nodes[1:] {
					if score := algorithms.VRHScore(key, node); score > best {
						best, owner = score, node
					}
				}
				for j, node := range nodes {
					if node == owner {
						counts[j]++
						break
					}
				}
			}
			mean := float64(samples) / float64(len(nodes))
			sigma := math.Sqrt(mean * (1 - 1/float64(len(nodes))))
			for _, count := range counts {
				if z := math.Abs(float64(count)-mean) / sigma; z > 7 {
					t.Fatalf("ownership skew: counts=%v maxZ=%.2f", counts, z)
				}
			}
		})
	}
}
