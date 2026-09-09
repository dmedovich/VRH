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

func BenchmarkOwnerScalarN64(b *testing.B) {
	benchmarkOwner(b, BackendScalar, 64)
}

func BenchmarkOwnerScalarN256(b *testing.B) {
	benchmarkOwner(b, BackendScalar, 256)
}

func BenchmarkOwnerScalarN1024(b *testing.B) {
	benchmarkOwner(b, BackendScalar, 1024)
}

func BenchmarkOwnerAVX2N64(b *testing.B) {
	ring, err := New(
		makeNodes(64),
		WithBackend(BackendAVX2),
	)
	if err != nil {
		b.Skip(err)
	}

	if ring.Backend() != "avx2" {
		b.Skip("AVX2 backend unavailable")
	}

	var owner uint64

	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		owner = ring.Owner(uint64(i))
	}

	_ = owner
}

func BenchmarkOwnerAVX2N256(b *testing.B) {
	ring, err := New(
		makeNodes(256),
		WithBackend(BackendAVX2),
	)
	if err != nil {
		b.Skip(err)
	}

	if ring.Backend() != "avx2" {
		b.Skip("AVX2 backend unavailable")
	}

	var owner uint64

	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		owner = ring.Owner(uint64(i))
	}

	_ = owner
}

func BenchmarkOwnerAVX2N1024(b *testing.B) {
	ring, err := New(
		makeNodes(1024),
		WithBackend(BackendAVX2),
	)
	if err != nil {
		b.Skip(err)
	}

	if ring.Backend() != "avx2" {
		b.Skip("AVX2 backend unavailable")
	}

	var owner uint64

	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		owner = ring.Owner(uint64(i))
	}

	_ = owner
}

func makeNodes(n int) []uint64 {
	nodes := make([]uint64, n)

	for i := range nodes {
		nodes[i] = uint64(i) + 1
	}

	return nodes
}
