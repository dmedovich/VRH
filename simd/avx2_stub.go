//go:build !amd64

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

func (avx2Backend) ScoreE19Nodes(key uint64, nodes []uint64, out []uint64) {
	panic("e19/simd: AVX2 backend is not available on this architecture")
}

func newAutoBackend() Backend {
	return scalarBackend{}
}
