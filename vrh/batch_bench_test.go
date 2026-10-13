package vrh

import (
	"strconv"
	"testing"

	"github.com/dmedovich/VRH/algorithms"
)

var batchSink []uint64

func BenchmarkOwnerBatch(b *testing.B) {
	keys := make([]uint64, 64)
	for i := range keys {
		keys[i] = algorithms.SplitMix64(uint64(i) + 12345)
	}
	for _, n := range []int{256, 1024, 16384} {
		nodes := make([]uint64, n)
		for i := range nodes {
			nodes[i] = algorithms.SplitMix64(uint64(i) + 98765)
		}
		r, err := New(nodes)
		if err != nil {
			b.Fatal(err)
		}
		b.Run("n="+strconv.Itoa(n)+"/"+r.Backend(), func(b *testing.B) {
			dst := make([]uint64, len(keys))
			b.ReportAllocs()
			for b.Loop() {
				dst = r.OwnerBatch(keys, dst)
			}
			batchSink = dst
		})
	}
}
