//go:build amd64

//not implement the N3
package simd

type avx2Backend struct{}

func newAVX2Backend() Backend {
	return avx2Backend{}
}

func (avx2Backend) Name() string {
	return "avx2"
}

func (avx2Backend) Available() bool {
	return false
}

func (avx2Backend) OwnerVRH(key uint64, nodes []uint64) uint64 {
	panic("vrh/simd: AVX2 does not implement the N3 score")
}

func (avx2Backend) ScoreVRHNodes(key uint64, nodes []uint64, out []uint64) {
	panic("vrh/simd: AVX2 does not implement the N3 score")
}

func newAutoBackend() Backend {
	return scalarBackend{}
}
