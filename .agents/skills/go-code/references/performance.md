# Performance

Identify the measured problem and workload before optimizing. A database service, CLI parser, and CPU-heavy library have different cost models; do not assume I/O always dominates.

## Choose the change from evidence

- For I/O-bound work, count round trips and inspect slow queries or requests. Consider joins, batching, `RETURNING`, pagination, and eliminating redundant fetches. Use execution plans where useful; `EXPLAIN ANALYZE` executes the statement, so choose a safe fixture and heed write side effects.
- For CPU-bound work, profile the representative workload. Replace repeated scans with an appropriate index or map when lookup behavior justifies its construction cost. Move invariant parsing or compilation outside repeated work.
- For memory pressure, inspect retained data and allocations. Stream or chunk large inputs and bound external response sizes. Preallocate when size is known and doing so does not retain excessive memory.
- Overlap independent I/O only with a validated limit and correct cancellation/error handling; see [data-and-io.md](data-and-io.md).
- Consider `strings.Builder` or simple `strconv` conversions on measured hot paths. Keep reflection, pooling, caches, and unsafe code only when their complexity and lifetime rules are justified; do not ban them solely by name.
- A cache requires an explicit validity and invalidation policy. A financial snapshot may need an observation instant, while other contracts differ. Preserve precision and overflow checks when optimizing arithmetic.

## Measure before and after

State the metric, dataset, environment, and acceptable regression risk. Benchmark the operation, keeping setup outside the timed work where appropriate and ensuring results are actually consumed. On Go 1.24+:

```go
func BenchmarkNormalize(b *testing.B) {
    input := []string{" Alpha ", " Beta "}
    for b.Loop() {
        out := normalize(input)
        if len(out) != len(input) {
            b.Fatal("unexpected result length")
        }
    }
}
```

This is a shape example; `normalize` is the target operation, not a provided helper. If it mutates input, arrange fresh inputs or an appropriate reset so later iterations represent the same workload. On older baselines use `b.N` and a suitable result sink.

Run the actual benchmark package, for example `go test -run='^$' -bench=BenchmarkNormalize -benchmem -count=6 ./path/to/package`. Compare repeated before/after runs with the project's existing benchmark tooling (`benchstat` when available); do not install tools solely to follow this example. Profiles (`-cpuprofile`, `-memprofile`, `go tool pprof`) help identify the path that deserves attention.

Change one meaningful factor at a time. Retain improvements supported by measurements beyond noise, then rerun correctness and relevant race tests. Ordering, rounding, API compatibility, and failure behavior remain part of correctness. Stop when the user's target is met; do not optimize unrelated code.

Sources: [Go diagnostics](https://go.dev/doc/diagnostics), [testing benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks), [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat).
