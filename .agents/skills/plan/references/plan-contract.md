# Plan contract and clarification

Use the project's plan format. This reference owns planning-specific records; product semantics stay with [requirements-definition](../../requirements-definition/SKILL.md), execution task contracts with [task-decomposition](../../task-decomposition/SKILL.md).

## Contents

1. [Clarify before finalizing](#clarify-before-finalizing)
2. [Plan record](#plan-record)
3. [Completion and evidence](#completion-and-evidence)

## Clarify before finalizing

Identify questions that affect outcome, scope, acceptance, dependencies, compatibility or the proposed implementation route. Investigate the relevant code and its callers, tests, schemas, configuration, documentation and prior decisions before escalating. Follow the actual boundary far enough to answer the question; reading an unrelated file does not establish discovery. Record unavailable sources or conflicting evidence rather than inventing a result.

Separate repository facts from desired behavior. A caller and test may establish the current default; they do not establish whether a requested new feature should preserve it. Reuse authoritative user answers without seeking repeat confirmation.

For unanswered questions:

- Group by a shared category or decision, such as behavior, data contracts or compatibility. Keep unrelated choices separate; avoid a large interview that delays independent settled work.
- Give each group the relevant evidence, remaining uncertainty and effect on the plan. Ask a concrete question with enough context for the user to decide.
- Offer a recommended option when useful, with its rationale, relevant alternatives/trade-offs and effect on scope or acceptance. Label it a proposal; do not turn absence of an answer into agreement.
- Route product decisions through [product elicitation](../../requirements-definition/references/elicitation.md) and open technical choices through [design-decision](../../design-decision/SKILL.md). Planning coordinates the questions; it does not become a second requirements or design owner.
- Routine authorized choices can be made from evidence without a permission ritual. A material choice whose missing answer changes requested behavior, scope or a deciding contract remains open until the appropriate authoritative decision is available.

Track each material question's disposition: source-backed answer, authoritative decision, or explicit agreed exclusion/deferment from the selected scope. Include the affected work and decision owner when known. Revisit dependencies and acceptance after answers; newly revealed deciding questions need the same discovery/clarification loop.

Close all material questions for the selected scope before finalizing it. Later scope may remain coarse and deferred, but silently dropping a requirement is not closure. If an answer is unavailable, show a provisional plan and specific blocker; independent ready branches can still be prepared. Do not require every possible future question to be resolved.

Example of a grouped question: “The existing importer rejects duplicate external IDs, and the downstream writer depends on that uniqueness. For the new bulk flow, should duplicates reject the batch or be reported per row? I recommend per-row reporting if partial success is intended; otherwise preserve batch rejection.” The current contract is evidence; the recommendation does not establish the new product decision.

## Plan record

For a small task these meanings fit in concise prose. For substantial work use identifiable phases and source links:

| Field | Meaning |
|---|---|
| Outcome and scope | Selected result, relevant exclusions and authoritative acceptance sources |
| Baseline and context | Identified plan version/content, requirement/decision/task revisions where applicable, verified relevant repository locations |
| Decisions and uncertainty | Evidence-backed answers, confirmed choices, proposed assumptions, unanswered/deferred questions and affected branches |
| Ordered work | Phases with observable results, prerequisites, decision checkpoints and bounded task links when decomposition is needed |
| Verification path | Expected task checks, assembled acceptance and responsibility, with environment/prerequisite gaps |
| Risks | Concrete failure conditions, impact, mitigation or bounded investigation |
| Readiness and authority | Content/contract readiness, unsatisfied start prerequisites, planning versus execution authorization |
| Stop/replan conditions | Completion evidence and changes/failed assumptions that invalidate affected work |
| Review | Reviewed scope/baseline, verdict, blocking/advisory findings, independent or disclosed self-review, evidence limits |

Use [task contract and coherence](../../task-decomposition/references/task-contract.md) for execution task fields, dependency coherence, coverage and assembled acceptance. Do not create a competing task schema. Readiness does not grant permission; a reviewed plan may still wait on a predecessor or external access.

Use existing project artifact conventions. Response-only planning does not require a file, branch, ticket or hidden runtime state. If an authorized repository artifact is appropriate, follow the workspace policy before writing it. Refer to verified paths, not invented future implementation locations. Proposed new paths must be labeled proposals.

## Completion and evidence

Before returning the selected scope as ready, apply [plan-review](../../plan-review/SKILL.md) to the candidate and linked tasks. A small plan gets a short review rather than an exemption. Supply the request, sources, baseline and available discovery/clarification records to the reviewer. Preserve its independent findings; do not give it an expected verdict.

Keep planned checks separate from checks actually executed. Planning need not run implementation tests to describe a feasible verification path. Missing runner/environment proof remains explicit; a reviewer agreeing with the plan does not establish implemented behavior.

Revise affected content, then review the new baseline and dependent assumptions. If requirements change, use their canonical change procedure first. Do not chase advisory wording indefinitely. Report READY only for the reviewed scope with no supported content blockers; report execution prerequisites separately. If material uncertainty or review evidence remains unavailable, retain provisional/blocked status and its recovery path.
