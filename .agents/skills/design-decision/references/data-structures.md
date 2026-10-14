# Data Structures — Picking the Right One

Choosing a data structure is an architectural decision: it sets the asymptotic cost of every operation that happens against your data. Get it right early; it's expensive to change later.

> **Core idea:** Drive the choice from **access patterns**, not from familiarity. "What do I read?", "What do I write?", "How often?", "Concurrently?", "Ordered?" — answer those, then pick.

## Contents

- Decision tree (start here)
- Quick selector by language (Go, Ruby, TypeScript)
- Detailed trade-offs with Big-O (Go)
- Concurrency-specific choices
- Decision process (five steps)
- Common access patterns → structure
- Worked example: bounded recent-items cache per tenant

## Decision Tree (start here)

```
Need to store data?
├─ Single record with named fields?         → Struct / Record / Hash-with-known-keys
├─ Collection?
│   ├─ Order matters?
│   │   ├─ Random access by index?          → Slice / Array / Vec
│   │   ├─ Insert/remove in middle frequently? → Linked list (rare; usually slice + careful indexing wins on cache)
│   │   └─ Need ordered + range queries?    → Sorted slice (binary search) or B-tree (DB indexes)
│   ├─ Lookup by key?
│   │   ├─ Single value per key?            → Map / Hash / Dict
│   │   ├─ Multiple values per key?         → Map of slices (multimap)
│   │   ├─ Need iteration in insertion order?→ Linked-hash (LRU-style), Ruby Hash (insertion-ordered by default)
│   │   ├─ Need eviction (size-bounded)?    → LRU cache (linked-list + map)
│   │   └─ Frequency / approximate?         → Bloom filter (membership), Count-Min Sketch (frequency), HyperLogLog (cardinality)
│   ├─ Membership only (no value)?          → Set (in Go: `map[T]struct{}`)
│   ├─ Priority / "next best"?              → Heap / priority queue
│   ├─ Prefix lookup / autocomplete?        → Trie
│   ├─ Hierarchical (parent/child, file system)? → Tree (n-ary)
│   ├─ Nodes + relationships?               → Graph (adjacency list / matrix)
│   └─ Stream / pipeline?                   → Channel (Go) / Iterator / Stream
└─ Concurrent access?
    ├─ Many readers, few writers?           → RWMutex + map, or sync.Map
    ├─ Producer / consumer?                 → Channel (Go), BlockingQueue (Java), Queue (Ruby)
    ├─ Atomic counter?                      → atomic.Int64 / Atomic*
    ├─ Bounded buffer?                      → Ring buffer (e.g., for log buffering)
    └─ Fan-out / pub-sub?                   → Channel + select, or pub-sub library
```

> **Citation:** *Algorithms* (Sedgewick & Wayne) ch.1.3 (bags, queues, stacks), ch.2.4 (priority queues), ch.3 (symbol tables: hash tables, BSTs), ch.5.2 (tries); *DDIA* ch.3 for storage-engine structures (LSM, B-tree, Bloom filters); *A Philosophy of Software Design* ch.6 (general-purpose vs special-purpose modules). Composite structures such as LRU caches, ring buffers, and Bloom filters are not in Sedgewick — cite *DDIA* ch.3 or say "no canonical chapter".

---

## Quick Selector by Language

### Go

| Need | Pick | Notes |
|---|---|---|
| Ordered list | `[]T` (slice) | Pre-allocate cap with `make([]T, 0, n)` if size known. |
| Key-value | `map[K]V` | Iteration order randomized — don't rely on it. Not concurrent-safe. |
| Concurrent map | `sync.Map` | Optimized for "write once, read many" or disjoint key sets. Not a drop-in replacement. |
| Set | `map[T]struct{}` | `struct{}` value takes zero bytes. |
| Group of fields | `struct` | Pass by pointer when large or mutable. |
| Communication | `chan T` | "Don't communicate by sharing memory; share memory by communicating." |
| Optional value | `*T` or `(T, bool)` | Pointers leak `nil`-deref; `(T, bool)` is safer for small types. |
| Behavior abstraction | `interface` | Define behavior, not data. |
| Stack / queue | `[]T` with `append` / `[1:]` | Or `container/list` for true linked list (rarely beats slice). |
| Priority queue | `container/heap` | Implement the `heap.Interface`. |
| Doubly-linked list | `container/list` | Useful for LRU (combined with `map`). |
| LRU cache | `hashicorp/golang-lru` or hand-rolled (`container/list` + `map`) | Don't reinvent eviction. |
| Ring / bounded buffer | Custom `[N]T` + indexes, or `chan T` with cap | Channels with cap are bounded queues. |
| Bit set | `[]uint64` with bit ops, or `math/big.Int` | For dense small-int sets. |
| Sorted set | Sorted `[]T` + `sort.Search` | No native B-tree in stdlib; use `github.com/google/btree` if needed. |
| Atomic counter | `atomic.Int64` | Cheaper than mutex for single-int. |

### Ruby

| Need | Pick |
|---|---|
| Ordered list | `Array` |
| Key-value | `Hash` (insertion-ordered by default) |
| Set | `Set` (`require 'set'`) |
| Struct-like | `Struct.new(...)` or `Data.define(...)` (3.2+, immutable) |
| Stack / queue | `Array` with `push/pop` or `push/shift`; `Queue` for thread-safe |
| Priority queue | gem `algorithms` or hand-rolled heap |
| LRU | gem `lru_redux` |
| Concurrent map | `Concurrent::Map` from `concurrent-ruby` |

### TypeScript / JS

| Need | Pick |
|---|---|
| Ordered list | `Array` |
| Key-value (any key type, ordered) | `Map` (preferred over plain object) |
| Set | `Set` |
| Object with known shape | `interface` / `type` |
| Concurrent | `Promise.all` / `Promise.allSettled` with `AbortController` for cancellation; for shared mutable state across workers, `SharedArrayBuffer` + `Atomics` |
| Iterators | generators (`function*`) for lazy sequences |

---

## Detailed Trade-Offs (Go)

| Type | Big-O (read / write / search) | When it wins | When it loses |
|---|---|---|---|
| `[]T` (slice) | O(1) / O(1) amortized append / O(n) | Sequential access, cache-friendly, known size | Frequent middle inserts/removes |
| `map[K]V` | O(1) avg / O(1) avg / O(1) avg | Lookup by key dominates | Memory overhead; iteration order randomized; not concurrent |
| `sync.Map` | O(1) | Disjoint keys per goroutine, or read-mostly | Not faster than `RWMutex + map` in mixed workloads |
| Sorted slice + `sort.Search` | O(log n) search / O(n) insert / O(log n) | Mostly-static data; range queries | Frequent inserts |
| Heap (`container/heap`) | O(log n) insert / O(log n) pop / O(1) peek | Top-K, priority queue, scheduling | Random access (it's not sorted) |
| Linked list (`container/list`) | O(1) insert/remove given node / O(n) search | LRU eviction (paired with map) | Cache misses kill performance for large N |
| Trie | O(L) lookup (L = key length) | Autocomplete, prefix routing | Memory-heavy for sparse keys |
| B-tree | O(log n) all ops, range scans | Disk-backed indexes (Postgres) | Higher constant than hash map for point lookup |
| Bloom filter | O(k) check (k = hash count) | Membership check before expensive lookup | False positives; can't delete (without counting Bloom) |
| HyperLogLog | O(1) add | Cardinality of huge sets in fixed memory | Approximate only |
| Ring buffer | O(1) push/pop | Streaming, bounded log buffer | Fixed size — overflows must be designed for |

> **Citation:** *DDIA* ch.3 ("Storage and Retrieval") for B-tree vs LSM; ch.4 ("Encoding and Evolution") for serialization trade-offs.

---

## Concurrency-Specific Choices

| Need | Go primitive | Ruby | JS |
|---|---|---|---|
| Mutual exclusion | `sync.Mutex` | `Mutex` (core class) | Async mutex or serialized queue for critical sections spanning `await`; workers sharing memory need an appropriate `Atomics` protocol |
| Read-heavy + occasional writes | `sync.RWMutex` | `Concurrent::Map` | `Map` for synchronous access; serialize multi-step async mutations where invariants require it |
| Atomic counter | `atomic.Int64` | `Concurrent::AtomicFixnum` | `Atomics.add` (SharedArrayBuffer) |
| Producer/consumer | `chan T` | `Queue` / `SizedQueue` | async iterator + AbortController |
| One-shot signal | closed `chan struct{}` | `Concurrent::Event` | resolved Promise |
| Wait for N to finish, including failed tasks | `sync.WaitGroup` | `Concurrent::CountDownLatch` | `Promise.allSettled`, then explicitly handle rejected results |
| Pool of workers | `chan job` + N goroutines | `Concurrent::ThreadPoolExecutor` | Worker threads |
| Once-only init | `sync.Once` | `Concurrent::Delay` | module-level promise |

**JavaScript is not race-free merely because one event-loop thread runs it.** A synchronous critical section without a yield cannot be interleaved by another task on that thread, but an `await` allows another operation to run between a read/check and a write. Two debits can both observe balance 10, accept 7, then both store 3. Use an async mutex or queue to serialize the whole in-process invariant, with release on every error/cancellation path; neither `Promise.all` nor `Promise.allSettled` provides mutual exclusion. Such a mutex does not coordinate other processes or database clients: use transactions, locks, or constraints in the shared store. `Atomics` addresses shared memory between workers, not arbitrary async I/O.

**Fail-fast aggregation is not an all-tasks completion barrier.** `Promise.all` rejects when the first input rejects; the remaining tasks keep running and are not automatically cancelled. Cleanup after that rejection can race with their resource use. When cleanup requires all tasks to finish, await `Promise.allSettled` and inspect failures before reporting success. Each task's promise must represent the completion of its work; detached work is not covered by either combinator.

> **Citation:** *DDIA* ch.7 (transactions, isolation, race conditions) and ch.8 (failure modes in distributed systems); [MDN execution model](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Execution_model) for event-loop/run-to-completion behavior; [MDN Promise.all](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all) and [Promise.allSettled](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/allSettled) for settlement behavior. The async critical-section example is an application of that model, not a quoted *DDIA* algorithm.

---

## Decision Process (apply on every new component)

1. **Enumerate operations.** What read patterns? What write patterns? Frequency?
2. **Identify the dominant access.** The structure should optimize for it; everything else is acceptable.
3. **Check concurrency.** Can goroutines, threads, workers, or async tasks interleave operations that maintain an invariant, including across `await`? Choose synchronization for that scope; multiple processes require shared-store coordination, not just an in-process lock. Structure-level thread safety alone does not make a multi-step operation atomic.
4. **Pick from the table.** Justify in one sentence in the PR ("Chose `map[id]struct{}` for the dedup set — O(1) membership, no value needed.").
5. **Defer optimization.** A `map[K]V` is often "good enough." Only reach for a heap or trie when you've measured pressure.

> **Citation:** *The Pragmatic Programmer* topic 39 ("Algorithm Speed") — estimate, then measure before optimizing.

---

## Common Access Patterns → Structure

Recurring shapes in backend services and the structure that usually fits. Use these as precedents, not rules; walk the tree if the fit is not obvious.

| Access pattern | Structure | Reasoning |
|---|---|---|
| Per-tenant recent unique items, bounded | Map of tenant → recency list + membership map; shared sorted set per tenant if needed | Enforce capacity per tenant and define which operations refresh recency; one global LRU keyed by tenant + id does not enforce per-tenant quotas |
| "Is this id in the allowed set?" | `map[string]struct{}` / `Set` | O(1) membership, no value needed |
| Top-K by some score over a stream | Heap with size-K cap | Maintain only K elements, O(log K) insert |
| Multi-stage processing pipeline | Channel chain / async iterator per stage | Backpressure and cancellation fall out naturally |
| Ordered rule evaluation, small N | Slice of rules, iterate in order | Order matters; for N < ~100 a slice beats anything fancier on cache locality |
| Static lookup table rebuilt on config change | Immutable map, swapped atomically | Read-heavy; rebuild-and-swap avoids locking on every read |
| Bulk import staging | Slice / array with pre-allocated capacity | Known size from input; avoids regrowth |
| Idempotency keys | Shared durable table keyed by `(tenant, operation, key)` with request fingerprint, status, and result | Atomic claim + result replay, not just set membership. Commit local DB effects with the dedup record/result in the same transaction; external unknown outcomes need provider idempotency/reconciliation. Retain for the retry/redelivery horizon; see `patterns-and-algorithms.md` for the protocol |
| Dedup across a huge stream, approximate OK | Bloom filter in front of the real store | Saves a lookup on the common "definitely new" path |
| Rate limiting per key | Token bucket per key in a map with TTL | O(1) per request; expire idle keys |

---

## Worked Example — "Track the 1,000 most recently seen ids per tenant, evict oldest"

**Requirements:** Up to 1,000 most-recent **unique** ids per tenant; oldest evicted on overflow; many concurrent readers, one writer. Define "seen" as an explicit writer `RecordSeen(id)` operation: recording an existing id refreshes its position, while membership reads do not. Recency follows the serialized record order, not client wall-clock time; if event timestamps are required instead, specify clock/out-of-order handling and a deterministic tie-breaker.

**Walk the tree:**
- Lookup by key? **Yes** — by id (for "have we seen this recently?").
- Need eviction? **Yes** — bounded size.
- Order matters? **Yes** — eviction is by recency.

→ **Recency list + membership map**, one per tenant (the LRU composition). In Go, `hashicorp/golang-lru` (`lru.New(1000)`) can supply this with `Add` on the writer path and `Contains`/`Peek` on the reader path; in Ruby/Node, verify equivalent semantics in `lru_redux`/`lru-cache`.

**Clarify refresh semantics:** a classic LRU `Get` refreshes recency, so those reads mutate the structure and require exclusive synchronization. Writer duplicate refresh is independent: `RecordSeen(A)` can move A to the newest position while membership reads remain non-mutating and may use a shared lock. At capacity 2, records **A, B, A, C → A, C** (oldest → newest), not B, C. A plain FIFO ring + membership map is suitable only when duplicates do **not** refresh (then this trace yields B, C); otherwise use the recency list or a proven duplicate-aware algorithm that retains the newest occurrence, ignores obsolete occurrences on eviction, and bounds its auxiliary storage. Do not equate "writer re-sees an id" with "reads must take a write lock".

**Concurrency:** protect writer refresh + overflow eviction as one atomic operation. Reader membership/lookups must observe a consistent structure without refreshing it. The Go library synchronizes its operations; other libraries need their own documented synchronization contract. Per-tenant instances already isolate tenant contention; choose lock strategy from the actual operations, not just the label "LRU".

> **Citation:** *DDIA* ch.11 (caches as derived data); no canonical chapter covers the exact recency-list + membership-map composition. The refresh/uniqueness invariant above is the worked example's design, not a quoted algorithm from *DDIA*.

**When to leave process memory:** if instances must agree on the same recent-id set, use shared state with an explicit per-tenant algorithm. Redis `maxmemory` + `allkeys-lru` is a **global, approximate key-eviction policy**: it neither enforces 1,000 members per tenant nor implements this writer-defined order; it may evict the whole tenant key.

**Exact shared variant:** use a Redis sorted set per tenant, with the **id as the unique member**. In one atomic Lua script/Redis Function, allocate a strictly increasing per-tenant record sequence, `ZADD` that score for the id (refreshing an existing member), compute `excess = ZCARD - 1000`, and, only when excess > 0, `ZREMRANGEBYRANK` the oldest ranks `0` through `excess - 1`. Keep the sequence key and sorted-set key in the same Redis Cluster hash slot and pass both explicitly as script keys. All writers must use this operation; membership reads do not refresh, and reads requiring the current agreed set must use the authoritative primary rather than a lagging replica. Validate key types, arguments, and sequence range before mutation: scripting atomicity prevents interleaving, but does not roll back earlier writes on a runtime error. Use sequence scores within Redis's exact integer range (up to 2^53), and reject overflow rather than silently introduce ties. If timestamp scores are chosen instead, define recency as `(timestamp, id)` with Redis's lexicographic member tie-break, handle late events/clock rollback explicitly, and never move an id backward on refresh. This is exact for the chosen ordering, not necessarily event-arrival order.

If membership is correctness-critical, disable global key eviction (`noeviction`), do not expire required tenant/sequence keys, handle rejected writes, and require persistence/failover guarantees appropriate to the invariant; `noeviction` alone is not durability. If arbitrary misses are harmless and an authoritative store can reconstruct the set, global eviction may be an acceptable cache policy, but it is not the recent-id algorithm.

> **Sources:** [Redis key eviction](https://redis.io/docs/latest/develop/reference/eviction/) documents global key eviction and approximate LRU; [Redis sorted sets](https://redis.io/docs/latest/develop/data-types/sorted-sets/) documents unique members and score/tie ordering; [Redis Lua scripting](https://redis.io/docs/latest/develop/programmability/eval-intro/) documents atomic execution. Atomic record/refresh/trim is this example's composition, not a guarantee of `allkeys-lru`.

**Justification (paste in PR):**
> Chose one recency-list + membership map per tenant over a plain map: ≤ 1k unique ids per tenant, O(1) record/refresh and membership lookup, with writer-only recency refresh. The tenant count/lifecycle needs a separate memory bound. Per-instance state is sufficient here only because a miss falls back to the authoritative DB and affects latency, not correctness; if instances must agree, use shared per-tenant state with atomic refresh + trim, not global eviction.

---

## Cross-References

- `patterns-and-algorithms.md` — when the algorithm picks the structure (Dijkstra needs a priority queue; consistent hashing needs a sorted ring).
- `engineering-principles.md` §1 (ADTs), §5 (mutable vs immutable), §7 (concurrency).
- `code-placement.md` — once the structure is picked, where does the code live?

Full canonical list lives in `../SKILL.md` ("The Canon").
