#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
out=docs/results/production-2026-10-13
mkdir -p "$out"
# Compile first, then time only one process at a time.
go -C benchmarks/production build -o /tmp/vrh-production-audit ./cmd/audit
go -C benchmarks/production test -c -o /tmp/vrh-production-bench
date -u +%FT%TZ > "$out/started.txt"
go version > "$out/environment.txt"
lscpu >> "$out/environment.txt"
git rev-parse HEAD >> "$out/environment.txt"
go -C benchmarks/production list -m all > "$out/modules.txt"
python3 - <<'PY'
import hashlib,json,pathlib
root=pathlib.Path('.')
files=[]
for d in ['algorithms','vrh','simd','benchmarks/production']:
    files.extend(p for p in (root/d).rglob('*') if p.is_file() and p.suffix in {'.go','.s','.mod','.sum','.md','.sh','.py'})
json.dump({str(p):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(files)},open('docs/results/production-2026-10-13/sources.json','w'),indent=2)
PY
GOMAXPROCS=1 taskset -c 4 /tmp/vrh-production-audit -mode matrix > "$out/matrix.jsonl" 2> "$out/matrix-progress.txt"
date -u +%FT%TZ > "$out/matrix-done.txt"
GOMAXPROCS=1 taskset -c 4 /tmp/vrh-production-audit -mode holdout > "$out/holdout.jsonl" 2> "$out/holdout-progress.txt"
date -u +%FT%TZ > "$out/holdout-done.txt"
GOMAXPROCS=1 taskset -c 4 /tmp/vrh-production-bench -test.run '^$' -test.bench 'Benchmark(Build|DirectNew|Backend|Batch|Rebuild|Metadata)$' -test.benchmem -test.benchtime 100ms -test.count 5 > "$out/api.txt" 2>&1
date -u +%FT%TZ > "$out/api-done.txt"
taskset -c 0-5 /tmp/vrh-production-bench -test.run '^$' -test.bench '^BenchmarkParallel$' -test.benchmem -test.benchtime 200ms -test.count 5 -test.cpu 1,2,4,6 > "$out/parallel.txt" 2>&1
date -u +%FT%TZ > "$out/completed.txt"
