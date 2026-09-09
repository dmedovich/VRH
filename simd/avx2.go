//go:build amd64

package simd

import (
	"github.com/dmedovich/VRH/algorithms"

	"golang.org/x/sys/cpu"
)

type avx2Backend struct{}

func newAVX2Backend() Backend {
	return avx2Backend{}
}

func (avx2Backend) Name() string {
	return "avx2"
}

func (avx2Backend) Available() bool {
	return cpu.X86.HasAVX2
}

func (avx2Backend) OwnerVRH(key uint64, nodes []uint64) uint64 {
	if len(nodes) == 0 {
		panic("vrh/simd: node list is empty")
	}

	if !cpu.X86.HasAVX2 {
		panic("vrh/simd: AVX2 is not available")
	}

	k := algorithms.SplitMix64(key)

	return ownerVRHAVX2(k, &nodes[0], len(nodes))
}

// ownerVRHAVX2 returns the node with the highest VRH score.
//
// k is already mixed with SplitMix64 so the assembly hot loop only
// performs the VRH node-scoring operations and the argmax reduction.
//
//go:noescape
func ownerVRHAVX2(
	k uint64,
	nodes *uint64,
	n int,
) uint64

func (avx2Backend) ScoreVRHNodes(key uint64, nodes []uint64, out []uint64) {
	if len(out) < len(nodes) {
		panic("vrh/simd: output buffer is too small")
	}

	if len(nodes) == 0 {
		return
	}

	if !cpu.X86.HasAVX2 {
		panic("vrh/simd: AVX2 is not available")
	}

	k := algorithms.SplitMix64(key)

	scoreVRHAVX2(k, &nodes[0], &out[0], len(nodes))
}

// scoreVRHAVX2 is implemented in avx2_amd64.s.
//
// It receives the already-mixed key k so that the assembly kernel
// only contains the node-scoring operations.
//
//go:noescape
func scoreVRHAVX2(
	k uint64,
	nodes *uint64,
	out *uint64,
	n int,
)

func newAutoBackend() Backend {
	if cpu.X86.HasAVX2 {
		return avx2Backend{}
	}

	return scalarBackend{}
}
