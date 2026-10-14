# Engineering Principles — Rules That Change a Decision

A reference for the principles that alter a design recommendation. Each entry states the rule, the situation where it decides something, and a citation. Basic definitions the agent already knows (what a pure function is, what covariance means) are left out; only the consequence for a design choice is kept.

> **How to use this file:** Find the section, apply the rule, cite the book in the PR or ADR. Section numbers are stable because other reference files point at them. Examples use Go, Ruby, and TypeScript; translate to the project's language.

## Contents

1. [Type Systems & Polymorphism](#1-type-systems--polymorphism)
2. [Modularity & Composition](#2-modularity--composition)
3. [Functional Foundations](#3-functional-foundations)
4. [Architectural Boundaries](#4-architectural-boundaries)
5. [State, Data & Idempotency](#5-state-data--idempotency)
6. [Control Flow & Generics](#6-control-flow--generics)
7. [Error Handling & Concurrency](#7-error-handling--concurrency)
8. [Naming, Refactoring & Code Review](#8-naming-refactoring--code-review)
9. [Testing & Contract Programming](#9-testing--contract-programming)
10. [Metaprogramming, DSLs & Codegen](#10-metaprogramming-dsls--codegen)
11. [Language Semantics](#11-language-semantics)
12. [Framework-Agnostic Engineering](#12-framework-agnostic-engineering)
13. [System vs Applied Code](#13-system-vs-applied-code)
14. [AI-Assisted Engineering](#14-ai-assisted-engineering)

---

## 1. Type Systems & Polymorphism

- **Interfaces at boundaries, concrete types in cores.** Dynamic dispatch costs an indirect call and blocks inlining; it is cheap at a port boundary and expensive in a hot loop. Define the interface where the dependency is *consumed* (Go convention), not where it is implemented. — *Clean Architecture* ch.11 (DIP).
- **Liskov is behavioural, not structural.** A type that satisfies the signature but breaks the caller's expectations (the `Square extends Rectangle` case) is a substitution bug the compiler cannot catch. When a subtype needs to weaken a postcondition, prefer composition. — *Clean Architecture* ch.9.
- **Generics follow shared semantics, not a repetition count.** Use type parameters when supported by the project version and a container or algorithm has the same contract across types; two concrete uses can justify them. Keep distinct business rules separate even when their code looks similar, and prefer concrete types when parameterization adds complexity without useful reuse. — *A Philosophy of Software Design* ch.6.
- **Sum types model state machines.** Rust enums, TypeScript discriminated unions, Go sealed interfaces: make illegal states unrepresentable instead of validating them at runtime. — *Clean Architecture* ch.5–6 (paradigms as constraints).

## 2. Modularity & Composition

- **Deep modules.** A module's public surface should be small relative to what it hides. Many tiny modules with one-line bodies leak complexity to every caller. When deciding whether to split, ask what the *interface* would be, not how many lines the file has. — *A Philosophy of Software Design* ch.4.
- **Composition over inheritance.** Inheritance fixes the relationship at compile time and couples the child to the parent's implementation; hierarchies deeper than two levels are a smell. Use delegation (Go embedding, Ruby forwardable, TS composition). — *Head First Design Patterns* ch.1; *Refactoring* "Replace Subclass with Delegate", "Replace Superclass with Delegate".
- **Pipelines of pure functions** for transformations (`applyRound(applyMargin(applyDiscount(base)))`). Each stage is testable alone and the order is visible. — *Fundamentals of Software Architecture* ch.11 (Pipeline architecture).

## 3. Functional Foundations

- **Functional core, imperative shell.** Keep the domain pure (same input → same output, no I/O) and push side effects to adapters. This is what makes domain tests millisecond-fast and parallel-safe. — *Clean Architecture* ch.6 (Functional Programming); *Software Engineering at Google* ch.11 (test isolation).
- **Declarative hides cost.** A SQL `JOIN` or a list comprehension is compact but may scan millions of rows. When performance is a criterion, look at the plan, not the syntax. — *DDIA* ch.2.
- **Lazy evaluation wins for infinite or expensive streams; loses in tight loops.** Ruby `Enumerator::Lazy`, Python generators, Go channels: use when the consumer may stop early. — *DDIA* ch.11.
- **Code against abstract data types, not representations.** Callers depend on `Set` semantics, not on `map[T]struct{}`; the representation can change later without touching them. — *A Philosophy of Software Design* ch.4.

## 4. Architectural Boundaries

- **One coherent responsibility per module.** Evaluate independent reasons to change, duplicated decisions, leaked internals, and dependency direction. A feature can legitimately touch transport, application logic, persistence, contracts, and tests; the number of changed files does not establish an SRP violation. — *Clean Architecture* ch.7 (SRP).
- **Coupling ranks worst → best: content, common, control, stamp, data.** Prefer passing data over sharing globals or passing control flags. — *Fundamentals of Software Architecture* ch.3.
- **Law of Demeter.** `a.b.c.do()` hard-codes another object's structure into yours. Exceptions: DTOs and fluent builders. — *The Pragmatic Programmer* topic 28 (Decoupling).
- **Dependencies point inward.** The domain never imports an adapter; extract an interface in the domain and implement it outside. — *Clean Architecture* ch.17–22; Cockburn 2005.
- **Inject dependencies at the composition root.** Service locators and singletons hide the dependency graph and defeat test doubles. — *Clean Architecture* ch.11; *Software Engineering at Google* ch.21.

## 5. State, Data & Idempotency

- **Expose hidden state.** Globals, `time.Now()`, random, and environment reads make a function untestable; inject a clock, a source of randomness, a config struct. — *A Philosophy of Software Design* ch.5.
- **Immutable at boundaries, mutable inside when it pays.** DTOs, events, and value objects are immutable; hot-path buffers may mutate. — *DDIA* ch.11 (immutable event logs).
- **Retries need an explicit duplicate-effect strategy.** An operation may be naturally idempotent or protected by business uniqueness constraints; otherwise use an idempotency protocol, not just a key's presence in a TTL set. HTTP method semantics alone do not protect incidental effects such as charging or sending messages. Scope keys, atomically arbitrate concurrent attempts, retain/replay results, and reconcile unknown external outcomes; see `patterns-and-algorithms.md` (Concurrency control). — *DDIA* ch.8, ch.12; *Building Microservices* ch.12.

## 6. Control Flow & Generics

- **Recursion for trees and divide-and-conquer; loops for sequences.** Go, Java, Python, and JavaScript do not guarantee tail-call optimisation, so deep recursion overflows the stack — convert to a loop with an explicit stack for unbounded depth. — *Clean Architecture* ch.4.
- **Evaluate generic reuse by contract and complexity.** Apply §1; no fixed number of repetitions establishes or rules out a useful abstraction. — *A Philosophy of Software Design* ch.6.

## 7. Error Handling & Concurrency

- **Errors as values vs exceptions is a language decision; error *translation* is an architecture decision.** Translate at every boundary: `pgx.ErrNoRows` → `domain.ErrNotFound` in the adapter; `domain.ErrInvalidPrice` → HTTP 422 in the handler. Never let a driver error reach a controller or a status code reach the domain. — *Clean Architecture* ch.22; *The Pragmatic Programmer* topic 24 (Dead Programs Tell No Lies).
- **Typed errors over string matching.** Go `errors.Is/As` on sentinels or typed errors; TS `instanceof` on error classes; Ruby exception hierarchies. — *The Pragmatic Programmer* topic 24.
- **Railway-oriented style** (Result / Either chains) suits pipelines where most steps can fail; exceptions suit deep call stacks with one recovery point.
- **Concurrency model follows the runtime:** goroutines + channels (CSP), async/await (JS, Python, Rust), threads + locks (JVM, Ruby with the GVL caveat), actors (Erlang, Akka). The cheapest fix for races is fewer shared mutable things, not more locks. — *DDIA* ch.7 (isolation), ch.8 (failure modes).
- **Check-then-act is a race.** Any "count, compare, insert" across two statements needs a transaction with the right isolation, a lock on the parent row, or a database constraint. See `patterns-and-algorithms.md` (Concurrency control). — *DDIA* ch.7 (write skew).

## 8. Naming, Refactoring & Code Review

- **Use the ubiquitous language.** If the business says "promotion", the type is `Promotion`, not `DiscountRule`. `Manager`, `Helper`, `Utils` suffixes mean the abstraction has not been found. — *Clean Code* ch.2; *Domain-Driven Design* ch.2.
- **Refactor only under green tests, and never while adding a feature** (Fowler's two hats). Use the catalog's names so reviewers share vocabulary: Extract Function, Move Function, Replace Conditional with Polymorphism, Replace Type Code with Subclasses, Introduce Special Case. — *Refactoring*, 2nd ed., ch.1–3 and the catalog.
- **Review small diffs.** Defect detection drops sharply past a few hundred changed lines. The reviewer checks correctness, then naming, then placement; style is the linter's job. — *Software Engineering at Google* ch.9.

## 9. Testing & Contract Programming

- **Test pyramid tied to layers.** Domain: unit, no I/O, milliseconds. Application: unit with in-memory fakes for ports. Adapters: integration against the real technology. Ports: handler tests with a stubbed use case. A handful of end-to-end tests for critical paths only. — *Software Engineering at Google* ch.11–14.
- **Coverage is a means.** Full coverage of getters is worth less than partial coverage of the core rules. — *Software Engineering at Google* ch.11.
- **Design by contract.** State preconditions, postconditions, and invariants at boundaries so callers know what is safe to pass without reading the implementation; assert them in tests (and in debug builds where the language supports it). — *The Pragmatic Programmer* topic 23.

## 10. Metaprogramming, DSLs & Codegen

- **Generate code from specs at build time** (OpenAPI → client, proto → structs, sqlc → typed queries). Single source of truth, no runtime cost; the price is build complexity and generated code in diffs. — *The Pragmatic Programmer* topic 30 (Transforming Programming); `spec-driven.md`.
- **Runtime metaprogramming** (Ruby `define_method`, JS Proxy, reflection) is worth it for ORMs, RPC frameworks, and fixtures; rarely in business code, because each use is a place static analysis stops working. — *The Pragmatic Programmer* topic 32 (Configuration).
- **A DSL is an architectural choice**, not a syntax choice: internal DSLs (Rake, Kotlin builders) are cheap to add and hard to tool; external DSLs (SQL, OpenAPI) need a parser and pay back only when non-programmers write them. — *The Pragmatic Programmer* topic 14 (Domain Languages).

## 11. Language Semantics

- **Know the memory model** before sharing state across goroutines/threads/workers: Go happens-before, the JVM memory model, JavaScript's single-threaded event loop with `SharedArrayBuffer` as the exception. Bugs from violating it look like code that "should work". — *DDIA* ch.7.
- **Equality and evaluation order differ by language** (Go structural `==` for comparable types, JavaScript coercion, Ruby `==` vs `eql?` vs `equal?`; Go evaluates arguments left to right, C leaves it unspecified). Check before relying on either in a design. — *Software Engineering at Google* ch.8 (language discipline).

## 12. Framework-Agnostic Engineering

- **Frameworks are details.** The domain should compile and test without Echo, Gin, Rails, NestJS, Django, or Spring. Test: delete the framework directory — does the domain package still build? Can a newcomer tell what the system *does* from the top-level directories rather than which framework it uses? — *Clean Architecture* ch.21 (Screaming Architecture); Cockburn 2005.
- **When a framework insists on owning the models** (ActiveRecord, Django ORM), keep business rules in plain objects or service modules that the framework models delegate to.

## 13. System vs Applied Code

The principle that anchors `code-placement.md`.

- **System code answers *how*:** DB connections, brokers, HTTP servers, retry/backoff, observability.
- **Applied code answers *what*:** pricing rules, eligibility, order state transitions — anything a domain expert would recognise.
- **Test:** can a new engineer explain the *business* from this file without knowing the stack? Then it is applied. Do they need to know about connection pooling? Then it is system.
- **In a hexagonal layout** (names vary; see `code-placement.md`): `domain/` and `app/` are applied; `adapters/` and `ports/` are system. Applied code never imports system code.
- — *Clean Architecture* ch.20 (Business Rules); *Implementing DDD* ch.4.

## 14. AI-Assisted Engineering

Newer than the canon; treat as practice, not doctrine.

- **The prompt is a spec.** Version it, review it, eval it before tuning it. — `spec-driven.md` §4.
- **Context is the program.** What goes into the context window (retrieved documents, tool results, memory) determines behaviour more than the model choice. — `ai-engineering.md` §3.
- **Irreversible actions go behind a human gate**, the same way irreversible side effects go behind an explicit boundary in any architecture. — `ai-engineering.md` §1.
- **Generated code gets the same review as human code.** Cheap generation raises the cost of trusting bad code; lint, test, and review it. — *Software Engineering at Google* ch.9 (Code Review).

---

## Cross-References

- **Code placement decisions:** `code-placement.md`
- **Architecture-style decisions:** `architecture-selector.md`
- **Data-structure decisions:** `data-structures.md`
- **Tactical patterns:** `design-patterns.md`
- **Pattern catalog and algorithm map:** `patterns-and-algorithms.md`
- **Spec-driven development:** `spec-driven.md`
- **AI-native engineering:** `ai-engineering.md`
- **Recording a decision:** `../assets/adr-template.md`

Full canonical list lives in `../SKILL.md` ("The Canon").
