---
name: plan
description: Prepare an evidence-backed work plan before execution or revise an existing plan when scope, constraints or dependencies change. Investigate before asking questions, close material planning decisions, compose with task-decomposition when needed, and automatically finish through plan-review. Planning alone does not authorize implementation.
---

# Plan

Turn a selected outcome into an actionable route through preparation, delivery and verification. Match detail to the work: a small coherent task can use a short response; a broad request may need phases and linked execution tasks.

## Essential workflow — always apply

1. Establish the requested outcome, scope, constraints, sources and authority from the request and project context. Inspect relevant implementation, callers, tests, contracts and established decisions before proposing technical work. Distinguish current behavior from desired behavior; do not invent paths, capabilities or product choices.
2. Identify deciding questions and try to answer them independently from relevant repository evidence before asking the user. Reuse authoritative answers already supplied. Group remaining questions by category or shared decision, explain their effect on the plan, and recommend a reasoned option when evidence cannot decide. Recommendations remain proposals until a required decision is made. Read the clarification procedure when questions remain.
3. Close every material question for the scope being finalized. Record evidence-backed answers and authoritative decisions; explicitly agreed exclusions/deferment may narrow that scope. Unresolved questions keep dependent work provisional/blocked, while unaffected branches can be prepared. Label an incomplete plan partial rather than presenting assumptions as settled.
4. Choose phases, prerequisites, observable results, verification responsibilities and stop/replan conditions from verified context. Use [requirements-definition](../requirements-definition/SKILL.md) for unresolved product behavior and [design-decision](../design-decision/SKILL.md) for open technical choices. Use [task-decomposition](../task-decomposition/SKILL.md) when the scope needs slices or bounded tasks; link its contracts instead of copying its schema or coherence rules.
5. Assemble the candidate plan with its source/baseline, task links where needed, risks and acceptance path. Keep contract readiness, satisfied start prerequisites and authorization distinct. Use the project's artifact conventions; return the plan in the response when persistence is not requested. Repository writes follow [git-delivery](../git-delivery/SKILL.md).
6. **Automatically apply [plan-review](../plan-review/SKILL.md) before reporting the selected plan scope ready.** Review the completed candidate together with any decomposed tasks, including small plans. Use an independent reviewer when available; otherwise perform and disclose a separate self-review pass. This is a workflow handoff, not a promise of installed hooks or a required runtime.
7. Correct supported planning defects within the authorized planning scope, route requirement/task/design gaps to their owners, and re-review changed content and dependent assumptions. Do not silently settle product decisions to obtain approval. Finish with the reviewed baseline, verdict and remaining execution prerequisites. If review cannot complete or a material question remains, report provisional/blocked scope; do not claim readiness or start execution.

## Navigation — choose the current need

| Need | Read | Prerequisite |
|---|---|---|
| Unanswered questions, conflicting evidence or recommendations | [Clarification](references/plan-contract.md#clarify-before-finalizing) | Identify selected scope and available sources |
| Substantial plan, handoff or readiness check | [Plan record](references/plan-contract.md#plan-record) | Apply the essential gates; reuse project format |
| Broad scope, existing backlog, small task or changed inputs | [Composition](references/composition.md) | Identify canonical requirements/decisions and existing tasks |

Read only the selected route and its prerequisites. Detail in references does not make the essential gates optional.

## Deliver and stop

Return the selected outcome, ordered work, verification path and reviewed verdict, with evidence-backed context, settled decisions, any provisional/blocked/deferred scope and unsatisfied execution prerequisites. Include review independence and limits. A review-ready plan is preparation, not implementation evidence or additional authority. Stop when the selected scope has passed plan-review or a concrete unresolved decision/evidence gap prevents readiness; continue only unaffected authorized work.
