---
name: design-decision
description: Project-grounded advisor for technical design decisions. Inspect relevant implementation, tests, and ADRs before recommending code placement, architecture styles, design patterns, data structures, caching, concurrency, messaging, idempotency, error handling, API contracts, or LLM/RAG/agent architecture. Use when the user is weighing design options, asks about layering or separation of concerns, faces a growing if/switch or a service split, reviews a design/spec for structure, or plans a feature whose solution is still open — even without saying "architecture". Prefer existing mechanisms when they meet the requirements; justify new abstractions with concrete code evidence. Produces a named recommendation, project evidence, the deciding trade-off, a source citation, and optionally an Architecture Decision Record. Not for line-level bug review or writing the implementation.
---

# Design Decision

A decision protocol for the moment an engineer asks *"how should I do this?"* It turns a vague design question into a named choice grounded in the project's actual code and constraints, with the canonical literature explaining the principle rather than replacing project evidence.

## Contents

1. [Navigation — choose the decision class](#navigation--choose-the-decision-class)
2. [Decision Protocol — always apply](#decision-protocol)
   - [Inspect implementation](#1-inspect-the-implementation-before-choosing-a-solution)
   - [Clarify deciding ambiguity](#2-clarify-if-the-question-is-ambiguous)
   - [Route](#3-route)
   - [Apply the decision method](#4-apply-the-references-decision-method)
   - [Name the choice](#5-name-the-choice)
   - [Verify citations](#6-cite)
   - [State the deciding factor](#7-justify-with-the-deciding-factor)
   - [Self-check](#8-self-check-before-answering)
3. [Response Format — required fields](#response-format)
4. [Recording the Decision — consequential choices](#recording-the-decision)
5. [Tone](#tone)

## Navigation — choose the decision class

The skill covers three primary decision classes and a set of cross-cutting ones:

| Decision class | Reference | What you get |
|---|---|---|
| Where does this code go? | [code-placement](references/code-placement.md) | 4-question tree → layer name → citation |
| Which architecture choices? | [architecture-selector](references/architecture-selector.md) | 4 selector questions → compatible choices across independent dimensions → trade-off table |
| Which data structure / primitive? | [data-structures](references/data-structures.md) | Access-pattern tree → type → Big-O justification |
| Which design pattern for this code? (growing `switch`, state transitions, duplicated rule, null checks, object construction) | [design-patterns](references/design-patterns.md) | Symptom → tactical pattern → "when not to" → citation |
| Types, errors, concurrency, testing, naming, refactoring, DSLs | [engineering-principles](references/engineering-principles.md) | Sectioned rules, each with a citation |
| API / event contracts, contract testing | [spec-driven](references/spec-driven.md) | Contract design, authoritative source, generation, conformance and consumer compatibility |
| Caching, locking, messaging guarantees, resiliency, migration, rate limiting, id generation, scaling algorithms | [patterns-and-algorithms](references/patterns-and-algorithms.md) | Problem → pattern / algorithm → source |
| LLM-backed systems (RAG, agents, evals) | [ai-engineering](references/ai-engineering.md) | AI-native patterns |

Load only the reference and section the question needs, after the mandatory Decision Protocol below. A fragment identifies a heading, not an automatic partial read. Locate it and read through the next heading of the same or higher level with bounded reads. If section reads are unavailable, load only the relevant reference file. Follow prerequisite links; do not load every reference by default.

When no decision method fits, read [Foundational Principles](references/foundational-principles.md#foundational-principles). For a concrete layering illustration, read [System vs Applied Code](references/system-applied-example.md#worked-example--system-vs-applied-code). Consult [The Canon](references/canon.md#the-canon) only when choosing a source; citation verification still applies.

---

## Decision Protocol

Follow these steps in order. Project evidence establishes the context; references explain the choices. A book citation is not evidence that a particular dependency, boundary, or capability exists in this project.

### 1. Inspect the implementation before choosing a solution

For an existing project, base the recommendation on relevant code, not directory names or the default layout in a reference. When repository access is available, inspect it instead of asking the user for facts you can discover. Use supplied snippets as evidence too, but distinguish them from files you inspected yourself.

Keep the investigation proportional: trace the affected code path and one useful precedent, not the whole repository. For a small placement question, the relevant caller, implementation, and nearby test may be enough. Inspect wider boundaries only when the decision affects them.

1. **Find the intended architecture.** Check project instructions and relevant ADRs or architecture docs, such as `docs/adr/`, `docs/architecture/`, `ARCHITECTURE.md`, `DESIGN.md`, or the README. Discover the actual paths; do not assume these examples exist.
2. **Trace the actual behavior.** Read the relevant symbols and their callers. For a backend feature, follow the entry point → service/use case → domain operation → persistence or external call, as applicable. For a data-structure question, inspect its readers, writers, and ownership. Establish responsibilities from the implementation, not from names like `services/` or `domain/`.
3. **Check the boundaries that decide this question.** Inspect applicable imports, interfaces, dependency wiring, transaction boundaries, tenant/auth context, error translation, and concurrency behavior. Check declared dependencies or configuration before assuming a library or infrastructure capability is available. Do not infer production topology, traffic, or consistency guarantees from a package name.
4. **Read a precedent and its tests.** Find an analogous implemented feature and the relevant tests. Identify reusable policies, repositories, transaction helpers, event publishers, or other mechanisms. Tests show intended behavior and existing test seams; they do not prove every invariant or production guarantee.
5. **Compare the requirement with what exists.** Start with retaining or extending the existing mechanism. Recommend the smallest sufficient change that meets the requirement and preserves useful boundaries. Add an abstraction, dependency, storage system, or service only when a concrete gap justifies it. Smallest sufficient means correct, not simply the fewest changed lines; do not copy an unsafe precedent to preserve consistency.

**Handle disagreements explicitly.** Docs describe intended architecture; code shows current behavior. If they differ, name both and explain whether the recommendation restores the intended boundary or follows the current implementation. Honor explicit project constraints; if they cannot meet the requirement safely, explain the conflict and propose a deliberate change rather than silently bypassing them. Reference defaults are candidates, not instructions to replace a working project convention.

**Keep evidence and uncertainty separate.** Record the key evidence as `path` + `symbol` + observed behavior, or identify the supplied excerpt. Distinguish observed facts, inferences, and unknowns. Cite only paths and symbols actually found; label suggested new paths as proposed. "No relevant precedent found in the inspected area", "code unavailable", and "greenfield" are different situations. If access or context is incomplete, say what is missing and offer only a conditional recommendation when that gap could change the choice. Use defaults for greenfield or after a scoped search finds no suitable precedent, and explain why they fit the known constraints.

### 2. Clarify if the question is ambiguous

When the semantics of the problem are unclear (what "recent" means, whether history must be replayable, how many teams ship this), ask one or two targeted questions before going further. Guessing at requirements is the most expensive failure mode in design work; a wrong architecture costs far more than a short clarification. If the answer would not change the recommendation, state the assumption and continue.

If the gap concerns product needs, behavioral rules or delivery scope, use [requirements-definition](../requirements-definition/SKILL.md) for that preparation rather than making a technical choice settle an unknown product decision. Use [task-decomposition](../task-decomposition/SKILL.md) when the requested result needs an execution breakdown. A focused design question with a known contract stays in this skill; design-only work does not authorize implementation.

### 3. Route

Pick the reference from the table above. If the question spans two classes (e.g., "should this cache be in-memory or Redis, and which layer owns it?"), route each part separately. If nothing fits, say explicitly that no decision tree applies and reason from the [foundational principles](references/foundational-principles.md#foundational-principles).

### 4. Apply the reference's decision method

Follow the method the reference actually defines. For an ordered early-exit tree such as code placement, evaluate questions in order and stop at the first matching branch; split mixed-responsibility code into separate decisions. For the architecture selector, consider all four questions: deployment topology, internal boundaries, read/write models, history, and workflow coordination are compatible dimensions, not mutually exclusive branches. For a pattern catalog, match the observed symptom and check its "when not to" conditions; do not invent a decision tree where none exists.

Keep this reasoning concise in the answer. When a proposed choice depends on a consistency, atomicity, eviction, or safety guarantee, state its assumptions and failure modes. A familiar pattern name is not proof of the guarantee.

### 5. Name the choice

State the layer, the style, or the structure in one line using the project's terminology. Retaining an existing mechanism is a valid recommendation; do not introduce a named pattern just to make the answer look architectural. Put the runner-up in the trade-off. If a missing requirement could change the choice, ask the targeted question or label the recommendation as conditional instead of pretending certainty.

### 6. Cite

Give one short source citation for the deciding principle. Prefer the canon for general design principles, and official specifications or version-appropriate documentation for concrete protocol and library behavior. Reference-file citations are starting points, not proof that a chapter supports every detail of a recommendation.

Cite a chapter only when you can substantiate its attribution. If the exact chapter is unverified, omit its number and say so; cite the book and topic instead. Never invent a chapter, quotation, URL, or claim to have verified a source. Name an outside source and why it is the better fit, such as RFC 9562 for UUIDv7. Distinguish literature-backed guidance from practice-based advice when no source has been verified.

### 7. Justify with the deciding factor

One sentence on *why this beats the runner-up*, tied to an observed project constraint or an explicit assumption. Not a list of pros — the single consideration that tipped it. When proposing a new mechanism, identify the concrete gap in the existing one that justifies the added complexity.

### 8. Self-check before answering

- Did I read the relevant implementation and tests, not just list directories or read an ADR? If unavailable, did I state that limitation?
- Does the recommendation cite concrete observed files/symbols or clearly identify the supplied snippets?
- Did I separate observed behavior, intended architecture, assumptions, and unknowns, and address any docs/code mismatch?
- Did I consider retaining or extending an existing mechanism, and explain the gap if I recommend a new one?
- Am I following a safe project precedent rather than copying a bug or importing a reference's default layout?
- Did I route through a reference, or explicitly say none applied?
- Is the choice clear and expressed in the project's terminology, or explicitly conditional on a missing deciding fact?
- Are major claims supported by project evidence or an appropriate source, with uncertain attributions labeled rather than invented?
- Is System vs Applied separation explicit — which layer, and why?
- Is the testing strategy named (unit / integration / e2e) and tied to the layers?
- Are concurrency, idempotency, and error translation across boundaries addressed where relevant?
- Have I stated the actual guarantee and failure assumptions, without confusing async invalidation with strict consistency, atomic intent recording with broker delivery, or bounded cache eviction with a business invariant?
- If the data is scoped to a tenant, user, or account: is the scoping enforced at the adapter, not trusted from the caller?
- For a larger call: is there a filled trade-off table and an ADR offered?

---

## Response Format

Keep the answer decision-shaped. A reader should be able to act on the first three lines.

Include the recommendation, single deciding factor against the runner-up, substantiated source, observed project evidence or limitations, project fit and smallest sufficient change, and layer-specific testing. Include a trade-off table for architecture-level decisions and relevant concurrency, idempotency, error translation, or tenant-scoping risks. An [optional response layout](references/response-example.md) illustrates these fields; the fields are required where applicable regardless of whether the example is read.

For small placement or data-structure questions, keep the recommendation, deciding factor, source, one project-evidence sentence, and one testing sentence; merge sections if useful. Do not omit the project evidence just to keep the answer short. For greenfield or unavailable code, replace it with the known constraints and explicit limitations, not invented paths. Expand only when the decision is expensive to reverse.

---

## Recording the Decision

For any choice that is hard to reverse — a new service boundary, a new persistence model, a new cross-service workflow — offer to write an Architecture Decision Record. Use [assets/adr-template.md](assets/adr-template.md). Put it where the project already keeps ADRs; if there is no such place, propose `docs/adr/NNNN-<slug>.md` and say so.

An ADR is worth writing when the next engineer who touches the area will ask "why is it done this way?" It is not worth writing for a placement decision that follows an existing precedent. — *Fundamentals of Software Architecture* ch.19 (Architecture Decisions).

---

## Tone

- **Mentoring, not decreeing.** Show the principle in action; explain why the runner-up loses.
- **Direct.** Trade-off tables are brief and tabular. Pick one option and say why.
- **Honest about fit.** If the project's convention contradicts the "textbook" answer, say which you are recommending and why, with evidence from the relevant implementation. Consistency is a real architectural characteristic, not a compromise, but does not justify repeating an unsafe pattern.
- **Show, don't lecture.** A ten-line code sample beats three paragraphs.
