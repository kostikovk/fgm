# Architecture Style Selector

Given a problem, which architectural choices fit? This guide takes you from problem-shape to a defensible combination in a few minutes.

> **How to use this:** First check what styles the project already uses (see "Project Conventions First" below). Then answer **all four** selector questions in order; this is **not an early-exit tree**. The answers address different axes and can compose, not just compete. Read the relevant style cards, confirm against the criteria, compare alternatives on the same axis or complete combinations in the trade-off table, and cite the book in the ADR or PR.

## Contents

- Project conventions first
- The four selector questions and composability (deployment, read/write models, persistence/history, workflows; internal boundaries are orthogonal)
- Style cards (modular monolith, layered, hexagonal, microservices, CQRS, event sourcing, saga, BFF, microkernel, serverless, edge/local-first)
- Trade-off analysis template
- Sensible defaults when the project has no convention

## Project Conventions First

An architecture recommendation that ignores the existing system is a recommendation to run two architectures at once. Before the selector questions:

1. **Identify the dominant style already in use.** Look at two or three existing services or modules. Are they layered, hexagonal, a Rails-style MVC monolith, event-driven? Note the messaging technology and how cross-service consistency is handled today.
2. **Look for a written default.** ADRs, an architecture guide, or a template repository often state "new services look like X". If one exists, it is the default and the burden of proof is on deviating from it.
3. **Weigh consistency as a criterion.** In the trade-off table below, "consistency with existing services" is a real architectural characteristic (*Fundamentals of Software Architecture* ch.4). A slightly less ideal style that the team already operates well usually beats a textbook-optimal one nobody has run in production here.

If the project is greenfield or has no discernible convention, say so and use the "Sensible Defaults" at the end of this file.

## The Four Selector Questions

The choices below are **composable across independent axes**: Q1 informs deployment topology, Q2 the read/write model, Q3 persistence/history, and Q4 workflow coordination. Internal boundaries (layered or hexagonal) are another axis: choose them based on business complexity, dependency isolation, and project conventions, not read volume. A modular monolith or a microservice can use hexagonal boundaries, CQRS, and either relational persistence or event sourcing. CQRS and event sourcing do not require each other. Evaluate all four questions even when an earlier answer is clear; a "no" still records a decision for that axis.

### Q1 — How many independent teams + how often do they ship?

| Answer | Likely deployment choice | Why |
|---|---|---|
| 1 team, low-frequency releases | **Modular monolith** | Fastest dev loop, single deploy, easy refactoring across boundaries. Don't pay distributed-systems tax for nothing. |
| 2–5 teams, weekly | **Modular monolith with bounded contexts** | Modules with strict boundaries (DDD bounded contexts). Refactor toward services only when a context becomes a deployment-rate or scaling bottleneck. |
| Many teams, daily/hourly | **Microservices** | Independent deployability is the win; you pay for it with operational complexity. |

> **Citation:** *Building Microservices* (Newman, 2nd ed.) ch.1; *Fundamentals of Software Architecture* (Richards & Ford) ch.17 (Microservices) and ch.18 (Choosing the Appropriate Architecture Style).

### Q2 — Do read and write models need to differ, and what is the load?

| Answer | Likely read/write choice |
|---|---|
| One model serves commands and queries well, at any read/write ratio | **Shared model**; layered vs hexagonal boundaries are a separate decision |
| Command and query responsibilities or model shapes diverge, at any read/write ratio | Consider **CQRS** (Command-Query Responsibility Segregation) — separate command/query models or interfaces; a denormalized projection is optional |
| Read volume is high, but the shared model still fits | Caching or **read replicas**, subject to freshness requirements; volume alone does not justify CQRS |

CQRS does not require reads ≫ writes, separate databases, or eventual consistency: both sides can share one database and transactionally consistent data.

> **Citation:** *DDIA* ch.11 ("Stream Processing", derived read models); *Implementing DDD* ch.4 (CQRS).

### Q3 — What must be retained as history, and what is authoritative?

| Answer | Likely persistence/history choice |
|---|---|
| Only current state is required | Standard relational state model behind a repository |
| Audit, regulatory traceability, undo, or point-in-time queries are required, and stored state history suffices | **Relational state + audit/temporal history**, with retention and integrity controls matching the requirement; time-travel alone does not require event sourcing |
| Domain events must be the authoritative record from which state is reconstructed or business history replayed | Consider **Event Sourcing** — state is a fold over authoritative events; justify why audit/temporal history is insufficient |

**Transactional Outbox is orthogonal to the history model**, not an alternative to audit history or event sourcing. When a state change and a publication intent must be recorded together, store them in the same database transaction. This atomically records intent, **not a DB+broker commit**: a relay delivers later, at least once, so consumers need deduplication or idempotency.

> **Citation:** *DDIA* ch.11 (event logs); *Implementing DDD* ch.8 (Domain Events).

### Q4 — Are workflows long-running across multiple services?

| Answer | Likely workflow choice |
|---|---|
| No, single-service requests | Standard sync APIs; any multi-write invariant still needs a proven transaction/coordination boundary |
| Yes, with strict consistency | Redesign the invariant around one proven atomic transaction boundary with the required isolation, often inside one service. If resources cannot share that boundary, assess a supported distributed commit protocol and its availability/recovery costs; a saga does not provide atomic commit |
| Yes, eventual consistency OK | **Saga pattern** — orchestration (a coordinator) or choreography (services react to events) |

**A service boundary is not a transaction boundary.** One service writing to two independent databases can commit the first write and fail before the second; merging deployables does not fix that invariant. Establish which resources commit together, the isolation required for concurrent requests, and the failure/recovery behavior. Distributed transactions (2PC) require participant support, can hold locks while a coordinator is unavailable, and add operational coupling; verify actual support rather than assuming every database or broker participates.

> **Citation:** *Building Microservices* ch.6 ("Workflow": sagas vs distributed transactions); *DDIA* ch.9 (the cost of atomic commit across nodes).

---

## Style Cards

These cards cover different axes, not mutually exclusive alternatives; combine the applicable choices rather than selecting exactly one card.

### Modular Monolith
- **Shape:** one deployable, internally split into modules with explicit boundaries (Go packages, Ruby gems, npm workspaces).
- **Use when:** team < 10 engineers; you don't yet know your service boundaries.
- **Win:** in-process function calls, one deployable, and simpler local transactions when the affected data shares a transactional store.
- **Watch:** discipline collapses without enforcement — internal interfaces become rotten because nothing prevents shortcuts.
- **Books:** *Building Microservices* ch.1 and ch.3 (monolith first; splitting only when justified); *Fundamentals of Software Architecture* ch.3 (Modularity) and ch.18.

### Layered Architecture (N-tier)
- **Shape:** Presentation → Application → Domain → Persistence, top-down dependencies.
- **Use when:** simple CRUD with conventional patterns.
- **Watch:** "anemic domain" — domain layer ends up as data classes only; logic leaks upward to the application layer. Add hexagonal discipline if you want a real domain.
- **Books:** *Patterns of Enterprise Application Architecture* (Fowler); *Fundamentals of Software Architecture* ch.10.

### Hexagonal (Ports & Adapters) / Clean Architecture
- **Shape:** domain at center, ports define what crosses the boundary, adapters implement them. See `code-placement.md`.
- **Use when:** business logic is non-trivial AND you want to swap technologies (DB, message broker) or test without them.
- **Win:** domain is pure; tests are fast; tech swaps are local.
- **Cost:** more files, more interfaces. Pays back at month 3, not week 1.
- **Books:** Cockburn 2005; *Clean Architecture* ch.17–22; *Implementing DDD* ch.4.

### Microservices
- **Shape:** many small services, owned by separate teams, communicating over the network.
- **Use when:** independent deployability or independent scaling matters more than the operational tax.
- **Win:** team autonomy, polyglot tech, blast-radius isolation.
- **Cost:** distributed-systems failure modes (timeouts, partial failure, schema drift), more infra, harder local dev.
- **Books:** *Building Microservices* (Newman) ch.1–3; *Fundamentals of Software Architecture* ch.17; *DDIA* ch.8 (failure modes).

### CQRS
- **Shape:** separate command and query models or interfaces. They may share one database and transactionally consistent data; denormalized projections, separate stores, and asynchronous updates are optional.
- **Use when:** command/query responsibilities or read/write shapes diverge enough to justify separation (reporting, full-text search, leaderboards), regardless of read/write ratio. Neither microservices nor event sourcing is required.
- **Cost:** maintaining separate models; projection complexity and eventual-consistency UX considerations only when asynchronous projections are chosen.
- **Books:** *Implementing DDD* ch.4; *DDIA* ch.11.

### Event Sourcing
- **Shape:** persist domain events as the authoritative record; current state is derived by folding those events. CQRS is optional.
- **Use when:** reconstruction or replay from authoritative domain events is required and worth the complexity. Audit, regulatory traceability, or time-travel alone may be satisfied by relational audit/temporal history.
- **Cost:** event-schema evolution over the retention lifetime, reconstruction logic, snapshots/projections where needed, learning curve.
- **Books:** *DDIA* ch.11; *Implementing DDD* ch.8.

### Saga (Orchestration vs Choreography)
- **Orchestration:** a coordinator service drives the workflow, calls each step, handles compensation.
- **Choreography:** each service reacts to events; the workflow emerges. Easier to start, harder to debug.
- **Use when:** a multi-step business process spans services and needs eventual consistency.
- **Books:** *Building Microservices* ch.6.

### Backend-for-Frontend (BFF)
- **Shape:** a thin gateway per client (web, mobile, partner API), tailored to that client's needs.
- **Use when:** client teams need to evolve fast; underlying services are too generic for any single client.
- **Books:** *Building Microservices* ch.14 (User Interfaces).

### Microkernel / Plugin
- **Shape:** small core + plugins extend it via well-defined extension points.
- **Use when:** the product *is* extensibility (rule engines, IDEs, CMS).
- **Books:** *Fundamentals of Software Architecture* ch.12.

### Serverless / Function-as-a-Service
- **Shape:** stateless functions invoked by events; managed infra.
- **Use when:** spiky traffic, lots of idle time, simple compute. Avoid for tight loops, long warmups, or strict latency floors.
- **Books:** *Building Microservices* ch.8 (Deployment: FaaS trade-offs). No chapter in the canon is dedicated to serverless; say so if asked.

### Edge / Local-First
- **Shape:** compute at the CDN edge or on the client; data syncs in background.
- **Use when:** ultra-low latency, offline tolerance (collaborative editors).
- **Books:** newer territory — Martin Kleppmann's "Local-First Software" essay is the canonical paper.

---

## Trade-Off Analysis Template

When you've narrowed to two alternatives on the same axis or two complete combinations, fill this in (it maps directly onto the "Options Considered" section of `../assets/adr-template.md`). Do not compare composable choices such as "hexagonal vs CQRS" as if only one could apply:

| Criterion | Option A | Option B | Notes |
|---|---|---|---|
| Maintainability | | | |
| Performance (p50/p99 latency, throughput) | | | |
| Scalability (vertical / horizontal) | | | |
| Cost (infra + dev time) | | | |
| Architectural consistency with existing services | | | |
| Operational complexity (deploy, observe, debug) | | | |
| Team familiarity | | | |
| Reversibility (can we change later?) | | | |

End with a one-paragraph **recommendation** naming the deciding factor. Don't hide the call behind hedges — pick one, explain why.

> **Citation for the framework:** *Fundamentals of Software Architecture* ch.4 ("Architecture Characteristics Defined"), ch.5 ("Identifying Architectural Characteristics"), ch.18 ("Choosing the Appropriate Architecture Style"), ch.19 ("Architecture Decisions").

---

## Sensible Defaults (when the project has no convention)

These are the choices that are cheapest to reverse and that most teams under ~10 engineers grow out of gracefully rather than painfully:

- **Deployment and boundaries:** a modular monolith with hexagonal discipline inside each module. Start with one deployable; strict module boundaries make later extraction easier, not mechanical — data ownership, transactions, network failures, and operations still need redesign.
- **Reliable publication tied to a DB write:** Transactional Outbox rather than a DB+broker dual-write. Commit the state change and publication intent in one DB transaction; delivery happens later, at least once, not atomically with the broker. Make consumers idempotent or deduplicate; other remote side effects are not made atomic by the outbox.
- **Read amplification:** a cache (in-process first, then a shared cache like Redis when several instances need the same view) or read replicas, subject to freshness requirements. Introduce CQRS when command/query responsibilities or model shapes justify separation, not merely when read volume grows; a shared DB and synchronous consistency remain options.
- **History and audit:** a normal relational model plus append-only audit or temporal history as required. Consider event sourcing when authoritative domain events are needed for reconstruction/replay and simpler history is insufficient; it is not mandatory even for time-travel if temporal history suffices.
- **Long-running cross-service workflows:** orchestration-style saga with an explicit coordinator; choreography once the team has good tracing.

Deviating from these is fine — the point is that the deviation gets a written reason (an ADR with the trade-off table above), not that it is forbidden.

---

## Cross-References

- `code-placement.md` — once you've chosen internal boundaries, this tells you where the code goes.
- `engineering-principles.md` §4 (Architectural Boundaries), §13 (System vs Applied).
- `patterns-and-algorithms.md` — pattern catalog (Saga, Outbox, Sidecar, Bulkhead, etc.) and scaling algorithms.
- `../assets/adr-template.md` — record the decision.

Full canonical list lives in `../SKILL.md` ("The Canon").
