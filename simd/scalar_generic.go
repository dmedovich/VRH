//go:build !amd64
//not implement the N3
package simd

func scalarOwner(k uint64, nodes []uint64) uint64 {
	return scalarOwnerGo(k, nodes)
}
