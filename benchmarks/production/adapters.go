// Package hrwbench provides adapters for an equal-weight, single-owner workload.
package production

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand/v2"

	"github.com/cespare/xxhash/v2"
	rendezvous "github.com/dgryski/go-rendezvous"
	"github.com/dmedovich/VRH/algorithms"
	"github.com/dmedovich/VRH/vrh"
	nspcc "github.com/nspcc-dev/hrw/v2"
)

// Key holds equivalent uint64, byte and string representations. Encoding is
// outside lookup timing; each implementation's own query hashing is included.
type Key struct {
	Value  uint64
	Bytes  [8]byte
	String string
}

func NewKey(v uint64) Key {
	k := Key{Value: v}
	binary.LittleEndian.PutUint64(k.Bytes[:], v)
	k.String = string(k.Bytes[:])
	return k
}

func Dataset(n int, layout string) ([]uint64, []Key) {
	nodes := make([]uint64, n)
	for i := range nodes {
		nodes[i] = uint64(i)
		if layout == "scattered" {
			nodes[i] = algorithms.SplitMix64(uint64(i) + 0xc0356)
		}
	}
	keys := make([]Key, 4096)
	for i := range keys {
		keys[i] = NewKey(algorithms.SplitMix64(uint64(i) + 0x814ac))
	}
	return nodes, keys
}

type Implementation struct {
	Name  string
	Owner func(Key) uint64
}

func Implementations(nodes []uint64) []Implementation {
	r, err := vrh.New(nodes)
	if err != nil {
		panic(err)
	}
	impls := []Implementation{{"vrh-ring", func(k Key) uint64 { return r.Owner(k.Value) }}}
	return append(impls, ExternalImplementations(nodes)...)
}

var Sizes = []int{1, 2, 3, 7, 15, 16, 17, 31, 32, 63, 64, 65, 256, 1024, 4096, 16384}
var Names = []string{"vrh-ring", "go-rendezvous-xxhash", "nspcc-score-scan-adapter", "xxhash-pair-reference"}

// Build includes representation conversion and each constructor's own checks.
func Build(name string, nodes []uint64) Implementation {
	switch name {
	case "vrh-ring":
		r, err := vrh.New(nodes)
		if err != nil {
			panic(err)
		}
		return Implementation{name, func(k Key) uint64 { return r.Owner(k.Value) }}
	case "go-rendezvous-xxhash":
		names := make([]string, len(nodes))
		for i, n := range nodes {
			names[i] = NewKey(n).String
		}
		r := rendezvous.New(names, xxhash.Sum64String)
		return Implementation{name, func(k Key) uint64 { return binary.LittleEndian.Uint64([]byte(r.Lookup(k.String))) }}
	case "nspcc-score-scan-adapter":
		r := NewNSPCCScan(nodes)
		return Implementation{name, r.Owner}
	case "xxhash-pair-reference":
		copyNodes := append([]uint64(nil), nodes...)
		return Implementation{name, func(k Key) uint64 { return PairOwner(copyNodes, k) }}
	default:
		panic(name)
	}
}

func Keys(label, stream string, count int) []Key {
	seed := sha256.Sum256([]byte("production-audit-2026-10-13/" + label + "/" + stream))
	rng := rand.NewChaCha8(seed)
	start := binary.LittleEndian.Uint64(seed[:8])
	keys := make([]Key, count)
	for i := range keys {
		v := start + uint64(i)
		if stream == "random" {
			v = rng.Uint64()
		}
		if stream == "highbits" {
			v = (uint64(uint32(start)+uint32(i)) << 32) | uint64(uint32(start>>32))
		}
		keys[i] = NewKey(v)
	}
	return keys
}

type NamedSet struct {
	Name  string
	Nodes []uint64
}

func KnownSets() []NamedSet {
	return []NamedSet{
		{"v1", []uint64{0, 1, 1 << 63}},
		{"v2", []uint64{0, 1 << 29, 1 << 59}},
		{"murmur", []uint64{0, 1 << 30, 1 << 63}},
		{"r2", []uint64{0, 0xa27065f34856f513, 0x89c197cd215bd44e}},
		{"r3", []uint64{0, 0xccb4e6a4d860bfaa, 0x969cd49b0c17f559}},
		{"adaptive", []uint64{0, 73184597756084224, 1152921521787764800}},
	}
}

// ExternalImplementations returns the comparison adapters without production
// VRH backends. This lets N3-only audits avoid constructing or reporting v2.
func ExternalImplementations(nodes []uint64) []Implementation {
	var impls []Implementation
	strings := make([]string, len(nodes))
	for i, node := range nodes {
		strings[i] = NewKey(node).String
	}
	rdv := rendezvous.New(strings, xxhash.Sum64String)
	impls = append(impls, Implementation{"go-rendezvous-xxhash", func(k Key) uint64 {
		// Normalize the returned 8-byte node name to uint64. Go's read-only
		// string-to-byte conversion here does not allocate; benches verify it.
		return binary.LittleEndian.Uint64([]byte(rdv.Lookup(k.String)))
	}})
	scan := NewNSPCCScan(nodes)
	impls = append(impls, Implementation{"nspcc-score-scan-adapter", scan.Owner})
	impls = append(impls, Implementation{"xxhash-pair-reference", func(k Key) uint64 { return PairOwner(nodes, k) }})
	return impls
}

type HashedNode struct {
	ID     uint64
	Digest uint64
}

func (n HashedNode) Hash() uint64 { return n.Digest }

func PrepareNSPCC(nodes []uint64) []HashedNode {
	prepared := make([]HashedNode, len(nodes))
	for i, node := range nodes {
		k := NewKey(node)
		prepared[i] = HashedNode{node, nspcc.Hash(k.Bytes[:])}
	}
	return prepared
}

// NSPCCScan is our O(N) adapter, not an upstream exported operation.
// It uses the distance formula from nspcc-dev/hrw/v2 v2.0.4 and chooses the
// minimum, matching the first element of that library's ascending Sort.
// Hashes are stored densely, as in go-rendezvous, to avoid a stride penalty
// from scanning ID/hash pairs. IDs are accessed only after finding the minimum.
type NSPCCScan struct {
	nodes  []uint64
	hashes []uint64
}

func NewNSPCCScan(nodes []uint64) *NSPCCScan {
	s := &NSPCCScan{nodes: append([]uint64(nil), nodes...), hashes: make([]uint64, len(nodes))}
	for i, node := range nodes {
		k := NewKey(node)
		s.hashes[i] = nspcc.Hash(k.Bytes[:])
	}
	return s
}

func (s *NSPCCScan) Owner(key Key) uint64 {
	k := nspcc.Hash(key.Bytes[:])
	index, best := 0, nspccDistance(s.hashes[0], k)
	for i := 1; i < len(s.hashes); i++ {
		if score := nspccDistance(s.hashes[i], k); score < best {
			index, best = i, score
		}
	}
	return s.nodes[index]
}

// MurmurHash3 fmix64, as used by nspcc v2.0.4's unexported distance.
// Formula: https://github.com/nspcc-dev/hrw/blob/v2.0.4/hrw.go
func nspccDistance(node, key uint64) uint64 {
	x := node ^ key
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	x *= 0xc4ceb9fe1a85ec53
	return x ^ (x >> 33)
}

// PairOwner is a straightforward HRW reference with XXH64(key || node), using
// fixed-width little-endian encoding. It is not a third-party HRW library.
func PairOwner(nodes []uint64, key Key) uint64 {
	var buf [16]byte
	copy(buf[:8], key.Bytes[:])
	owner, best := nodes[0], uint64(0)
	for _, node := range nodes {
		binary.LittleEndian.PutUint64(buf[8:], node)
		score := xxhash.Sum64(buf[:])
		if score > best || score == best && node < owner {
			owner, best = node, score
		}
	}
	return owner
}
