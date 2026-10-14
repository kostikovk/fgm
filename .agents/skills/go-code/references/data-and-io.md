# Data Access, HTTP, Context, and Concurrency

## Database access

Use the project's existing database library and transaction seam. Do not introduce pgx, replace an ORM, or assume a read-only process for an unrelated task.

- Parameterize values. Dynamic identifiers need a trusted allowlist and the driver's identifier-quoting facility; value placeholders cannot substitute for table/column names.
- Select columns matching the scan contract. Distinguish no rows, nullable values, scan failures, and iteration failures. Close rows and check their final error; collectors may perform these tasks for you.
- Reuse existing mappings from database conditions to application errors. Do not turn every unique violation into success: inspect the relevant constraint and operation's idempotency contract.
- Avoid N+1 queries, unnecessary read-after-write calls, and unbounded result sets. Batch or stream when appropriate. Choose indexes using the query and execution plan, not a universal rule that every multi-column filter needs a composite index.
- Design idempotency keys and conflict handling from the requested operation. An upsert alone does not guarantee correct repeated writes or preservation of user-owned fields.
- Keep transactions short and avoid unrelated network calls inside them. Observe a consistent lock order. Retry the entire transaction on a retryable serialization failure only when the surrounding operation permits safe bounded retries; PostgreSQL `40001` can occur at REPEATABLE READ and SERIALIZABLE.
- Reuse a process-owned connection pool; make sizing and shutdown ownership explicit. Follow the repository's migration workflow, including its rollback policy. Do not invent a migrator command or edit applied migrations as a routine fix.
- If the application contract is read-only, preserve both code and database enforcement; a failing write is not a reason to switch credentials or pools.

### When the target uses pgx v5

`CollectRows`/`CollectOneRow` close rows and return collection errors. `RowToStructByName` requires a matching row shape; choose mapping/nullable types according to the target's schema. Check the pinned pgx version before assuming a helper exists.

`BeginFunc` calls commit when the callback returns nil and rollback on error/panic; transaction-control failures can still be returned. Its context applies to transaction control, not automatically to statements inside the callback. Pass `ctx` to each `tx.Query`/`tx.Exec` too. With manual transactions, arrange rollback on every uncommitted path and understand cleanup behavior after cancellation.

Sources: [database/sql](https://pkg.go.dev/database/sql), [pgx v5](https://pkg.go.dev/github.com/jackc/pgx/v5), [PostgreSQL serialization retries](https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html).

## HTTP and untrusted input

- Reuse an appropriately configured client/transport. Requests carry caller context; the operation has a suitable deadline or documented long-lived lifecycle. The default client has no overall timeout.
- Close response bodies, bound reads where input size is untrusted, and handle status codes before interpreting successful payloads. Reading to EOF can enable connection reuse, but do not drain an unbounded hostile body solely for that purpose.
- Review redirect behavior for credentials, especially custom authentication headers. Restrict redirects or remove credentials at the boundary; do not assume default protections cover every header.
- Redact secrets in URLs, errors, and logs without losing error identity needed internally. Validate payloads instead of returning successful zeroed records after decoding fails.
- Retry transient failures only when operation semantics allow it, with bounded attempts, backoff/jitter, and cancellation. Idempotent writes may be retryable with their explicit guarantees; reads are not automatically safe or appropriate to retry forever.
- Honor applicable `Retry-After` behavior. Pagination needs a clear terminal condition and protection against repeated cursors or unbounded pages. Validate untrusted file sizes and formats before processing.
- Keep TLS verification enabled; insecure test transport settings must remain confined to their test scope.

Source: [net/http](https://pkg.go.dev/net/http).

## Context and lifecycle

Pass `ctx` first to operations that block or perform I/O; do not pass nil or replace caller context halfway through its work. Keep context out of long-lived structs unless the API's documented lifecycle requires it. Create roots at application or test boundaries, including non-`main` entry points where appropriate.

Cancel derived contexts to release resources. Use context values for request metadata, not dependencies or optional arguments. `WithoutCancel` (Go 1.21+) removes cancellation **and the deadline**, leaving `Done` nil; detached work needs a separate bound and an owner that waits for it.

Validate configuration at startup using the project's established source (environment, file, or secret provider). Keep process exit at the top level after cleanup-bearing functions return: `os.Exit` and `log.Fatal` skip defers. Coordinate signal cancellation, server shutdown, resource closing, and background task completion.

For protocol-based stdio programs, reserve stdout for the protocol and send diagnostics to stderr. Follow the target's language and logging conventions. Stable structured messages with contextual attributes help diagnostics; redact secrets and sensitive data rather than assuming amounts or identifiers are always safe to log.

Source: [context](https://pkg.go.dev/context).

## Concurrency

Start sequential when it meets the need. For concurrent work, define the concurrency limit, ownership of results, cancellation behavior, and partial-failure policy before selecting a primitive.

- For Go 1.25+, `WaitGroup.Go` registers a task whose callback must not panic. Do not also call `Done` for that task. When the group is empty, register work before `Wait`; reuse it only after previous waits have returned.
- For older baselines, call `Add` before launching the goroutine and defer `Done` inside it. Before Go 1.22's declared loop-variable semantics, explicitly capture per-iteration values where needed.
- Use a validated positive limit. A semaphore acquisition that can block must observe cancellation. After acquisition, recheck cancellation if the contract forbids starting further work; `select` can choose a ready work case even when cancellation is ready too.
- A wait group waits for completion; it does not carry errors or cancellation. Use per-item outcomes for intentional partial success. Use `errgroup.WithContext` for fail-fast behavior when the dependency already exists or its addition is authorized. `SetLimit` limits active tasks, but its blocked `Go` call is not itself context-aware; check scheduling needs.
- Writing disjoint result indices can avoid a mutex if the slice is not concurrently resized and readers wait for completion. Preserve order when promised; do not expose completion order accidentally.
- Protect shared mutable state and avoid concurrent read/write access to ordinary maps. Only the channel owner coordinating all sends should close it.
- Give potentially blocking channel operations a termination path. Not every select needs `ctx.Done()`; nonblocking selects and deliberate shutdown waits have different contracts.
- Test synchronization through channels or version-compatible `synctest`, with bounded waits. Race tests detect executed races, not all possible interleavings or goroutine leaks.

Sources: [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup), [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup), [testing/synctest](https://pkg.go.dev/testing/synctest).
