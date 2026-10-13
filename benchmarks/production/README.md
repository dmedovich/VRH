# Current production Ring comparison

This nested module imports the working tree at `../..`. It does not import the
archived N3 prototype. Production code is not changed by the audit.

From the repository root:

```sh
go test ./...
go test -race ./...
go vet ./...
go -C benchmarks/production test ./...
go -C benchmarks/production test -race ./...
go -C benchmarks/production vet ./...
bash benchmarks/production/run.sh
python3 benchmarks/production/summarize.py
```

The run script assumes the audited machine's CPU topology: CPU 4 for serial
tests and physical cores 0–5 for parallel tests. Adjust affinity on other hosts
and use a fresh output directory rather than overwriting a published run.
The complete audit takes several minutes, particularly the independent
million-key comparison at 16,384 nodes. Compilation happens before timing;
no timed audits run concurrently.

See [PROTOCOL.md](PROTOCOL.md) for the frozen matrix, sample counts, timing
boundaries, statistical gates and limitations. Build measurements include
each adapter's owned data and closure. `BenchmarkDirectNew` separately measures
the unwrapped production constructor. Allocated bytes per construction include
temporary validation data, and must not be called retained ring memory.

`OwnerBatch` comparisons call the production API; competitors use loops over
their scalar owner operation. Only the production in-place benchmark feeds
previous owners back as future keys, so its timing is not directly comparable
to the random-key batch benchmarks. The production batch methods are also
checked against scalar owners with fresh immutable input.

Outputs: raw JSONL owner counts/times, Go benchmark text, source hashes,
environment, summary JSON, full Russian report, and a post appendix. The
summary checks sample totals and the number of repeated measurements.

The benchmark does not claim all possible uint64 inputs, every CPU or a
universal minwise guarantee. Long-stream medians are means over whole passes;
parallel numbers describe aggregate throughput, not per-request p95/p99.
