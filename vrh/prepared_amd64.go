//go:build amd64

package vrh

func ownerPrepared(k uint64, nodes []uint64) uint64 {
	return ownerN3Prepared(k, &nodes[0], len(nodes))
}

//go:noescape
func ownerN3Prepared(k uint64, nodes *uint64, n int) uint64
