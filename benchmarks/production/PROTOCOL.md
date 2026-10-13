# Production Ring audit, 2026-10-13

Measure the current working-tree public `vrh.New`, `Ring.Owner`,
`Ring.OwnerBatch`, and `Ring.Nodes`, not the archived PreparedN3 prototype.
Save source SHA-256 hashes because the working tree contains uncommitted work.

Comparators: go-rendezvous + XXH64 public Lookup; a local NSPCC v2.0.4
single-owner scan verified against upstream Sort; direct XXH64(key||node)
reference. NSPCC scan and pair scan are explicitly local adapters.
All use identical unique uint64 IDs and key values. Little-endian byte/string
representations and node preparation are outside Owner timing. Query hashing
and decoding the winning ID are inside. No string-to-uint64 application hash
is measured. External package versions are pinned in go.mod/go.sum.

Long-stream matrix: sizes 1,2,3,7,15,16,17,31,32,63,64,65,256,1024,4096,16384;
sequential and scattered IDs; ChaCha8 random, sequential, and high-32-bit
varying keys. Also six known triples (v1, v2, Murmur, R2, R3, adaptive N3)
with all three key streams. Total 114 cases. Independent deterministic seed
per case, 262144 keys through 64 nodes, 65536 through 1024, 16384 above.
Five full timed passes, rotated implementation order; warmup 4096 keys;
one separate quality pass; GC disabled only during timed passes. Report
median plus min/max of pass means, not per-request latency percentiles.
Use a single pinned logical CPU, GOMAXPROCS=1. Do not run other audits
concurrently with timing. CPU frequency is not locked.

Quality: save every node count, max absolute error in percentage points,
max binomial Z, Pearson statistic and CV / expected noise CV. Thresholds:
maxZ <= 7, chi-square <= N-1+8*sqrt(2*(N-1)); singleton handled exactly.
These are conservative regression gates, not calibrated omnibus p-values.
At 4096/16384 nodes this matrix cannot resolve small relative imbalance.
Add a separate independent random-key holdout with 1048576 keys on each
known triple and on a scattered 16384-node ring (64 observations/node).
Record failures honestly; do not tune the algorithm or omit failed cases.

API benchmarks: five samples of at least 100ms each; report medians and
allocation bytes/counts. Build at all sizes for every comparator (each
constructor's own validation semantics; VRH checks duplicates). Batch lengths
0,1,16,64,1024 at rings 3,16,17,256,16384 with reused, nil, undersized and
in-place destination; compare reusable/nil batches with loops over comparator
Owner. Short dst includes allocation by the API. In-place performance uses
previous output as subsequent input and is reported separately, never as an
equivalent random-key comparison. Backend Auto/Scalar and OwnerScalar reference
at 3,16,17,64,1024,16384. Rebuild after one node replacement plus 1,64,4096
queries at 3,17,256,16384, all comparators. Nodes copy allocations and metadata
methods separately. Shared immutable ring parallel Owner at 16,256,16384 with
GOMAXPROCS 1,2,4,6 and six physical cores available; report aggregate throughput,
not individual request latency. Standard GC stays enabled in API benchmarks.

Correctness: empty/duplicate/invalid backend inputs, Auto/Scalar equivalence,
AVX2 rejection, boundary uint64s, all sizes 1..257 plus large rings, OwnerScalar
reference, batch buffer modes and aliasing contract, input/Nodes isolation,
order independence, add/remove stability, concurrent immutable readers,
NSPCC agreement with upstream Sort. Run normal tests, race tests and vet for
production and audit modules. Portable implementation: run 386 tests if host
supports it, and cross-compile arm64 packages; compilation is not execution.

Keep raw JSONL, benchmark text, environment, validation logs and machine-readable
summary in docs/results/production-2026-10-13. Generate a complete report and
a concise comparison appendix for the post from those saved results.
