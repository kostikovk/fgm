# Code Placement — Where Should This Code Live?

The single most-asked question in a hexagonal or clean-architecture codebase is: *"Should this go in the domain, the application layer, an adapter, or a transport handler?"* This guide answers it with a decision tree that works regardless of what the project calls those layers.

> **Why this matters:** Misplacement is the cheapest mistake to introduce and the most expensive to fix. A pure business rule that ends up next to the database code becomes untestable without a database. A repository call that leaks into the domain makes the business core depend on a specific vendor. Both look harmless in the PR and hurt for years.

## Contents

- The four layers and how to map their names to the project
- The decision tree (Q1 pure rule → Q2 use case → Q3 driving/inbound transport → Q4 driven/outbound technology)
- Worked examples (A validation, B full slice across layers, C enum converter, D multi-tenant scoping)
- Common anti-patterns
- Quick lookup table

## The Four Layers

Dependencies point **inward**; outer layers know about inner layers, never the reverse.

```
ports/      ← inbound: HTTP / message / gRPC / CLI handlers (system code)
  ↓
app/        ← orchestration / use cases (applied code)
  ↓
domain/     ← entities, value objects, business invariants (pure applied code)
  ↑
adapters/   ← outbound: DB / broker / cache / external-API implementations (system code)
```

`adapters/` and `ports/` both sit at the *outside* of the hexagon. They depend on interfaces declared in `domain/` or `app/`, not the other way around. This is the **Dependency Inversion Principle** in action — *Clean Architecture* ch.11.

### Map the names to the project

This guide uses `domain/`, `app/`, `adapters/`, `ports/` because they are short. Real codebases use many synonyms. Before answering, find the project's equivalents and use *their* names in the recommendation:

| Role in this guide | Common names in the wild |
|---|---|
| `domain/` | `core/`, `model/`, `entities/`, `business/`, `lib/<context>/` (Ruby), `internal/domain/` (Go) |
| `app/` | `application/`, `usecases/`, `services/` (careful: also used for infra), `interactors/`, `commands/` + `queries/` |
| `adapters/` | `infrastructure/`, `infra/`, `persistence/`, `repositories/` (implementations), `gateways/`, `clients/` |
| `ports/` | `api/`, `http/`, `handlers/`, `controllers/`, `transport/`, `delivery/`, `consumers/`, `cmd/` |

A terminology note: in Cockburn's original paper a *port* is the interface and *adapters* sit on both sides of it (inbound and outbound). Many Go and Java codebases instead use `ports/` for inbound handlers and `adapters/` for outbound implementations, which is the convention in this guide. Either is fine. What matters is that the domain has no outward imports; follow the project's naming and say which convention you are using.

If the project has no layered layout at all (a flat `lib/`, a Rails app with fat models), do not invent one for a single feature. Recommend the placement that keeps the business rule free of I/O within the existing structure, and note the refactor that would introduce the boundary if the team wants it.

---

## The Decision Tree

Walk these questions **in order**. The first "yes" is your answer for a single responsibility, not for every piece of code using the same technology. If a handler mixes transport, orchestration, and business rules, split those responsibilities and classify each separately.

### Q1. Is this code a **pure business rule** with no I/O?

Examples: "a price is valid if it is > 0", "a margin is computed as (price − cost) / price", "two products are in the same family if their parent id matches", "an invoice cannot be voided after it is paid".

→ **`domain/`**

A pure rule:
- has no imports of a database driver, HTTP client or server, message broker, cache, file system, `time.Now()` (inject a clock), or any framework.
- can be tested with a plain unit test — no fixtures, no containers, no mocks.
- encodes an *invariant of the business*, not a technology.

> **Citation:** *Clean Architecture* ch.20 ("Business Rules"), *Implementing Domain-Driven Design* (Vernon) ch.4 ("Architecture"), *A Philosophy of Software Design* ch.4 ("Modules Should Be Deep").

### Q2. Does this code **orchestrate domain operations** to fulfill a use case?

Examples: "register a user" (validate password → save to repo → publish event), "apply a discount to an order" (load order → check eligibility → write adjustment → enqueue notification).

→ **`app/`**

A use case:
- depends on `domain/` types and on **port interfaces** (e.g., `UserRepository`), not on concrete adapters.
- coordinates *what* happens, in what order, under which transaction.
- contains no SQL, no HTTP framework code, no message-broker calls — those go through ports.

> **Citation:** *Clean Architecture* ch.22 ("The Clean Architecture") — the "use case" layer; *Implementing DDD* ch.14 ("Application").

### Q3. Does this code **drive the application from an inbound entry point** (HTTP routes, message consumers, gRPC servers, CLI commands, scheduled jobs)?

Examples: an HTTP handler that decodes a JSON body and calls an app-layer use case, a queue consumer that decodes a message and dispatches to a use case, a CLI subcommand, a cron entry point.

→ **`ports/`** — driving / inbound adapters under this guide's naming convention.

An inbound handler:
- decodes the transport (JSON, protobuf, message envelope) → calls into `app/`.
- maps domain errors to transport responses (HTTP 422, message nack, gRPC status, exit code).
- contains no business logic. If you find an `if margin > 0` in a handler, it belongs in `domain/`.

Concrete HTTP or broker technology does not make a driving handler an outbound adapter; direction is determined by who initiates the use case.

> **Citation:** *Clean Architecture* ch.32 ("Frameworks Are Details"), Cockburn 2005.

### Q4. Does this code **implement an outbound port used by the application** to access a database, queue, file system, external API, or other concrete technology?

Examples: a `PostgresUserRepository.Save`, a `KafkaPublisher.Publish`, a `RedisCache.Get`, an object-storage client wrapper, a payment-provider HTTP client.

→ **`adapters/`** — driven / outbound adapters.

An outbound adapter:
- implements an interface that lives in `domain/` or `app/` (the **port**).
- contains SQL, publisher configuration, HTTP-client retries, and other outbound technology details.
- is replaceable: a `PostgresUserRepository` and an `InMemoryUserRepository` should be swappable behind the same interface.

> **Citation:** Cockburn 2005 ("Hexagonal Architecture" paper) — an adapter translates between a port's API and one concrete technology; Cockburn distinguishes primary (driving) adapters from secondary (driven) adapters such as repositories.

---

## Worked Examples

The examples are in Go with a hexagonal layout because the boundaries are visible in the file tree. The reasoning is identical in Ruby, TypeScript, Python, or Java.

### Example A — "Validate that a margin is positive"

A reviewer asks: "Where does `validateMargin(price, cost float64) error` go?"

- It's a pure business rule. No I/O. ✅
- Encodes a domain invariant ("margins are non-negative"). ✅
- Reusable across every use case that touches a price.

→ `domain/pricing/margin.go`

```go
// domain/pricing/margin.go — pure, framework-free.
package pricing

import (
    "errors"
    "math"
)

var ErrNegativeMargin = errors.New("margin must be non-negative")
var ErrInvalidAmount = errors.New("price and cost must be finite")

func ValidateMargin(price, cost float64) error {
    if math.IsNaN(price) || math.IsNaN(cost) || math.IsInf(price, 0) || math.IsInf(cost, 0) {
        return ErrInvalidAmount
    }
    if price < cost {
        return ErrNegativeMargin
    }
    return nil
}
```

The `float64` signature illustrates placement, not an exact monetary representation. Reject non-finite inputs before comparing; direct comparison also avoids overflow in `price-cost`. For actual money, use the project's exact minor-unit or decimal representation with defined currency, range, and rounding rules. Unit-test positive, zero, and negative margins, `NaN` in either argument, infinities, and large finite values.

### Example B — "Save a price override after validation"

The use case: validate the override, then persist it.

- The validation is `domain/`.
- The orchestration ("validate, then save") is `app/`.
- The actual SQL is `adapters/postgres/`.
- The HTTP route that triggers it is `ports/http/`.

```go
// domain/override/override.go
type Override struct {
    ID    string
    Price float64
    Cost  float64
}

func (o Override) Validate() error {
    return pricing.ValidateMargin(o.Price, o.Cost)
}

// app/override/save.go
type OverrideRepo interface { // <- port interface lives near its use case
    Save(ctx context.Context, o domain.Override) error
}

type SaveOverride struct{ repo OverrideRepo }

func (uc *SaveOverride) Execute(ctx context.Context, o domain.Override) error {
    if err := o.Validate(); err != nil {
        return err
    }
    return uc.repo.Save(ctx, o)
}

// adapters/postgres/override_repo.go
type PostgresOverrideRepo struct{ db *gorm.DB }

func (r *PostgresOverrideRepo) Save(ctx context.Context, o domain.Override) error {
    return r.db.WithContext(ctx).Create(&overrideRow{...}).Error
}

// ports/http/override_handler.go
func (h *Handler) PostOverride(w http.ResponseWriter, r *http.Request) {
    var req overrideRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, 400, errorBody{"bad request"}); return
    }

    err := h.uc.Execute(r.Context(), req.toDomain())
    switch {
    case errors.Is(err, pricing.ErrNegativeMargin):
        writeJSON(w, 422, errorBody{"negative margin"})
    case err != nil:
        writeJSON(w, 500, errorBody{"internal"})
    default:
        w.WriteHeader(204)
    }
}
```

Each layer is independently testable:
- `domain/` — pure unit tests, no fixtures.
- `app/` — unit tests with a fake `OverrideRepo` (in-memory map).
- `adapters/postgres/` — integration test against a real Postgres.
- `ports/http/` — handler test with `httptest.NewRecorder`.

### Example C — "Convert a domain enum to a DB string"

A `toDomain` / `fromDomain` converter that maps a domain enum (say `OrderStatus`) to and from its DB string.

- Lives in `adapters/postgres/` — it's an *adapter detail* (the DB representation of a domain concept).
- BUT must enforce the invariant: if the DB value is unrecognized, return a typed error (`ErrInvalidOrderStatus`) — never silently coerce to a default. The database may not enforce the set of valid values (no CHECK constraint, or a migration that widened the column), so the converter is the last line of defense for domain integrity.

### Example D — "Multi-tenant scoping"

Every read or write of tenant-owned data must be scoped to the tenant (`tenant_id`, `account_id`, `org_id` — whatever the project calls it). **Where does that scoping live?**

- The *requirement* is a domain invariant: "tenants are isolated."
- The *enforcement* is an adapter concern: reads and writes include tenant predicates, using a tenant scope authorized against the authenticated principal, not a tenant id trusted merely because it arrived in a request. For child tables without a tenant column, scope through the parent and protect the ownership check through the write.
- The *scoping contract* is expressed in the port interface: `ListByTenant(ctx, tenantID)` — there is no unscoped method to call by accident. This interface does not itself prove authorization or race-free enforcement.

**Ownership check and write must agree on an authorization point.** If parent ownership is enforced as immutable, explicitly name that assumption. If ownership can change, a separate check followed by an insert is a check-then-act race, even inside a normal `READ COMMITTED` transaction. One possible implementation is to lock and verify the parent in the same transaction as the child write, hold the lock through commit, and ensure ownership-transfer paths honor that lock. Another is a tenant-aware schema/constraint protocol that preserves the association under concurrent transfers. A single scoped SQL statement is not automatically enough if its snapshot can race with a transfer; verify the actual locking and constraint semantics. Define what happens to existing children on transfer, not just who may insert them.

→ The interface lives in `app/`; the SQL lives in `adapters/postgres/`. The domain layer should not know SQL exists. Apply defense in depth: the adapter enforces the authorized scope on every operation. Integration-test child writes racing with ownership transfer and unauthorized tenant scopes; sequential mock tests cannot establish this isolation.

---

## Common Anti-Patterns

### "Minimum diff" placement
Putting a function "where it's smallest to add" without considering layer. **This is not a valid placement justification.** A pure business rule belongs in `domain/` even if that means creating a new file.

### Shared `utils/` or `helpers/`
Symptoms of skipped abstraction work. Before adding to `utils/`, ask: is this a domain concept (then it has a name), an adapter detail (then it lives near the adapter), or generic (then it's standard library or a vendored generic helper)?

> *Clean Code* (Martin) ch.2 — `Util`, `Helper`, `Manager` are name smells.

### Logic in handlers
An HTTP handler with `if`/`else` branches over business state. Lift to `app/` (or `domain/` if pure). The handler should look like: decode → call use case → map result.

### Logic in repositories
Business policies or invariants hidden in persistence code are the smell, not every computation performed by SQL. Filtering, `COUNT`, `SUM`, and projections may belong in a repository or query adapter; `CountByTenant` does not require loading all rows into the domain. Keep business decisions in the domain/use case using those query results, and preserve their atomicity requirements. Do not force repositories to expose only CRUD or move efficient query computation into memory merely to satisfy a layer label.

### Cross-package imports from `domain/`
The instant `domain/X` imports `adapters/Y`, you've inverted the dependency rule. Refactor by extracting an interface in `domain/` and implementing it in `adapters/`.

---

## Quick Lookup

| Code | Belongs in |
|---|---|
| Pure business validation, computation, invariant | `domain/` |
| Domain entity, value object, aggregate | `domain/` |
| Use-case / "story" orchestration | `app/` |
| Port interface (defining what the outside world must provide) | `app/` (or `domain/`) |
| Driving / inbound HTTP handler, message consumer, gRPC server, CLI, cron entry point | `ports/<protocol>/` |
| Driven / outbound DB / cache / broker publisher / external-API implementation | `adapters/<tech>/` |
| SQL filtering, aggregation, projection | repository / query implementation in `adapters/<tech>/` |
| Domain ↔ DB row converter | `adapters/<tech>/` |
| Domain ↔ wire format (JSON DTO) | `ports/<protocol>/` |
| Authentication middleware | `ports/http/` |
| Tenant scoping (SQL-level) | `adapters/<db>/` (with the scoped interface declared in `app/`) |
| Cross-cutting logging, tracing | `adapters/observability/` or framework-provided |
| Injected clock, ID generator | interface in `domain/` or `app/`; real implementation in `adapters/` |

---

## Cross-References

- `engineering-principles.md` §13 — System vs Applied code (the underlying principle).
- `architecture-selector.md` — picking the *style* before placing code in *layers*.
- `data-structures.md` — picking the right type once you know the layer.

Full canonical list lives in `../SKILL.md` ("The Canon").
