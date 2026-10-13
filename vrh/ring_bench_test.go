package vrh

import (
	"testing"
)

func benchmarkOwner(b *testing.B, backend Backend, nodesCount int) {
	nodes := make([]uint64, nodesCount)

	for i := range nodes {
		nodes[i] = uint64(i) + 1
	}

	ring, err := New(nodes, WithBackend(backend))
	if err != nil {
		b.Fatal(err)
	}

	var owner uint64

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		owner = ring.Owner(uint64(i))
	}

	_ = owner
}

func BenchmarkOwnerScalarN3(b *testing.B) {
	benchmarkOwner(b, BackendScalar, 3)
}

func BenchmarkOwnerScalarN16(b *testing.B) {
	benchmarkOwner(b, BackendScalar, 16)
}

func BenchmarkOwnerScalarN64(b *testing.B) {
	benchmarkOwner(b, BackendScalar, 64)
}

func BenchmarkOwnerScalarN256(b *testing.B) {
	benchmarkOwner(b, BackendScalar, 256)
}

func BenchmarkOwnerScalarN1024(b *testing.B) {
	benchmarkOwner(b, BackendScalar, 1024)
}
