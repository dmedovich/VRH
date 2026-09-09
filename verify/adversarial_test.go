package verify

import (
	"github.com/dmedovich/VRH/algorithms"
	"testing"
)

func TestVRHComplementCollision(t *testing.T) {
	nodes := []uint64{
		0,
		1,
		2,
		42,
		0x123456789abcdef0,
		0x8000000000000000,
		0xffffffffffffffff,
	}

	for key := range uint64(1000) {
		seen := make(map[uint64]uint64, len(nodes))

		for _, node := range nodes {
			score := algorithms.VRHScore(key, node)

			if previous, exists := seen[score]; exists {
				t.Fatalf(
					"key=%d: collision: nodes %016x and %016x have score %016x",
					key,
					previous,
					node,
					score,
				)
			}

			seen[score] = node
		}
	}
}
