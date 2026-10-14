---
name: requirements-definition
description: Turn a product request into scoped, verifiable requirements when intended behavior, acceptance, or product decisions need clarification before decomposition or implementation.
---

# Requirements Definition

Produce a usable behavioral contract that preserves the product need. Separate what people need from the proposed implementation and from what the system currently does. A clear small request can use a short task description; create a larger requirements artifact only when scope, ambiguity, or change tracking warrants it.

## Required boundaries

- Establish the affected people, problem, intended outcome, scope and relevant constraints from available evidence. Inspect existing behavior and contracts when they affect the request; code describes current behavior, not automatically the desired product decision. A sufficiently clear supplied request can support compact requirements preparation without code access; record any implementation-context gaps at handoff.
- Keep confirmed requirements, observed facts, proposed assumptions, unresolved decisions and deferred scope distinct. Do not promote an AI suggestion to a confirmed requirement. Cite its source or record that confirmation is missing. Explicit authoritative session instructions already establish their stated requirements; do not ask for the same approval again.
- Define observable behavior with acceptance examples and a feasible evidence plan. Include applicable error cases and necessary guarantees, rather than deferring them silently to implementation.
- Identify unresolved decisions that change the outcome, permitted behavior or acceptance. Keep affected scope blocked; independent scope can proceed when its boundaries and dependencies are settled. Do not require a product-owner approval ritual for already clear, authorized work.
- Before handoff, state the ready scope, blocking questions, deferred scope and checks that will prove it. For substantial work, use stable requirement IDs and the [requirement contract](references/requirement-contract.md).
- When requirements change, perform [impact analysis](references/requirement-contract.md#requirement-changes) before reusing affected plans or evidence. Preserve unaffected ready work.

## Choose the route

| Request or uncertainty | Read / apply | Result |
|---|---|---|
| Chaotic request, competing needs, or unclear rules | [Elicitation](references/elicitation.md) | Grounded outcomes, rules, examples and open decisions |
| Substantial requirements, acceptance or readiness handoff | [Requirement contract](references/requirement-contract.md) | Traceable records and a scope-specific readiness decision |
| Changed requirements or existing acceptance evidence | [Requirement changes](references/requirement-contract.md#requirement-changes) and [evidence](references/requirement-contract.md#evidence-and-product-outcomes) | Affected work identified; stale readiness/evidence marked |
| Need to see intake translated into a usable contract | [Worked example](references/worked-example.md) | An illustrative request, decisions and scoped handoff |

Read the selected reference before performing that branch. Routes add detail; they do not make the required boundaries optional.

## Handoff and completion

Return the scoped contract, its sources and acceptance/evidence plan, plus explicit unknowns and readiness. Use [task-decomposition](../task-decomposition/SKILL.md) when the ready scope needs slices or technical tasks. Use [design-decision](../design-decision/SKILL.md) for open technical choices and [code-change](../code-change/SKILL.md) for implementation; their TDD and AAA responsibilities remain unchanged. Requirements work alone does not authorize implementation or external publication.

Technical verification shows conformity to the contract. Assembled-scenario acceptance shows the pieces work together. Product validation checks whether the delivered behavior satisfies the original need; distinguish these claims and their evidence.
