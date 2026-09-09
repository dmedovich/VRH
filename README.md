# VRH

VRH is a scoring function for rendezvous/HRW hashing, built as a composition
of bijective (mod 2⁶⁴) transforms, with an AVX2-vectorized implementation of
the ownership scan.

- Arbitrary, sparse `uint64` node identifiers — no dense/sequential bucket
  numbering required.
- Proven no-collision property: for a fixed key, distinct nodes always
  produce distinct scores (see [`verify/`](./verify)).
- AVX2 backend with automatic fallback to a scalar implementation on
  unsupported hardware.
- Zero allocations per lookup.

VRH does **not** change the asymptotic complexity of rendezvous hashing —
ownership lookup remains O(N). It reduces the constant factor for the case
where node identifiers are arbitrary and dense renumbering is not an option.
If your bucket set is already densely numbered `[0, n-1]`, algorithms like
Jump Consistent Hash or Power Consistent Hash are asymptotically better
fits — see the design writeup linked below for the tradeoff.

## Install

```
go get github.com/dmedovich/VRH
```

Requires Go 1.27.1+. AVX2 acceleration requires `amd64`; other architectures
automatically use the scalar backend with identical results.

## Usage

```go
package main

import (
	"fmt"

	"github.com/dmedovich/VRH/vrh"
)

func main() {
	nodes := []uint64{101, 202, 303, 404}

	ring, err := vrh.New(nodes, vrh.WithBackend(vrh.BackendAuto))
	if err != nil {
		panic(err)
	}

	owner := ring.Owner(12345)
	fmt.Println("owner:", owner)

	// Batch lookup reuses an internal scoring buffer and honors the
	// configured backend (scalar or AVX2).
	keys := []uint64{1, 2, 3, 4, 5}
	owners := ring.OwnerBatch(keys, nil)
	fmt.Println("owners:", owners)
}
```

## Package layout

| Package | Contents |
|---|---|
| [`vrh`](./vrh) | Public API: `Ring`, `New`, `Owner`, `OwnerBatch` |
| [`algorithms`](./algorithms) | `SplitMix64`, `VRHScore` — the scoring primitives |
| [`simd`](./simd) | Backend interface, scalar implementation, AVX2 implementation |
| [`verify`](./verify) | Executable checks of the no-collision property, including adversarial inputs |
| [`metrics`](./metrics) | Optional helpers for measuring balance/collisions/migration of a live ring |

## Design and benchmarks

This repository contains only the library. The design writeup (bijectivity
proof, related work.

## Limitations

- Equal-weight ownership only; weighted rendezvous hashing is not
  implemented.
- Benchmarked on AMD Ryzen 5 2600 only; no cross-architecture performance
  claim is made by this repository.
- Ownership lookup is O(N). This library trades a lower constant for
  support of arbitrary node identifiers — it is not a drop-in replacement
  for O(1)/O(log N) algorithms that assume dense bucket numbering.

## License
see [LICENSE](./LICENSE).