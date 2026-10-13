# VRH

VRH is a Go rendezvous hashing library for arbitrary `uint64` node IDs. It
selects the node with the highest N3 score for each `uint64` key. The score
mapping is identified by `algorithms.ScoreVersion == 3`.

N3 is a permutation of the 64-bit node space for each fixed key, so distinct
node IDs never tie. Ownership balance was checked empirically on structured,
random and adversarial node sets; permutation alone does not prove uniform
ownership for every possible set. Lookups scan all nodes in O(N) time.

Changing from a previous VRH score version remaps keys. Every participant in a
cluster must use the same score version. There is no automatic negotiation.

## Install

```sh
go get github.com/dmedovich/VRH
```

Requires Go 1.27.1 or newer. The large-ring N3 owner scan uses native amd64
code; other architectures use portable Go with identical scores and owners.
`BackendAuto` and `BackendScalar` select this scalar implementation. The old
AVX2 implementation used score v2, so `BackendAVX2` is unavailable for N3 and
returns an error from `New`.

## Usage

```go
package main

import (
    "fmt"

    "github.com/dmedovich/VRH/vrh"
)

func main() {
    ring, err := vrh.New([]uint64{101, 202, 303, 404})
    if err != nil {
        panic(err)
    }
    fmt.Println(ring.Owner(12345))

    keys := []uint64{1, 2, 3, 4}
    owners := ring.OwnerBatch(keys, nil)
    fmt.Println(owners)
}
```

`New` rejects empty or duplicate node sets and copies the supplied IDs.
`Nodes` returns a copy. A ring can be queried concurrently after construction.
`Owner` allocates no memory per lookup; `OwnerBatch` also allocates no memory
when its destination has room for all keys. The destination may be the same
slice as the keys.

## Score

`algorithms.VRHScore(key, node)` is the reference score. `VRHKey` prepares a
key once for scoring multiple nodes, and `VRHScoreK` accepts that prepared key.
The node transform and tail are exposed for precomputed scans. `vrh.Ring`
precomputes transformed IDs for rings with more than 16 nodes and uses the
validated native owner scan on amd64. `OwnerScalar` provides a reference scan.
The exact [N3 formula and permutation argument](docs/n3.md) are documented
separately.

The library supports equal weights only. It is not a cryptographic hash or a
defense against chosen inputs. No universal 33/33/33 guarantee is claimed.
The measured performance and distribution are specific to the documented
datasets and machine.

Licensed under [Apache 2.0](LICENSE).
