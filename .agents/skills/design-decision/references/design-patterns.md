# Design Patterns — Symptom → Tactical Pattern

Use this file when the question is "what pattern fits this piece of code?" — a conditional that keeps growing, a class that knows too much, a domain concept that has no home. These are *in-process* patterns (GoF, DDD tactical, PoEAA). Patterns that cross a process or network boundary live in `patterns-and-algorithms.md`.

> **How to use:** Find the symptom you actually see in the code, not the pattern you already like. Apply the "when not to" column honestly — most pattern misuse comes from applying one to a problem that has not appeared yet. Cite the source.

The single most important rule: **a pattern is a refactoring target, not a starting point.** Write the straightforward code, let the second or third variation appear, then name the pattern that the duplication is pointing at. — *Refactoring* (Fowler) ch.1; *A Philosophy of Software Design* ch.6 (general-purpose vs special-purpose).

---

## Contents

1. [1. Choosing behaviour at runtime](#1-choosing-behaviour-at-runtime)
2. [2. Creating objects](#2-creating-objects)
3. [3. Composing structure](#3-composing-structure)
4. [4. Notification and decoupling](#4-notification-and-decoupling)
5. [5. Modelling the domain](#5-modelling-the-domain)
6. [6. Data access](#6-data-access)
7. [Smells that look like patterns](#smells-that-look-like-patterns)
8. [Cross-References](#cross-references)

## 1. Choosing behaviour at runtime

| Symptom | Pattern | What it does | When not to | Source |
|---|---|---|---|---|
| A `switch`/`if` on a type or mode selects one of several algorithms, and the list grows | **Strategy** | One interface, one implementation per algorithm, chosen by the caller or a factory | Two stable branches that will never grow — a plain `if` is clearer | *Head First Design Patterns* ch.1; *Refactoring* "Replace Conditional with Polymorphism" |
| Behaviour depends on *which state the object is in*, and transitions are the interesting part | **State** | Encapsulate state-specific behavior and transitions; validate allowed transitions explicitly. Illegal transitions are impossible only if the particular API/type design enforces that, not because State is used | Three states with trivial transitions — an enum and a transition table are enough; typestate is a separate stronger API constraint | *Head First Design Patterns* ch.10; *Refactoring* "Replace Type Code with Subclasses" + "Replace Conditional with Polymorphism" |
| A request should be offered to alternative handlers until one handles it (e.g., first applicable resolver) | **Chain of Responsibility** | Handlers linked in sequence; each handles or forwards under an explicit dispatch protocol | Every successful request must pass all required checks — use an explicit pipeline/all-required protocol, not first-accept dispatch | *Design Patterns* (GoF) ch.5 |
| Several algorithms share a skeleton but differ in a step or two | **Template Method** (inheritance) or **higher-order function** (composition) | Skeleton calls hook points | Prefer passing functions over subclassing; Template Method locks the hierarchy | *Head First Design Patterns* ch.8 |
| An operation must be queued, logged, undone, or retried as a unit | **Command** | Reify the request as an object with `execute` (and `undo`) | You only ever call it immediately — a function is a command | *Head First Design Patterns* ch.6 |
| A business rule ("eligible for discount", "overdue") is duplicated across queries and checks | **Specification** | Rule as an object with `isSatisfiedBy(x)`; composable with and/or/not; can translate to a query | One-off predicate used in one place | *Domain-Driven Design* (Evans) ch.9 (Specification) |

**Do not conflate dispatch with mandatory checks.** First-handler dispatch stops once a handler takes responsibility. Mandatory validation or approval forwards after each successful check and stops on rejection; accept only after every required check has passed. A middleware chain can implement that all-required protocol, but its forwarding/short-circuit rules must enforce it. For example, manager approval must not bypass required finance approval. Test that earlier success cannot skip a later rejection, and keep each action behind completion of its required checks.

## 2. Creating objects

| Symptom | Pattern | What it does | When not to | Source |
|---|---|---|---|---|
| Construction logic is repeated, or the concrete type depends on runtime input | **Factory Method / Simple Factory** | Centralize the decision of *which* type to create | Only one concrete type exists — call the constructor | *Head First Design Patterns* ch.4 |
| Families of related objects must be created together consistently (e.g., a provider's client, parser, and mapper) | **Abstract Factory** | One factory interface, one implementation per family | Single family; it becomes ceremony | *Design Patterns* (GoF) ch.3 |
| A constructor with many optional parameters, or an object that must be valid at every step of assembly | **Builder** | Step-by-step construction; validate at `build()` | Two or three parameters — use named/keyword arguments or an options struct | *Design Patterns* (GoF) ch.3 |
| Exactly one instance is needed and shared | **Dependency injection of a single instance** — not **Singleton** | Pass the instance in from the composition root | The Singleton pattern hides a global, blocks tests, and couples to the accessor; treat it as a smell | *Head First Design Patterns* ch.5 (and its own caveats); *Clean Architecture* ch.11 |

## 3. Composing structure

| Symptom | Pattern | What it does | When not to | Source |
|---|---|---|---|---|
| A third-party or legacy interface does not match the one your code wants | **Adapter** | Wrap it behind your interface; the only place that knows the foreign shape | The interfaces already match | *Head First Design Patterns* ch.7; Cockburn 2005 (the outbound adapter *is* this pattern) |
| Behaviour must be added in layers without subclass explosion (logging, caching, retry, metrics around a call) | **Decorator** | Wrap the same interface; stack wrappers | Cross-cutting behaviour that a middleware chain or interceptor already provides | *Head First Design Patterns* ch.3 |
| Callers need a simple entry point over a complicated subsystem | **Facade** | One narrow interface hiding many collaborators | The subsystem is already simple; a Facade over two calls is noise | *Head First Design Patterns* ch.7; *A Philosophy of Software Design* ch.4 (deep modules) |
| Tree of parts and wholes treated uniformly (menus, org charts, rule groups) | **Composite** | Leaf and container share an interface | Flat lists | *Head First Design Patterns* ch.9 |
| Control access, lazy-load, or add remote/network behaviour transparently | **Proxy** | Same interface, different lifecycle or location | Decorator when you are adding behaviour rather than controlling access | *Head First Design Patterns* ch.11 |
| Two models of the same concept must coexist (yours and an upstream/legacy system's) | **Anti-Corruption Layer** | Translate at the boundary so foreign concepts never leak into your domain | Upstream model is genuinely the same as yours — then it is just an adapter | *Domain-Driven Design* (Evans) ch.14; *Implementing DDD* ch.3, ch.13 |

## 4. Notification and decoupling

| Symptom | Pattern | What it does | When not to | Source |
|---|---|---|---|---|
| One change must notify many interested parties without the source knowing them | **Observer** (in-process) | Subscribers register; source publishes | One fixed subscriber — call it directly | *Head First Design Patterns* ch.2 |
| A meaningful business fact happened and other parts of the domain must react | **Domain Event** | Named past-tense fact (`OrderPaid`) raised by the aggregate; handlers react; can be published outward via Outbox | Technical notifications (cache invalidation) — those are infrastructure events, not domain events | *Implementing DDD* ch.8 (Domain Events) |
| Callers pass `null` and every consumer checks for it | **Null Object** / **Special Case**, or a **Result / Option type** | An object with neutral behaviour, or a type that forces the caller to handle absence | Absence is genuinely exceptional — raise/return an error instead | *Refactoring* "Introduce Special Case" |

## 5. Modelling the domain

| Symptom | Pattern | What it does | When not to | Source |
|---|---|---|---|---|
| A concept is defined by its attributes, not identity (money, address, date range, quantity) | **Value Object** | Immutable; equality by value; carries its own validation and operations | The thing has a lifecycle and identity — that is an Entity | *Domain-Driven Design* ch.5; *Implementing DDD* ch.6 |
| A concept must be tracked over time regardless of attribute changes | **Entity** | Identity-based equality; mutable state guarded by methods | Attributes fully define it — Value Object | *Domain-Driven Design* ch.5; *Implementing DDD* ch.5 |
| A cluster of objects must change together under one invariant (order + lines + total) | **Aggregate** | One root; all changes go through it; one transaction per aggregate; reference other aggregates by id | Making the aggregate huge to "be safe" — large aggregates serialize all writes and kill throughput | *Domain-Driven Design* ch.6; *Implementing DDD* ch.10 |
| Domain logic that does not belong to any single entity (transfer between two accounts) | **Domain Service** | Stateless operation in the domain layer, expressed in the ubiquitous language | Logic that naturally belongs to one entity — put it there; services are not a dumping ground | *Implementing DDD* ch.7 |
| Loading and saving aggregates is scattered SQL | **Repository** | Collection-like interface per aggregate root, declared in the domain/app layer, implemented in an adapter | CRUD on non-aggregate tables (reporting rows) — a query object is enough | *Domain-Driven Design* ch.6; *Implementing DDD* ch.12; *PoEAA* |
| Application-level orchestration: load, call domain, save, publish | **Application Service / Use Case** | Thin; no business rules; owns the transaction | Putting the rule *inside* the service — that is an anemic domain | *Implementing DDD* ch.14; *Clean Architecture* ch.22 |
| Business rule scattered across a controller, a model callback, and a job | Move it into the domain (Entity, Value Object, or Domain Service) | Rule has one home and one test | — | *Clean Architecture* ch.20; `code-placement.md` |

## 6. Data access

| Symptom | Pattern | What it does | When not to | Source |
|---|---|---|---|---|
| Simple CRUD, model shape ≈ table shape, small team | **Active Record** | Object wraps a row and knows how to persist itself | Rich domain logic — persistence and behaviour tangle; move rules to plain objects the record delegates to | *PoEAA* (Active Record) |
| Domain model must stay free of persistence | **Data Mapper** | A mapper moves data between objects and tables; the domain never sees SQL | Trivial CRUD — the extra layer costs more than it saves | *PoEAA* (Data Mapper); *Clean Architecture* ch.20 |
| Several changes must commit or roll back together | **Unit of Work + explicit transaction boundary** | Track changes and flush/commit at the use-case boundary; tracking changes alone does not provide atomic commit | Reuse an existing ORM Unit of Work/transaction helper only after verifying its actual guarantees. A session is not universally a shared transaction (e.g., GORM `Session` does not begin one); avoid a duplicate wrapper when a suitable mechanism already exists | *PoEAA* (Unit of Work); [GORM transactions](https://gorm.io/docs/transactions.html), [GORM session](https://gorm.io/docs/session.html) |
| The same row loaded twice in one request diverges | **Identity Map** | One in-memory object per row per session | Provided by the ORM; hand-roll only without one | *PoEAA* (Identity Map) |

---

## Smells that look like patterns

| Smell | Why it hurts | Replace with |
|---|---|---|
| **Singleton** as a global access point | Hidden dependency; untestable; order-of-initialization bugs | Inject the single instance |
| **Manager / Helper / Utils** classes | The abstraction was never found; they accrete unrelated code | Name the domain concept, or move the helper next to its one caller |
| **Deep inheritance** (> 2 levels) | Fragile base class; behaviour scattered across the hierarchy | Composition + Strategy/Decorator |
| **Anemic domain model** | Entities are bags of getters; every rule lives in a service | Move rules onto Entities and Value Objects |
| **God object / God aggregate** | One class or aggregate touches everything; every write contends on it | Split by invariant; reference other aggregates by id |
| **Pattern-driven design** ("we need a Factory here") | Ceremony before the second variation exists | Write the direct code; refactor when duplication shows up |

*Sources:* *Clean Code* ch.2 (naming smells); *Refactoring* ch.3 (the smell catalog); *A Philosophy of Software Design* ch.4–6; *Implementing DDD* ch.10 (aggregate sizing).

---

## Cross-References

- `code-placement.md` — which layer the pattern's classes live in (Value Objects and Entities in `domain/`, Repository interface in `domain/` or `app/`, implementation in `adapters/`).
- `patterns-and-algorithms.md` — patterns that cross process boundaries (Outbox, Saga, Circuit Breaker, caching).
- `engineering-principles.md` §2 (composition over inheritance), §8 (refactoring catalog), §13 (System vs Applied).

Full canonical list lives in `../SKILL.md` ("The Canon").
