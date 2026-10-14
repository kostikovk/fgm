---
name: task-decomposition
description: Turn defined product requirements into completed behavioral slices and traceable technical tasks, enablers, or research spikes. Use when a product request spans multiple changes, needs an executable backlog for AI-assisted work, or an existing task set has coverage or dependency gaps. Route unresolved product behavior to requirements-definition and open technical choices to design-decision; this skill prepares work rather than implementing it.
---

# Task Decomposition

Produce a task set whose assembled result satisfies the selected product requirements. Keep product outcomes, behavioral slices, technical work, and their evidence distinct. A small coherent request can remain one task; do not create a backlog merely to apply this skill.

## Essential workflow — always apply

1. Locate the product request and canonical requirements. Inspect relevant implementation, contracts, tests, and project constraints before naming technical work. Use [requirements-definition](../requirements-definition/SKILL.md) when behavior or scope is unresolved; preserve its requirement IDs and statuses. Do not invent product decisions, repository paths, a stack, or implementation capabilities.
2. Select the delivery scope and identify completed behavioral slices before splitting their technical work. Preserve necessary access, security, and data guarantees in the first usable slice. Use the slicing reference when the request needs multiple slices or the proposed split is primarily by technical layer.
3. Give each execution-ready task a bounded result, requirement links, context, scope, dependencies, acceptance checks, evidence expectations, and stop conditions. Read the task-contract reference before preparing work for execution or reviewing an existing task set. Research spikes and technical enablers need appropriate evidence, not fabricated individual user value.
4. Check requirement coverage, justified task scope, ownership, dependency order and contracts, and acceptance of the assembled scenario. Green isolated components or mocked dependencies do not establish assembled behavior. Identify who or which task performs that acceptance.
5. Distinguish ready work from provisional planning. A product or technical decision blocks only the work that depends on it; independent ready branches may continue. Keep later scope coarse and explicitly deferred. Route open technical choices to [design-decision](../design-decision/SKILL.md); pass ready implementation work to [code-change](../code-change/SKILL.md), retaining its TDD/AAA gates.
6. For changed requirements, invoke the change procedure owned by [requirements-definition](../requirements-definition/SKILL.md), then revise affected slices, tasks, dependencies, and acceptance evidence. Mark affected task readiness and prior evidence for reassessment; preserve unaffected ready work.

## Compose with planning and review

Use [plan](../plan/SKILL.md) when the task set needs an overall delivery route, decision checkpoints, risks or a verification sequence. A planning outline can precede decomposition; return the resulting task contracts, dependency mapping and assembled acceptance responsibility to the completed plan. Planning does not replace this skill's task contract.

When decomposition is part of a plan, include its tasks in the mandatory [plan-review](../plan-review/SKILL.md) before that plan is reported ready. Review findings about task coverage, coherence or contracts return here; corrected tasks trigger re-review of affected plan content. Standalone decomposition does not require creating a plan merely to finish a task set.

## Navigation — read for the current need

| Need | Read | Prerequisite |
|---|---|---|
| Large scope, layer-only tickets, or uncertainty about where to split | [Slicing](references/slicing.md) | Identify the selected requirements and outstanding decisions |
| Execution-ready tasks, enablers/spikes, or coverage/dependency review | [Task contract and coherence](references/task-contract.md) | Read canonical requirement semantics through requirements-definition; select the delivery scope |
| A concrete example of the product-to-task mapping | [Worked example](references/worked-example.md) | Use the contract reference for real task preparation; example choices are illustrative |

## Deliver and stop

Return the selected outcome and scope, slice/task map with requirement references, ready/provisional/blocked/deferred status, dependency order and contract gaps, coverage gaps, and the assembled-scenario acceptance responsibility. Include decision owners or missing inputs where known, without pretending unknown authority is established.

Stop decomposition when near-term work has coherent, verifiable contracts and the assembled result has an acceptance path. Further splitting should resolve an independent outcome, knowledge gap, or coordination boundary. Do not use fixed hours, files, tokens, model sessions, or mandatory agent topology as readiness rules. Preparing tasks does not authorize implementation, publishing tickets, or external actions.
