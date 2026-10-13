package algorithms

import (
	"math"
	"testing"
)

func TestScoreNodeAvalanche(t *testing.T) {
	// Test every input/output bit pair, not just the average Hamming distance.
	// This is a diffusion regression check; ownership distribution is tested
	// separately. Inputs and thresholds are fixed, so the test is reproducible.
	const samples = 8192
	var counts [64][64]int
	for i := range uint64(samples) {
		key, node := SplitMix64(i+0xa187), SplitMix64(i+0x812dc)
		k := VRHKey(key)
		base := VRHScoreK(k, node)
		for input := range uint(64) {
			diff := base ^ VRHScoreK(k, node^(uint64(1)<<input))
			for output := range uint(64) {
				counts[input][output] += int(diff >> output & 1)
			}
		}
	}
	for input := range counts {
		for output, count := range counts[input] {
			if math.Abs(float64(count)/samples-0.5) > 0.05 {
				t.Fatalf("input bit %d output bit %d: flip frequency %.4f", input, output, float64(count)/samples)
			}
		}
	}
}
