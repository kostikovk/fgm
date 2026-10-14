# Distributed-System Patterns and Algorithms — Problem → Solution Map

Use this file when the question is "what is the standard solution to this scaling / resiliency / coordination problem?" It maps a problem shape to a named pattern or algorithm, then points to where the deeper trade-off lives.

> **How to use:** Find the problem row, name the pattern, then cross-check `architecture-selector.md` (for style-level consequences) or `data-structures.md` (for the in-process structure it implies). Cite the book.

## Contents

**Patterns**
- Interaction and edge (API Gateway, BFF, Sidecar, Ambassador)
- Data and transactions (Saga, Outbox, CQRS, Event Sourcing, Sharding, Replication)
- Caching (cache-aside, read/write-through, write-behind, stampede, invalidation)
- Concurrency control (optimistic/pessimistic locking, MVCC, distributed locks, leader election, idempotency keys)
- Messaging semantics (delivery guarantees, idempotent consumer, competing consumers, ordering, CDC)
- Resiliency (circuit breaker, bulkhead, retry, backpressure, DLQ, timeouts, load shedding, health checks)
- Migration and deployment (strangler fig, canary, expand/contract, feature flags, parallel run, tolerant reader)
- Edge and local-first (local-first, CRDTs)

**Algorithms**
- Distributing data (consistent hashing, Merkle trees, quorum)
- Approximate / memory-bounded (Bloom filter, Count-Min Sketch, HyperLogLog)
- Traffic control (token/leaky bucket, fixed/sliding window, distributed counters)
- Load balancing (round robin, least connections, P2C)
- Identifier generation (UUIDv7/ULID, UUIDv4, Snowflake, sequences)
- Ordering and consensus (Raft, Lamport/vector clocks)
- Search and graphs

---

## 1. Patterns

### Interaction and edge

| Problem | Pattern | Notes | Source |
|---|---|---|---|
| Many clients, one entry point for auth / rate limiting / TLS / routing | **API Gateway** | Keep it thin; business logic does not belong here | *Building Microservices* ch.5 |
| Different clients (web, mobile, partner) need different shapes of the same data | **Backend for Frontend (BFF)** | One thin gateway per client type | *Building Microservices* ch.14 |
| Cross-cutting infra (logs, metrics, mTLS) duplicated in every service | **Sidecar** | Separate container next to the service | *Fundamentals of Software Architecture* ch.17 |
| Network concerns (retries, circuit breaking) polluting service code | **Ambassador** | A proxy the service talks to as if it were the remote | *Building Microservices* ch.5 |

### Data and transactions

| Problem | Pattern | Notes | Source |
|---|---|---|---|
| Multi-step business process across services, eventual consistency acceptable | **Saga** — orchestration (coordinator) or choreography (events) | Each step is a local transaction with a compensating action. Avoid 2PC: it holds locks across the network while any participant or the coordinator may be down, and most brokers and cloud databases do not support it | *Building Microservices* ch.6; *DDIA* ch.9 (atomic commit) |
| DB change and event-publication intent must commit atomically | **Transactional Outbox** | Store the event/intent in the same DB transaction as the business change; actual broker delivery is asynchronous. A retrying relay may publish duplicates, so consumers must deduplicate | *Implementing DDD* ch.8 |
| Read shape diverges sharply from write shape | **CQRS** | Separate command/query models; a shared transactional store can provide strong consistency, while asynchronously updated projections introduce lag. Does not require Event Sourcing | *Implementing DDD* ch.4; *DDIA* ch.11 |
| The authoritative state should be an event history, reconstructable by replay | **Event Sourcing** | State = fold over stored events; costs projections and schema evolution. Audit alone does not require ES: an audit log or temporal tables may suffice. ES does not require CQRS | *DDIA* ch.11; *Implementing DDD* ch.8 |
| Single DB no longer scales writes | **Sharding** — range, hash, or directory-based | Hash for uniformity, range for range queries, directory for flexibility | *DDIA* ch.6 |
| Same data needed by many services | **Replication** + owned-by-one-service rule | One writer, many readers; never two writers | *DDIA* ch.5 |

### Caching

The first question is not "which cache" but "what staleness can this read tolerate?" Distinguish a **bounded staleness budget** (an explicit maximum age/lag) from **strict freshness** (no stale reads under the required consistency model). A bound needs justified assumptions about source freshness, fill timing, expiry, and failure handling; a TTL setting alone is not proof. For strict freshness, use authoritative reads at the required isolation/consistency level, or a proven cache-coherence protocol covering every writer, reader, and failure path. Escalate location only when needed: in-process → shared cache (Redis/Memcached) → CDN/edge.

| Problem | Pattern | Notes | Source |
|---|---|---|---|
| Read-heavy lookups, some staleness fine | **Cache-aside** (lazy loading) | App checks cache; on miss reads the store and populates. Simple, but stale entries/fills require expiry or invalidation with the race/budget handling below | *System Design Interview* vol.1 ch.1; *DDIA* ch.11 (caches as derived data) |
| Same, but keep loading logic out of the app | **Read-through** | Cache layer/library loads on miss; app sees one interface | *System Design Interview* vol.1 ch.1 |
| Write path should synchronously update both store and cache | **Write-through** | Adds write latency; does not by itself guarantee no stale reads. Concurrent writes/fills, partial failure, other writers, and replica lag require a coherence protocol | *System Design Interview* vol.1 ch.1 |
| Write-heavy, brief loss window acceptable | **Write-behind** (write-back) | Buffer writes in cache, flush asynchronously; fastest writes; data loss if cache dies before flush | *DDIA* ch.3 (write buffering) |
| Hot key expires → hundreds of simultaneous misses | **Stampede protection** — request coalescing (single-flight), jittered TTL, probabilistic early refresh, or a short lock on rebuild | The classic outage after a deploy or cache flush | No canonical chapter; Nishtala et al., "Scaling Memcache at Facebook" (NSDI 2013) is the standard reference |
| Limit cache divergence within the permitted staleness budget | **TTL** and/or **event-driven invalidation** | Outbox/CDC invalidation has an asynchronous window and can race with fills. TTL limits entry residency under its expiry semantics, not necessarily source-data age; neither mechanism alone proves strict freshness or a hard staleness bound | *DDIA* ch.11 (keeping derived data in sync) |
| Repeated lookups for keys that don't exist | **Negative caching** | Cache the "not found" with a short TTL | Same Memcache paper; no canonical chapter |
| Several instances need the same cached view | **Shared cache** (Redis/Memcached) | Move out of process only for this reason; adds a network hop and a dependency | `data-structures.md` worked example |

**Invalidation race:** a miss reads an old DB snapshot; a writer commits and its invalidation deletes the cache key; the delayed miss then repopulates the stale value (*stale-fill race*). Deletion alone does not close this race. For a bounded budget, account for delayed fills, replica/CDC lag, expiry resets, and outages; use version/generation checks or another proven protocol to prevent obsolete fills, and bypass/fail closed when the bound cannot be maintained. Do not promise "stale for at most TTL" while these races are unknown. Strict freshness needs authoritative reads or a coherence protocol with a defined consistency point, not merely faster asynchronous invalidation.

Placement note: the cache is an adapter concern. The domain asks a port for the value; the adapter decides whether it came from memory, Redis, or the database. See `code-placement.md`.

### Concurrency control

| Problem | Pattern | Notes | Source |
|---|---|---|---|
| Concurrent updates to the same row, conflicts rare | **Optimistic locking** — version column / compare-and-set | Reject stale writes and retry; no locks held across user think time | *DDIA* ch.7; *PoEAA* (Optimistic Offline Lock) |
| Conflicts frequent, or the operation must not fail | **Pessimistic locking** — `SELECT … FOR UPDATE` | Short transactions only; order lock acquisition to avoid deadlocks | *DDIA* ch.7; *PoEAA* (Pessimistic Offline Lock) |
| Readers must not block writers | **MVCC** (default in Postgres, InnoDB) | Know the isolation level you actually run; read committed ≠ repeatable read ≠ serializable | *DDIA* ch.7 |
| Invariant spans several rows ("at most N children") | **Serializable isolation**, or lock the parent row, or encode the invariant as a DB constraint | This is *write skew*: the check-then-act race that a domain-layer `if` can never close | *DDIA* ch.7 |
| Coordinate workers across instances and reject effects from stale holders | **Distributed lock / lease** with fencing token | A paused holder may act after its lease expires; re-checking before the effect leaves a race. The target must atomically enforce monotonically increasing fencing tokens with the protected mutation, or use equivalent transactional coordination; a lock alone does not protect external effects | *DDIA* ch.8–9 |
| Exactly one node coordinates | **Leader election** via a consensus store or DB lease | Design for the split-brain window; followers must reject stale leaders | *DDIA* ch.9 |
| The same request may arrive twice | **Idempotency key protocol** | Scope key by tenant + operation; store a request fingerprint, atomically claim execution, and persist status/result for retries. A TTL set alone is insufficient; see the protocol below | *DDIA* ch.12; *Building Microservices* ch.12 (Resiliency) |

**Idempotency protocol:** use a shared durable record keyed by `(tenant, operation, idempotency_key)`, with a fingerprint of the canonical request. A unique constraint/atomic claim arbitrates concurrent attempts; reject reuse with a different fingerprint. Track statuses such as in-progress, succeeded, failed, or unknown, replay the recorded result, and define recovery for abandoned claims rather than blindly re-executing. For local DB effects, commit the claim/dedup record, business effect, and result in the same transaction. For external effects, a local record cannot atomically cover the provider call: reuse a provider-supported idempotency key and reconcile unknown outcomes (for example, a timeout after possible acceptance) before retrying without protection. Set retention to cover the documented retry/redelivery horizon; after expiry the key no longer prevents duplicates, so longer-lived invariants need durable business uniqueness constraints. See `data-structures.md` for the storage choice.

### Messaging semantics

| Problem | Pattern | Notes | Source |
|---|---|---|---|
| Which delivery/effect guarantee to promise | **At-most-once** / **at-least-once** / **effectively-once** | State the boundary: exactly-once effects can be supported within a defined transactional/idempotency protocol, not inferred from broker delivery guarantees for arbitrary external systems. At-least-once delivery + atomically deduplicated effects can provide effectively-once behavior within that scope and retention window | *DDIA* ch.11–12 |
| Consumer receives duplicates | **Idempotent consumer** (Inbox) | Record processed message ids in the same transaction as the side effect; the natural pair to the Outbox | *Implementing DDD* ch.8; *Enterprise Integration Patterns* (Idempotent Receiver) |
| Throughput: many consumers on one stream | **Competing consumers** | Scales out; you lose global ordering — partition by key if order matters within a key | *Enterprise Integration Patterns* |
| One event, many interested parties | **Publish–subscribe** (topic) vs **point-to-point** (queue) | Topic when producers should not know consumers; queue when exactly one worker must act | *Enterprise Integration Patterns*; *Building Microservices* ch.4 |
| Ordering matters | **Partition by entity key** with defined producer/relay sequencing and serial effect application per key | A broker's partition order is not automatically business/commit order; verify its producer, retry, replay, and consumer guarantees. Cross-key ordering needs separate coordination | *DDIA* ch.11 |
| Large payloads on the bus | **Claim Check** | Store the blob, send a reference; keeps brokers fast and cheap | *Enterprise Integration Patterns* |
| Need events from DB changes without touching every write path | **Change Data Capture** (logical replication, Debezium) | Can capture table changes or relay an Outbox; asynchronous delivery may lag. Raw row changes couple consumers to table schema, so translate them when consumers need domain events | *DDIA* ch.11 |
| Request/response over messaging | **Correlation id + reply queue** | Usually a sign a synchronous call was wanted; ask before building it | *Enterprise Integration Patterns* |
| Sync call or async event? | Sync when the caller needs the result now (read-your-writes); async for side effects the caller does not wait for | Async adds a failure mode (lost/duplicated) in exchange for decoupling | *Building Microservices* ch.4 |

**Ordering boundary:** distinguish business version/commit order, broker append/delivery order, and effect completion order. If two outbox workers publish entity versions 2 then 1, partitioning preserves that publication order rather than repairing it. Serialize publication per entity or use a proven version/sequence protocol, including retries and relay recovery; consumers must apply effects in the required sequence with explicit gap, duplicate, and replay handling. One consumer can still complete effects out of order if it starts concurrent work. Check the chosen broker's version-specific configuration; for Kafka, [producer settings](https://kafka.apache.org/40/configuration/producer-configs/) describe ordering implications of retries, idempotence, and in-flight requests. See [Transactional Outbox](https://microservices.io/patterns/data/transactional-outbox.html) for the relay's ordering obligation. Test reversed relay publication and retry/replay paths before promising business ordering.

### Resiliency

| Problem | Pattern | Notes | Source |
|---|---|---|---|
| A failing dependency drags down callers | **Circuit Breaker** | Open after N failures, half-open probe, close on success | *Building Microservices* ch.12 |
| One slow tenant / endpoint starves the others | **Bulkhead** | Isolated pools (threads, connections, queues) per partition | *Building Microservices* ch.12 |
| Transient failures | **Retry with exponential backoff + jitter** | Jitter prevents thundering herds; cap total attempts; only retry idempotent operations | *DDIA* ch.8 |
| Downstream can't keep up | **Backpressure** / bounded queues | Let the producer feel the pressure instead of buffering unboundedly | *DDIA* ch.11 |
| Poison messages block a queue | **Dead-Letter Queue** | Park after N attempts; alert; replay tool | *Building Microservices* ch.6 |
| A hung dependency ties up threads forever | **Timeouts + deadline propagation** | Every outbound call has a timeout; pass the remaining budget downstream (`context` deadline, gRPC deadline) so a chain fails fast as a whole | *Release It!* ch.5; *Building Microservices* ch.12 |
| Overload turns into total collapse | **Load shedding** — reject early when queues exceed a threshold | Better to fail 10% of requests fast than 100% slowly | *Release It!* ch.5 (Shed Load) |
| Dependency down, but a degraded answer is acceptable | **Fallback / graceful degradation** | Serve cached or default data; make the degradation visible in metrics | *Release It!* ch.5 |
| Orchestrator must know when to restart or route away | **Health checks** — liveness vs readiness | Liveness = process alive; readiness = can serve traffic (dependencies reachable). Do not couple liveness to downstream health | *Release It!* ch.10 (Control Plane) |

### Migration and deployment

| Problem | Pattern | Notes | Source |
|---|---|---|---|
| Replace a monolith without a big-bang rewrite | **Strangler Fig** | Route slice by slice to the new system behind a facade | *Building Microservices* ch.3 |
| Deploy with minimal blast radius | **Canary** / **Blue-Green** | Canary = % of traffic; blue-green = full switch with instant rollback | *Software Engineering at Google* ch.24 |
| Schema change under live traffic | **Expand / Migrate / Contract** | Add new column, dual-write, backfill, switch reads, drop old | *Building Microservices* ch.3 (decomposing the database) |
| Ship code without releasing the feature | **Feature flags** | Decouple deploy from release; use as kill switch; delete flags once settled — stale flags are technical debt with a half-life | *Software Engineering at Google* ch.24; *Building Microservices* ch.8 |
| Validate a rewrite against the old path | **Parallel run / shadow traffic** | Mirror only operations whose effects are safe to duplicate. For mutating paths, isolate shadow storage and replace/block payment, messaging, email, and other external effects; only the authoritative path may perform production effects. Compare outputs and serve the authoritative response; discarding the shadow response does not suppress its effects | *Building Microservices* ch.3 |
| Consumers break on every producer change | **Tolerant Reader** | Read only the fields you need; ignore unknown ones; never fail on additions | *Building Microservices* ch.5 |

**Shadow safety is enforced, not just configured in a prompt or inferred from an HTTP method.** A `GET` may still mutate state or trigger a billable operation. Audit all effect paths and prevent shadow credentials/network access from reaching production writers and providers; use isolated dependencies or recorded responses. Shared idempotency keys alone are not a sandbox for every effect. Shadow mutations may be simulated in disposable storage, never silently applied to production. Check that test traffic cannot charge money, publish production events, send mail, or change authoritative data, including on retries and failures.

### Edge and local-first

| Problem | Pattern | Notes | Source |
|---|---|---|---|
| Ultra-low latency, offline tolerance, collaboration | **Local-first** + background sync | Data lives on the client; server is a sync peer | Kleppmann et al., "Local-First Software" (2019) |
| Concurrent edits without a central lock | **CRDTs** | Conflict-free by construction; not free — state size grows | *DDIA* ch.5 (conflict resolution) |

### AI-native patterns

Covered in `ai-engineering.md`: RAG, semantic caching, model routing, multi-agent orchestration, human-in-the-loop boundaries.

---

## 2. Algorithms

When the design problem is "how do I compute / distribute / estimate this efficiently at scale", these are the standard answers.

### Distributing data

| Problem | Algorithm | Why | Source |
|---|---|---|---|
| Spread keys across N nodes, minimal reshuffle when N changes | **Consistent hashing** (with virtual nodes) | Only K/N keys move on membership change | *DDIA* ch.6; *System Design Interview* vol.1 ch.5 |
| Detect which replicas diverged, cheaply | **Merkle trees** | Compare hashes top-down; sync only differing subtrees | *DDIA* ch.5 |
| Tune consistency vs availability in replicated writes | **Quorum (N, W, R)** with W + R > N | Under strict quorums over the same fixed N-replica set, each successful W-write set intersects each R-read set. Intersection alone does not guarantee freshness or linearizability: ordering/version selection, concurrent-write conflicts, partial/failed writes, and membership/failure handling need a protocol. Sloppy quorums do not ensure this intersection | *DDIA* ch.5; ch.9 (linearizability) |

### Approximate / memory-bounded

| Problem | Algorithm | Trade-off | Source |
|---|---|---|---|
| "Is X in the set?" before an expensive lookup | **Bloom filter** | False positives, no deletes (without counting variant) | *DDIA* ch.3 (used by LSM storage engines) |
| Estimate the frequency of a supplied key in an insertion-only stream | **Count-Min Sketch** | Point-frequency estimates may overestimate due to collisions and do not underestimate under the standard nonnegative-update/no-overflow assumptions. The sketch does not retain key identities or enumerate top-K | *System Design Interview* vol.2 |
| Find approximate top-K / heavy-hitter identities in a stream | **Candidate tracker + frequency estimator**, or **Space-Saving / Misra–Gries** | Retain candidate keys and choose an algorithm with a stated error/threshold guarantee; CMS alone cannot discover identities. Candidate retention can miss keys, so do not promise exact top-K without sufficient exact counting/storage | Practice-based algorithm selection; verify the chosen heavy-hitter algorithm's specification and guarantees |
| Count distinct in a huge set | **HyperLogLog** | ~0.8% standard error with 16k registers (~12 KB, the Redis default); ~2% at ~2 KB | *System Design Interview* vol.2 |

### Traffic control

| Problem | Algorithm | Notes | Source |
|---|---|---|---|
| Rate limiting with bursts allowed | **Token bucket** | Tokens refill at rate r; burst up to capacity. The usual default | *System Design Interview* vol.1 ch.4 |
| Rate limiting with smooth output | **Leaky bucket** | Constant drain rate; excess dropped or queued | *System Design Interview* vol.1 ch.4 |
| Simplest possible limiter | **Fixed window counter** | One counter per window; allows 2× burst at window edges | *System Design Interview* vol.1 ch.4 |
| Exact rolling-window limit, moderate volume | **Sliding window log** | Store timestamps; memory O(requests). Exact admission requires atomic prune/count/admit with defined time and window-boundary semantics | *System Design Interview* vol.1 ch.4 |
| Approximate limit, high volume | **Sliding window counter** | Weighted blend of current and previous window; cheap, but assumes the previous window's traffic was uniformly distributed. Bursts can make the estimate substantially undercount | *System Design Interview* vol.1 ch.4 |
| Limit across many instances | Admission state in a **shared store** (e.g., Redis Lua script) | Preserve the chosen algorithm's semantics and atomically coordinate admission and expiry across all paths. Sharing counters or centralizing at a gateway alone does not make an approximate algorithm exact | *System Design Interview* vol.1 ch.4 |
| Process by priority | **Heap / priority queue** | O(log n) insert and pop | *Algorithms* ch.2 |

**Weighted counters are not a hard rolling-window bound.** For limit 100 per 60 seconds, accept 100 requests at `t = -0.001s` (the end of the previous window). At `t = 59.4s`, their estimated contribution is only `100 × (1 - 59.4/60) = 1`, so the counter can admit 99 more while estimating 100. All **199** requests are still in the trailing 60 seconds. Error depends on burst placement and can approach a 2× allowance in this implementation; do not promise small overshoot. This is an illustrative counterexample, not a quoted source guarantee. Use a properly coordinated sliding window log when the requirement is an exact count in every rolling interval.

### Load balancing

| Problem | Algorithm | Notes | Source |
|---|---|---|---|
| Homogeneous backends, stateless | **Round robin** / **weighted round robin** | Default; weights for heterogeneous capacity | *System Design Interview* vol.1 ch.1 |
| Requests vary a lot in cost | **Least connections** / **least response time** | Follows actual load, not request count | *System Design Interview* vol.1 ch.1 |
| Avoid herd on one backend under partial information | **Power of two choices (P2C)** | Pick two at random, send to the less loaded; near-optimal with tiny state | *System Design Interview* vol.1 ch.1 |
| Same client/key should hit the same backend (local cache, session) | **Consistent hashing** (see "Distributing data" above) | Sticky without a central session table | *DDIA* ch.6 |

### Identifier generation

A frequent design question with a small decision table.

| Need | Pick | Notes | Source |
|---|---|---|---|
| Collision-resistant without coordination, roughly time-sorted | **UUIDv7** or **ULID** | 128-bit, timestamp prefix improves index locality; not a global ordering guarantee. Same-millisecond monotonicity and clock rollback handling depend on the generator | [RFC 9562](https://www.rfc-editor.org/rfc/rfc9562) §5.7, §6; [ULID spec](https://github.com/ulid/spec) |
| Unique everywhere, ordering irrelevant | **UUIDv4** | Random; fragments B-tree indexes under heavy insert — prefer v7 for primary keys | *System Design Interview* vol.1 ch.7 |
| Compact 64-bit, time-ordered, very high throughput | **Snowflake-style** (timestamp + node id + sequence) | Needs node-id assignment and clock-skew handling | *System Design Interview* vol.1 ch.7 |
| Single database, simplest | **DB sequence / auto-increment** | Not inherently unique across shards; predictable ids reveal allocation patterns, not exact row counts. PostgreSQL sequences are not gapless: rollback, conflicts, and crashes can leave gaps | *System Design Interview* vol.1 ch.7; [PostgreSQL sequence functions](https://www.postgresql.org/docs/current/functions-sequence.html) |
| Human-facing numbering with a gapless or accountable issuance requirement | **Transactional row-locked counter + durable document in the same transaction** | Increment a counter row (not a PostgreSQL sequence) and persist the numbered document at finalization, with a uniqueness constraint per numbering series; expose the number only after durable commit. Retain a void/cancellation trace; do not delete or reuse issued numbers. Serialize allocation; never use unprotected `MAX(n)+1`. Legal rules vary by jurisdiction and document type | *DDIA* ch.7 (transactions/locking); [PostgreSQL sequence functions](https://www.postgresql.org/docs/current/functions-sequence.html) (non-gapless behavior); verify applicable numbering rules separately |

### Ordering and consensus

| Problem | Algorithm | Notes | Source |
|---|---|---|---|
| Agree on a value across nodes despite failures | **Raft** (or Paxos) | Use a library or managed service; never hand-roll | *DDIA* ch.9 |
| Order events without a global clock | **Lamport timestamps** / **vector clocks** | Scalar Lamport timestamps respect causality, but independent events can have equal counters. A deterministic total comparison order requires a shared tie-break rule such as lexicographic `(counter, unique node ID)`, with counters unique and increasing per node. Even that order cannot distinguish concurrent events from causally related ones and does not provide consensus or ordered delivery. Vector clocks (version vectors) detect concurrency at the cost of O(nodes) per timestamp | *DDIA* ch.9 (Lamport timestamps), ch.5 (version vectors) |

### Search

| Problem | Algorithm / structure | Source |
|---|---|---|
| Full-text search | **Inverted index**; rank with **BM25** / TF-IDF | *DDIA* ch.3 |
| Prefix / autocomplete | **Trie** | *Algorithms* ch.5 |
| Nearest objects on a map | **Geohash**, **quadtree**, or **S2 cells** | *System Design Interview* vol.2 ch.1 |

### Graphs

| Problem | Algorithm | Source |
|---|---|---|
| Shortest path | **Dijkstra** for non-negative weights; **A\*** for non-negative weights and a known goal with an admissible heuristic that never overestimates remaining cost and is zero at the goal. For graph-search, use a consistent heuristic if closed nodes are never reopened, or reopen nodes when a better path is found. An arbitrary heuristic does not guarantee an optimal path | *Algorithms* ch.4 (Dijkstra); A\* has no chapter in the canon |
| Dependency ordering | **Topological sort** | *Algorithms* ch.4 |

---

## Cross-References

- `architecture-selector.md` — the style-level decision that usually precedes picking a pattern.
- `design-patterns.md` — tactical, in-process patterns (Strategy, State, Specification, Aggregate, Anti-Corruption Layer); this file is for patterns that cross a process or network boundary.
- `data-structures.md` — the in-process structure a pattern implies (heap for priority, LRU for caching, trie for prefix).
- `engineering-principles.md` §5 and §7 — idempotency and error handling, which most of these patterns depend on.

Full canonical list lives in `../SKILL.md` ("The Canon").
