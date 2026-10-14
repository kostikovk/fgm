---
name: plan-review
description: Critically assess a proposed work plan against the original request, authoritative requirements, repository facts, dependencies, and verification paths. Use after a finalized plan or directly to review an existing plan; return evidence-backed readiness and corrections without implementing or rewriting the work.
---

# Plan Review

Determine whether the selected plan content can guide the requested work. Review the plan as a proposal, not as proof that its assumptions, checks, or user decisions are established.

## Essential workflow — always apply

1. Establish the original request, selected scope, plan version and authoritative requirement/decision baseline. Inspect relevant source, callers, contracts, tests and configuration to verify material claims. If required evidence is unavailable, state the gap rather than inventing a finding or readiness.
2. Use an independent reviewer when available. If independent review is unavailable, perform a separate review pass after drafting and disclose self-review and its limits. Independence is useful scrutiny, not evidence that planned work has executed.
3. Check request fidelity, realistic execution order, prerequisite contracts, verification and completion conditions. When tasks or slices are involved, read the [task contract and coherence](../task-decomposition/references/task-contract.md); use that owner for coverage, dependency and assembled-acceptance criteria rather than inventing another task schema.
4. Check that the planner investigated relevant code and authoritative sources before asking the user, grouped remaining questions by shared decision, and distinguished reasoned recommendations from confirmed choices. Material questions affecting selected scope must be resolved or explicitly block the dependent branch; independent ready branches remain reviewable.
5. Report only supported defects and useful advisories. Each finding identifies its location, triggering condition, consequence, required correction and responsible owner. Distinguish blocking defects from optional improvements; a clean plan receives no fabricated findings.
6. Return `READY`, `REVISE`, or `BLOCKED` using the verdict rules below. After corrections, re-review affected content and changed dependent assumptions. Stop when material defects and evidence gaps are resolved; do not restart planning automatically or prolong review with cosmetic objections.

The companion [plan](../plan/SKILL.md) must invoke this review after every finalized plan, including compact plans. This skill can also be called directly. Mandatory review is a workflow instruction, not an installed automation or runtime hook.

Review is read-only: do not repair or rewrite the plan, implement tasks, publish findings, or create external artifacts without a separate authorized request. Corrections return to their existing owner: requirement meaning to [requirements-definition](../requirements-definition/SKILL.md), task coherence to [task-decomposition](../task-decomposition/SKILL.md), open technical choices to [design-decision](../design-decision/SKILL.md), and plan composition to `plan`.

## Verdict and delivery — always apply

- **READY:** selected content has no supported blocking defect or required evidence gap. This states content readiness; pending predecessors, credentials, environments and execution authority remain separate start prerequisites.
- **REVISE:** evidence establishes a material defect in the selected content. State what must change and why.
- **BLOCKED:** required evidence or a deciding input is unavailable, preventing a readiness judgment. State the missing input and affected branch.

Give an overall verdict for selected scope and branch-level distinctions when needed. A known defect can coexist with missing evidence: report both, and do not imply that fixing the defect resolves the evidence gap. Preserve unaffected ready branches.

Return the reviewed scope/baseline, reviewer independence, verdict with reasons, supported blocking findings, useful advisories, execution prerequisites and evidence limits. Missing action records leave discovery order or check execution unverified; a good final artifact alone cannot establish them. A small clean plan can receive a short review.

## Navigation — read for the current need

| Need | Read | Prerequisite |
|---|---|---|
| Ambiguity, repository assumptions, questions or recommendation history | [Grounding and unresolved decisions](references/review-procedure.md#grounding-and-unresolved-decisions) | Establish the request and reviewed baseline |
| Multi-stage plan, backlog, producer/consumer boundaries or acceptance gaps | [Feasibility and coverage](references/review-procedure.md#feasibility-and-coverage) | Read the decomposition task contract when tasks are present |
| Supported findings, mixed branch verdicts or a corrected plan | [Findings and re-review](references/review-procedure.md#findings-and-re-review) | Keep reviewed and corrected versions distinguishable |
