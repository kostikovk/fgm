# Reviewing Plan Content

Use the relevant section after establishing the original request, selected scope and reviewed version. The entrypoint owns verdicts and read-only boundaries.

## Contents

1. [Grounding and unresolved decisions](#grounding-and-unresolved-decisions)
2. [Feasibility and coverage](#feasibility-and-coverage)
3. [Findings and re-review](#findings-and-re-review)

## Grounding and unresolved decisions

Compare the plan against the user's actual request, later corrections and authoritative requirements or decisions. Identify which source controls a disputed claim; an illustrative example, stale draft or AI proposal is not a confirmed requirement. Use [requirements-definition](../../requirements-definition/SKILL.md) for requirement semantics, uncertainty and change handling.

Verify material repository facts with the affected code, callers, tests, contracts and configuration. Trace enough of the actual path to determine whether the proposed change and check are plausible. A filename, ticket title or assertion that a utility exists does not prove its semantics. When a fact cannot be checked, identify the evidence gap and affected conclusion. Do not infer a defect from absence of access.

Review the planner's question handling:

- Did it inspect the relevant available sources before requesting answers?
- Which answers were found, which user choices were already confirmed, and which material decisions remain open?
- Were remaining questions grouped by category or shared decision, with their effect on the plan explained?
- Were recommendations supported by repository evidence and explicit tradeoffs, and kept distinct from the user's confirmed choices?

An unanswered question blocks only dependent work. A stated recommendation can support a provisional option; it cannot silently become a user decision. Existing user authority and clear reversible implementation choices need no invented approval gate. Route materially open product or technical decisions to their owner; preserve established scope.

Assess the plan artifact separately from its preparation history. Supplied read logs are supplied evidence; fresh inspection verifies current facts. If observable actions are missing, do not claim discovery happened before questions. Report that execution check as unverified. A missing historical record does not itself establish a defective plan, but missing material facts or decisions can prevent content readiness.

## Feasibility and coverage

For a small coherent plan, check that the result, order, dependencies and completion evidence are sufficient without requiring a backlog. For a task set, apply [Task contracts and coherence](../../task-decomposition/references/task-contract.md), including its coverage mapping, dependency review and distinction between contract readiness and execution readiness. Do not copy or replace its schema.

Trace each selected outcome to planned work and an appropriate acceptance path. Check producer/consumer responsibilities and ordering against actual interfaces. Parallel execution needs compatible contracts; a proposed agent arrangement is not proof of independence. Flag cycles, absent integration responsibility or incompatible interfaces with a concrete consequence.

Check whether the proposed validation could distinguish success from failure at the relevant boundary. Unit checks, mocks, typechecks and builds each prove bounded claims; they do not automatically establish assembled behavior or product outcomes. Reuse [decomposition's evidence boundaries](../../task-decomposition/references/task-contract.md#assign-acceptance-and-preserve-evidence-boundaries) and the technical criteria of affected skills. Review executable work against [code-change](../../code-change/SKILL.md) and its TDD/AAA composition; this read-only review does not execute the planned implementation.

Separate an adequately defined prerequisite from a content gap. For example, a task scheduled after a defined migration may be content-ready while waiting for the migration. An unverified migration capability or unspecified consumer contract may leave the dependent plan blocked or require revision. Identify actual environment, access or authority requirements; plan approval grants no new execution permissions.

Avoid extending scope with speculative hardening or architecture. Recommend an addition only when a selected requirement, required guarantee or demonstrated prerequisite justifies it. If an alternative is merely preferable, treat it as advisory and explain the tradeoff.

## Findings and re-review

Use the project's report format or concise prose. Every finding must convey:

| Element | Meaning |
|---|---|
| Location | The plan step, branch, task, decision or acceptance criterion affected |
| Trigger | The supported condition under which the issue matters |
| Consequence | How it prevents the selected outcome, safe ordering or valid acceptance |
| Correction | The necessary change or decision, without implementing it |
| Owner | The existing responsibility that must resolve it; disclose unknown authority |

A blocking finding requires supported material impact. An advisory offers a useful improvement without preventing readiness. Consolidate findings caused by the same mechanism; do not multiply symptoms or repeat corrected issues. Formatting preferences, unsupported hypotheticals and reviewer taste do not justify `REVISE`.

Apply the entrypoint's verdicts to selected scope and affected branches. If supported defects and missing evidence coexist, preserve both categories; the overall selected scope is not ready, even if independent branches are. Report prerequisites separately from verdicts. Do not claim an independent reviewer or check runner when only self-review or supplied records exist.

Return correction requests rather than a replacement plan. The plan owner revises composition, while requirements, decomposition and design owners resolve their corresponding gaps. Re-review corrected locations, their affected dependencies and any changed requirement baseline. Reopen unaffected content only when new evidence changes its assumptions.

End with `READY` when material selected-scope content is coherent and sufficiently grounded. If evidence or a decision remains unavailable, report the bounded gap. Do not automatically call the planner from the reviewer, require another review of the review, or invent new objections merely to sustain an iteration.
