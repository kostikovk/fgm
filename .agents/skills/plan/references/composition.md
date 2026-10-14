# Compose planning with existing owners

## Small coherent request

Use a compact outcome, verified context, ordered action, checks and stop condition. Do not manufacture tickets, a PRD, an ADR or time estimates merely to apply planning. A completed candidate still passes through [plan-review](../../plan-review/SKILL.md). Planning-only authorization remains planning-only after READY.

## Broad work

Start with an evidence-backed outline. Use [requirements-definition](../../requirements-definition/SKILL.md) when behavior or scope needs a decision, and [design-decision](../../design-decision/SKILL.md) before dependent work assumes an unsettled technical choice.

Use [task-decomposition](../../task-decomposition/SKILL.md) to produce behavioral slices and technical task contracts. Incorporate its requirement/task mapping, readiness, dependencies and assembled acceptance responsibility into the completed plan by reference. Review the combined plan and tasks; a reviewed outline does not approve details added later.

The route is outline → decomposition when needed → completed candidate → plan-review → material corrections and re-review. This is not a second task hierarchy or a mandatory decomposition step.

## Existing backlog or supplied plan

Reuse stable task IDs, requirement links and existing contracts. Planning adds the overall sequence, decision points, risks and verification path; return missing task contracts or coherence defects to decomposition.

A request to review an existing plan can use plan-review directly, without first rewriting it. A backlog can be planned without repeating already settled requirements. Inspect the baseline and supplied records before assuming they are current.

## Changed inputs and review findings

For changed requirements use [requirement changes](../../requirements-definition/references/requirement-contract.md#requirement-changes), then update affected phases, tasks and acceptance. Preserve unaffected scope; identify evidence invalidated by the change.

Route a review finding to the narrowest owner: plan ordering/context to plan, product meaning to requirements-definition, technical choices to design-decision, task coverage/contracts to task-decomposition. Return corrected content to review; the review skill does not recursively activate plan or silently rewrite the artifact.

Execution begins only for ready, authorized work through [code-change](../../code-change/SKILL.md), retaining its TDD/AAA gates. An unavailable deciding answer or required review evidence is a blocker, not a reason to implement the recommendation.
