# Design Sources

Consult when selecting a source; the [citation-verification gate](../SKILL.md#6-cite) remains mandatory. This bibliography does not verify individual chapter attributions.

## The Canon

Cite books by short form. These are the primary sources for general principles, not an exhaustive or independently verified bibliography. Use official specifications, documentation, or papers for details outside their scope, and follow the citation-verification rule above.

**Architecture and design**
- *Clean Architecture* — Robert C. Martin (SOLID, boundaries, business rules)
- *Fundamentals of Software Architecture* — Mark Richards & Neal Ford (styles, characteristics, trade-offs)
- *Building Microservices*, 2nd ed. — Sam Newman (when and how to split)
- *Hexagonal Architecture* (2005 paper) — Alistair Cockburn (ports and adapters, the original)
- *Patterns of Enterprise Application Architecture* — Martin Fowler (Repository, Service Layer, Unit of Work, locking)
- *Head First Software Architecture* — Richards, Ford, Sadalage
- *Release It!*, 2nd ed. — Michael Nygard (stability patterns: timeouts, load shedding, health checks)
- *Enterprise Integration Patterns* — Hohpe & Woolf (messaging: pub/sub, competing consumers, idempotent receiver, claim check)

**Domain modelling**
- *Domain-Driven Design* — Eric Evans (ubiquitous language, aggregates, bounded contexts)
- *Implementing Domain-Driven Design* — Vaughn Vernon (hexagonal in practice, CQRS, domain events)

**Data and distributed systems**
- *Designing Data-Intensive Applications* (DDIA) — Martin Kleppmann
- *System Design Interview*, Vol. 1 & 2 — Alex Xu (worked large-scale examples)
- *Algorithms*, 4th ed. — Sedgewick & Wayne (the data-structure catalog)

**Code, modules, and craft**
- *A Philosophy of Software Design* — John Ousterhout (deep modules, complexity)
- *Refactoring*, 2nd ed. — Martin Fowler (the named catalog; use 2nd-edition names)
- *The Pragmatic Programmer* — Hunt & Thomas (design by contract, decoupling, codegen)
- *Clean Code* — Robert C. Martin (naming)
- *Design Patterns* — Gamma, Helm, Johnson, Vlissides (the GoF catalog)
- *Head First Design Patterns* — Freeman & Robson (the same catalog, taught by example)
- *Software Engineering at Google* — Winters, Manshreck, Wright (testing, code review, dependency management)
