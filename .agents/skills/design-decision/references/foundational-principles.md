# Foundational Design Principles

Use these criteria after the [Decision Protocol](../SKILL.md#decision-protocol), especially when no catalog method fits.

## Foundational Principles

Five principles recur in every decision. When no tree applies, reason from these.

1. **High cohesion, low coupling.** Modules contain everything related and nothing unrelated. Separate *System* concerns (how: DB, queue, HTTP) from *Applied* concerns (what: business rules). Within an adopted Clean/Hexagonal boundary, Applied code must not import concrete System implementations. Layered and Active Record designs have different dependency conventions: reuse a safe existing service/transaction seam rather than requiring a migration for one feature. Honor explicit project constraints, name any architectural drift, and keep a proposed migration separate from the current recommendation. — *Clean Architecture* ch.7, 17–22; *A Philosophy of Software Design* ch.4.

2. **Type and logic rigor.** Use the type system. Keep the domain core pure and push side effects to the edges (functional core, imperative shell). Define preconditions, postconditions, and invariants at boundaries. — *Clean Architecture* ch.5–6; *The Pragmatic Programmer* topic 23.

3. **Contract integrity.** Identify the authoritative source of the contract: a maintained OpenAPI/AsyncAPI/proto specification, code/schema that generates the API description, or a prompt spec paired with evals. Agree the intended behavior before implementing it and keep derived artifacts aligned with that source. Drift between the supported contract and implementation is a defect. — *Building Microservices* ch.5 (explicit schemas, contract testing); *The Pragmatic Programmer* topic 23; [spec-driven.md](spec-driven.md).

4. **Mechanical sympathy.** Respect the cost model of the layer below: cache lines, network round-trips, dynamic dispatch, serialization. Most performance surprises are ignorance of that layer. — *DDIA* ch.3, 7.

5. **Reversibility.** Prefer the choice that is cheapest to undo. A modular monolith can become services; the reverse is painful. An adapter can be swapped; a domain model entangled with a framework cannot. — *Fundamentals of Software Architecture* ch.4–5; *Building Microservices* ch.1.
