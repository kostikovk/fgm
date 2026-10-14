# Requirement contract and lifecycle

This is the shared owner of requirement semantics used by requirements-definition and task-decomposition. Keep one authoritative contract or identify the existing authoritative records; downstream tasks reference it instead of copying full requirements. The tables below describe information to preserve, not a mandatory document template.

## Contents

- [Requirement record](#requirement-record)
- [Scope readiness](#scope-readiness)
- [Requirement changes](#requirement-changes)
- [Evidence and product outcomes](#evidence-and-product-outcomes)
- [Sources and adaptation](#sources-and-adaptation)

## Requirement record

For substantial or evolving work, record:

| Information | Purpose |
|---|---|
| Stable ID and parent outcome | Trace behavior to its product need without relying on ticket wording |
| Behavior and scope | Identify actor, triggering circumstances, expected result and relevant boundaries |
| Source and rationale | Show where the requirement came from and why it exists |
| Status | Distinguish confirmed behavior, proposed assumptions, unresolved decisions, deferred or superseded scope |
| Acceptance examples | Show representative success, boundary and failure outcomes relevant to the rule |
| Evidence plan | Identify how acceptance can be checked and any prerequisites or evidence gaps |
| Relationships | Identify dependencies, conflicts and affected scenarios or constraints |

Observed facts/current behavior are supporting evidence, not automatically intended requirements. Keep assumptions and open questions separately identifiable even when stored beside requirements. A source may be an explicit request, agreed product decision or existing authoritative contract. When it is absent, say so; AI-generated proposals are not confirmation. Do not imply stakeholder sign-off that did not occur.

Write each requirement so its behavior can be checked without choosing internal classes, files or algorithms. Implementation details belong here only when they are actual constraints, with their rationale recorded. Cross-cutting constraints may apply to multiple scenarios; reference their IDs rather than duplicating them.

A compact clear task can state the parent need, behavior, applicable constraints and acceptance directly. IDs and a separate document are useful when there are multiple linked requirements, parallel work or ongoing change tracking, not a prerequisite for every edit.

## Scope readiness

Readiness applies to a specified scope and contract version, not the entire product request by default. State:

- **Ready:** intended behavior and boundaries are settled, necessary constraints are known, acceptance is feasible, and dependencies can be satisfied without guessing product decisions.
- **Blocked:** a named unknown, conflict or missing prerequisite changes the behavior or acceptance of this scope. Identify the affected requirements and the decision or evidence needed.
- **Deferred:** the scenario is outside the selected delivery scope, with its disposition visible. Deferred work is not delivered work.

Optional labels may follow the project's tracker. Explain the decision rather than relying on a status word. An assumption may support exploration or a clearly authorized provisional experiment; it does not make outcome-changing product behavior execution-ready. If an independent subset is ready, describe its boundary and why unresolved scope cannot invalidate it. Pass technical uncertainty to design-decision; pass unknown product behavior back to elicitation.

## Requirement changes

When a source, behavior, constraint or acceptance rule changes:

1. Update the authoritative record, preserving the stable ID for the same evolving requirement and identifying replacements when meaning changes substantially. Use the project's version/history mechanism to show what changed and why.
2. Trace affected scenarios, tasks, interfaces, dependencies, tests and existing evidence. Include consumers of shared constraints; do not stop at the directly edited ticket.
3. Mark affected readiness, task plans and evidence stale until reassessed. Retain historical evidence as evidence of the earlier contract, not proof of the new behavior. Reuse unaffected evidence only with an explicit reason it still applies.
4. Revise the affected scope and acceptance/evidence plan; route technical implications to design-decision or implementation as appropriate. Keep independent ready work progressing.

If traceability is incomplete, report that impact uncertainty and inspect the likely consumers before claiming readiness. Avoid replacing the entire backlog merely because one requirement changed.

## Evidence and product outcomes

Define the expected evidence alongside behavior. Automated tests, scenario walkthroughs, demonstrations, accessibility checks or measurements may be appropriate; choose according to the claim and environment. Writing an acceptance criterion or test plan is not execution evidence.

Keep these claims distinct:

| Claim | Evidence needed |
|---|---|
| Technical task meets its contract | Checks of the task's behavior, constraints and relevant regressions |
| Delivered scenario works as assembled | End-to-end or integration acceptance spanning the participating components and dependencies |
| Product need is satisfied | Validation with intended use and affected people, or relevant outcome observations |

Assign or identify how assembled acceptance and product validation will happen. A business metric may require observation after release; record the pending measurement and its source rather than claiming the outcome from passing unit tests. Missing production access need not prevent requirements work, but it limits what can be verified now.

## Sources and adaptation

[NASA — How to Write a Good Requirement](https://www.nasa.gov/reference/appendix-c-how-to-write-a-good-requirement/) informs verifiability, rationale, avoiding unsupported implementation constraints and traceability. [SEBoK — System Requirements Definition](https://sebokwiki.org/wiki/System_Requirements_Definition) informs requirement attributes and early verification criteria; [Requirements Management](https://sebokwiki.org/wiki/Requirements_Management) informs change and impact tracking. [NASA — Product Realization](https://www.nasa.gov/reference/5-0-product-realization/) informs verification versus validation.

Stable IDs, scope-specific readiness, context references and stale-evidence handling are this repository's practical adaptation for AI-driven work, not a prescribed schema from those sources.
