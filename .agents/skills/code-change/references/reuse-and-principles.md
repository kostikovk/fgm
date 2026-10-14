# Reuse, Extraction, and Implementation Principles

Apply the shared discovery and testing gates in [Code Change](../SKILL.md) before using this guidance.

## Contents

1. [Extract shared knowledge deliberately](#extract-shared-knowledge-deliberately)
2. [Apply principles to the observed problem](#apply-principles-to-the-observed-problem)
3. [Sources](#sources)

## Extract shared knowledge deliberately

Extract when callers share semantics, invariants, ownership, and a reason to change together. Textual similarity alone is insufficient. Keep independent rules separate when unification would require caller-specific flags, unrelated dependencies, or branching for different policies.

Place the shared operation at the narrowest appropriate owner. Keep a business policy in its feature/domain module; put a genuinely general technical operation in a cohesive utility module. Preserve dependency direction and the existing public surface unless changing it is part of the task. Prefer meaningful names over catchall `utils`, `common`, or `helpers` collections.

Inspect whether an existing wrapper enforces a policy, transaction, authorization, observability, or stable contract before removing it. A small body alone does not make a layer redundant.

## Apply principles to the observed problem

- **KISS:** choose the simplest correct contract and implementation. Count complexity across callers, dependencies, state, and operations, not just lines in the edited function. Prefer straightforward control flow and existing conventions.
- **YAGNI:** add capabilities and extension points for current requirements or demonstrated variation. Tests, necessary refactoring, and required reliability are part of fulfilling the current contract.
- **DRY:** maintain one authoritative implementation of a shared rule when its semantics and ownership align. Allow independent implementations for independently evolving behavior; reconsider an abstraction that accumulates special cases.
- **SRP:** group behavior that changes for the same reason and separate independently changing responsibilities. Function size and changed-file counts are signals, not fixed gates.
- **OCP:** use an existing extension seam when it suits real variants. Introduce one when observed variation warrants it; editing a simple function can be the appropriate change.
- **LSP:** preserve the behavior callers rely on when replacing an implementation, including errors, side effects, ordering, and other promised guarantees. Matching a type signature is insufficient.
- **ISP:** expose the capabilities a consumer needs. Split an interface when unrelated consumers are forced to depend on unrelated operations; an interface is not required for every type.
- **DIP:** keep policy independent of concrete external mechanisms where the adopted architecture or a real substitution need requires a boundary. Pass dependencies at the existing composition point; avoid adding layers merely to name the principle.
- **Composition and cohesion:** assemble focused capabilities, keep related state and behavior together, and make ownership explicit. Prefer composition when inheritance would couple independent behavior.
- **Least astonishment:** make names, signatures, defaults, errors, and side effects match the repository's conventions and callers' reasonable expectations. State whether arguments are mutated, resources are acquired, or external I/O occurs when that behavior is not apparent from the contract. A `calculateTotal` operation should not unexpectedly persist an order or send a notification. Deliberate behavior changes require an explicit contract update and affected-caller checks.
- **Command–Query Separation (CQS):** distinguish observing state from changing it. Queries should not unexpectedly mutate caller-visible or domain state; commands should make their effects clear. Practical combined operations such as `pop`, or commands returning an ID or updated record, are valid when their effects are part of the contract. CQS does not require CQRS, separate databases, or a blanket ban on returning values from commands.

When a change reshapes an interface or boundary, use [deep modules and information hiding](../../design-decision/references/engineering-principles.md#2-modularity--composition) and [preconditions, postconditions, and invariants](../../design-decision/references/engineering-principles.md#9-testing--contract-programming). Read the relevant section, not the entire catalog. For performance changes, establish a representative baseline and compare before/after measurements while preserving correctness; use the existing [performance guide](../../go-code/references/performance.md) for Go-specific profiling and benchmark details, or the target stack's tooling for other languages.

For frontend work, inspect existing components/design tokens and ownership of local, derived, and server state before adding another component, hook, or store. Preserve props/events and effect cleanup contracts. For backend work, inspect the owner of business invariants and the actual transaction, retry, authorization, and resource-lifecycle boundaries. Apply checks relevant to the changed behavior rather than requiring every mechanism on every change.

## Sources

The decision protocol and least-astonishment checks encode the user's development conventions, informed by the repository's `design-decision`, `go-code`, and testing skills. [Fowler's YAGNI](https://martinfowler.com/bliki/Yagni.html), [Martin's SRP](https://blog.cleancoder.com/uncle-bob/2014/05/08/SingleReponsibilityPrinciple.html), [Metz's Wrong Abstraction](https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction), and [Fowler's CQS](https://martinfowler.com/bliki/CommandQuerySeparation.html) support the corresponding trade-offs. The proportional SOLID checks are practical guidance, not a claim that those sources define every item above.
