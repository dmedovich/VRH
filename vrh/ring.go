package vrh

import (
	"errors"
	"fmt"
	"github.com/dmedovich/VRH/algorithms"
	"github.com/dmedovich/VRH/simd"
)

// Ring represents a set of nodes participating in ownership selection.
type Ring struct {
	nodes   []uint64
	backend simd.Backend
}

// New creates a ring using the automatically selected backend.
func New(nodes []uint64, opts ...Option) (*Ring, error) {
	cfg := defaultOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	if len(nodes) == 0 {
		return nil, errors.New("vrh: at least one node is required")
	}

	copied := make([]uint64, len(nodes))
	copy(copied, nodes)

	seen := make(map[uint64]struct{}, len(copied))

	for _, node := range copied {
		if _, exists := seen[node]; exists {
			return nil, fmt.Errorf("vrh: duplicate node ID: %d", node)
		}

		seen[node] = struct{}{}
	}

	backend, err := selectBackend(cfg.Backend)
	if err != nil {
		return nil, err
	}

	return &Ring{
		nodes:   copied,
		backend: backend,
	}, nil
}

// selectBackend resolves the requested backend.
func selectBackend(backend Backend) (simd.Backend, error) {
	switch backend {
	case BackendAuto:
		return simd.Auto(), nil

	case BackendScalar:
		return simd.Scalar(), nil

	case BackendAVX2:
		b := simd.AVX2()

		if !b.Available() {
			return nil, errors.New("vrh: AVX2 backend requested but AVX2 is not available")
		}

		return b, nil

	default:
		return nil, fmt.Errorf("vrh: unknown backend: %d", backend)
	}
}

// Nodes returns a copy of the node IDs in the ring.
func (r *Ring) Nodes() []uint64 {
	nodes := make([]uint64, len(r.nodes))
	copy(nodes, r.nodes)
	return nodes
}

// Len returns the number of nodes in the ring.
func (r *Ring) Len() int {
	return len(r.nodes)
}

// Backend returns the name of the active scoring backend.
func (r *Ring) Backend() string {
	return r.backend.Name()
}

func (r *Ring) Owner(key uint64) uint64 {
	return r.backend.OwnerVRH(key, r.nodes)
}

// OwnerScalar returns the scalar E19 owner.
//
// This method exists primarily for correctness tests and benchmarks.
func (r *Ring) OwnerScalar(key uint64) uint64 {
	owner := r.nodes[0]
	best := algorithms.VRHScore(key, owner)

	for i := 1; i < len(r.nodes); i++ {
		node := r.nodes[i]
		score := algorithms.VRHScore(key, node)

		if score > best {
			best = score
			owner = node
		}
	}

	return owner
}
