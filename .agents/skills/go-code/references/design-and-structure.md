# Design and Structure

## Packages and boundaries

Follow the target's documented dependency direction and implemented seams. Do not require `domain/usecase/adapter/storage`, forbid sibling imports, or migrate architecture for an ordinary Go change. An import cycle is a real constraint; resolve responsibility and dependency ownership rather than adding a pass-through abstraction by reflex.

Keep a package cohesive and its exported surface intentional. Prefer names describing a concept over catchall `utils`/`common` packages. File organization follows the repository: methods need not all live in one file, and a useful package may contain one type.

Prefer APIs that hide meaningful complexity. A forwarding layer can still earn its place by enforcing authorization, transactions, observability, or a stable boundary; inspect callers before deleting it.

## Interfaces

- Start with the concrete capability the consumer needs. Introduce an interface for a real substitution seam or abstraction contract, not for every struct.
- Consumer-owned narrow interfaces are a useful default; public shared protocols may appropriately belong elsewhere. Method count is a design signal, not a fixed limit.
- Accept an interface when its abstraction benefits the caller; return a concrete type when callers need its capabilities. Iterator, plugin, and factory APIs can legitimately return interfaces.
- Prefer meaningful operations over a generic `Do` method unless the protocol itself is generic. Do not redesign an established client merely to simplify one mock.
- Add a compile-time satisfaction assertion when the relationship is otherwise easy to break, for example `var _ io.Reader = (*Reader)(nil)`.

## Construction and types

- Pass dependencies explicitly at the existing composition boundary. Avoid hidden mutable clients/configuration and unexpected initialization side effects; retain deliberate registration and immutable lookup tables where appropriate.
- Required values fit ordinary parameters. Use a parameter struct when related inputs become hard to distinguish, and functional options when optional configuration justifies them. Neither a fixed argument limit nor options for every constructor is necessary.
- Document constructors that acquire resources or start background work, including who closes or stops them. Make zero values useful where practical; otherwise validate construction and keep invalid state inaccessible.
- Choose pointer receivers for mutation, identity, large values, and synchronization-bearing types. Use value receivers for suitable small values. Check method sets and interface satisfaction before changing receivers.
- Validate enum inputs; an invalid zero value is useful when absence is invalid, while a meaningful default is valid for other APIs.
- Use keyed struct literals for evolving contracts. Embed a type only when promotion of its methods and fields is part of the intended API.
- JSON tags and omission options are wire-contract choices. `omitzero` requires Go 1.24+; preserve field names and absence semantics rather than changing tags as style cleanup.

## Functions, names, and comments

Keep the normal flow easy to read with guard clauses and cohesive operations. Split a function when responsibilities or independent testable behavior justify it; line counts, parameter counts, and condition operand counts are not correctness gates.

Use MixedCaps and consistent initialisms (`HTTP`, `URL`, `ID`), short package names, and identifiers sized to their scope. Avoid stutter when a clearer existing pattern fits. A boolean parameter or map-versus-switch choice is assessed by readability and semantics, not banned automatically.

Document exported contracts according to the project's policy: units, invariants, errors, ownership, and concurrency safety. Explain non-obvious reasons and external constraints. Remove obsolete commented-out code; do not erase useful tracked TODOs or public documentation merely for brevity. After renaming, inspect callers, serialized names, logs, and documentation.

## Generics and dependencies

Use type parameters when multiple concrete types share an algorithm and constraints express its requirements. Preserve readable concrete code when generic plumbing adds complexity. Do not impose an arbitrary number of duplicate implementations before allowing generics.

Inspect `go.mod` and existing utilities first. Prefer a suitable standard-library operation, but retain chosen frameworks, ORMs, and helper libraries when they serve the target. Adding a module, promoting an indirect requirement, or replacing a dependency is an explicit scoped decision; an indirect requirement is not permission to introduce a new direct import. Consult [modern-go.md](modern-go.md) for API version gates.

Sources: [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), [Go module reference](https://go.dev/ref/mod).
