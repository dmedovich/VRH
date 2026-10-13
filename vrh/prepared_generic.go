//go:build !amd64

package vrh

import "github.com/dmedovich/VRH/algorithms"

func ownerPrepared(k uint64, nodes []uint64) uint64 {
	owner := nodes[0]
	best := algorithms.VRHScorePreparedK(k, owner)
	for _, node := range nodes[1:] {
		if score := algorithms.VRHScorePreparedK(k, node); score > best {
			best, owner = score, node
		}
	}
	return owner
}
