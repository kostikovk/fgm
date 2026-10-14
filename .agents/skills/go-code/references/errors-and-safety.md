# Errors, Values, Arithmetic, Time, and Resources

## Errors

- Check returned errors. An intentional best-effort operation needs an explicit outcome policy; do not discard an error that can invalidate the result.
- Wrap with useful operation context using `%w` when callers should retain access to the underlying error. Returning `err` unchanged is appropriate when another wrapper adds nothing. Exposing a dependency error through wrapping is an API decision.
- Match conditions with `errors.Is` and extract typed errors with `errors.As`; avoid message matching. On Go 1.26+, `errors.AsType[T](err)` returns `(T, bool)`. Use `var target *ValidationError; errors.As(err, &target)` on older baselines.
- Reuse established sentinels when callers branch on a condition; use typed errors when callers need structured data. Translate storage/vendor conditions at the existing boundary without hiding unrelated technical failures.
- For independent partial failures, preserve per-item status and aggregate errors (`errors.Join`, Go 1.20+, when appropriate). Fail fast when continuing would violate the operation's contract.
- Avoid repeatedly logging an error at each returning layer. The boundary deciding its outcome usually owns logging; redact secrets and separate internal diagnostics from public error text.
- Return errors for invalid external data. Reserve panic for programmer invariants; do not use it as ordinary validation. Recover only at a deliberate isolation boundary with a defined response policy.

## Nil, zero values, and ownership

- A typed nil pointer inside an interface is not a nil interface. Return an actual `nil` interface for success when the contract requires it.
- Represent absence separately from a valid zero when the distinction matters: pointer, nullable type, validity flag, or explicit status. Do not convert a decode or scan failure into an empty successful result.
- Reading a nil map is valid; assigning into it panics. Initialize writable maps.
- Preserve serialization contracts. With `encoding/json`, nil slices/maps encode as `null`, while non-nil empty slices/maps encode as `[]`/`{}`. Do not normalize these globally without checking callers.
- A slice shares its backing array. `slices.Clone` (Go 1.21+) separates the outer slice, but does not deep-copy pointer-bearing elements. A full slice expression `s[:len(s):len(s)]` prevents append from reusing spare capacity; it does **not** prevent mutation of existing elements.
- Map iteration order is unspecified. Sort when output, hashing, or processing order must be deterministic; do not sort merely because a map exists. Prevent concurrent writes and concurrent read/write access with ownership or synchronization.
- Do not copy values containing a used mutex, wait group, or other non-copyable synchronization state.

## Numeric correctness and money

- Choose the representation from the contract. Binary floating point is suitable for approximate measurements; exact monetary amounts require integer minor units or an established exact decimal representation.
- Reuse the project's money operations for currency, overflow, precision, and rounding checks. Do not assume a `Money` type, decimal dependency, currency scale, or rounding mode exists.
- Define where rounding happens and preserve it across aggregation and conversion. Integer division truncates; it is not a substitute for the specified rounding policy.
- Check narrowing conversions and arithmetic overflow where values may exceed the target range. Define how a zero denominator is handled before dividing.
- For decimal libraries, use their documented comparison operations rather than assuming struct equality represents numeric equality.

## Time

- For calculations that depend on an observation instant, pass that instant or reuse an existing clock seam. Ordinary timestamping does not require introducing a clock abstraction everywhere.
- Distinguish elapsed duration from calendar arithmetic. `24*time.Hour` is elapsed time; `AddDate` in the contract's location expresses calendar days across DST. Reuse existing day/month helpers and test relevant boundaries.
- Follow the target's storage and display timezone contract; do not assume a particular timezone or UTC storage without project evidence.
- Use `time.Time.Equal` to compare instants; `==` also compares location and monotonic-clock representation.
- Repeated `time.After` calls allocate timers. With Go 1.23's new timer behavior, unreachable timers can be collected; a blanket claim that they leak is incorrect. Reuse a timer only when its lifecycle and baseline semantics are understood; see [modern-go.md](modern-go.md) for compatibility settings.

## Resources

- Register cleanup after successful acquisition. A defer in a long loop runs at function exit; use a scoped helper or explicit close when each iteration should release immediately.
- Close rows, response bodies, and files on all paths. Check iteration errors and finalization errors that affect correctness, especially buffered output and durable writes.
- Roll back unfinished transactions. Cleanup may need a separately bounded context if the operation's context has already been canceled; follow the driver's API.

Sources: [Go error wrapping](https://go.dev/blog/go1.13-errors), [errors](https://pkg.go.dev/errors), [slices.Clone](https://pkg.go.dev/slices#Clone), [time](https://pkg.go.dev/time), [Go 1.23 timers](https://go.dev/wiki/Go123Timer).
