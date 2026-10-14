# Learning Cycle

## Contents

1. [Extract a supported lesson](#extract-a-supported-lesson)
2. [Reassess a lesson](#reassess-a-lesson)
3. [Route a shared procedure defect](#route-a-shared-procedure-defect)
4. [Evaluate learning](#evaluate-learning)

## Extract a supported lesson

Start from a concrete discrepancy: the requirement, intended invariant, or explicit preference versus the observed action or result. Use relevant task artifacts rather than reconstructing an entire conversation.

1. **Establish the episode.** Identify what was attempted and what exposed the mistake: a regression, a reproducible failure, a review finding, an incorrect assumption corrected by the user, or an abandoned approach with a demonstrated defect. Include unsuccessful attempts when they explain why the final approach works. Record a shared decision as shared experience rather than inferring who is at fault.
2. **Check the explanation.** Separate the observed failure from its suspected cause. A passing retry alone does not explain a flaky test; a tool error alone does not prove a code defect. A user correction establishes their stated requirement or preference, not a technical cause beyond what they said. If diagnosis is still needed, use the active debugging workflow instead of inventing an explanation.
3. **Identify the changed decision.** Ask what a future agent should notice or do before repeating the mistake. Prefer a concrete check, an existing project mechanism, or a corrected assumption. "Be more careful" and a transcript of commands are not actionable lessons. No reusable decision means no memory entry is needed.
4. **Bound the lesson.** State its trigger, relevant project/technology conditions, exceptions, and verification. Distinguish a local convention from a technical invariant. Do not require the exact fix from one incident when another mechanism can preserve the same invariant.
5. **Assess support.** Mark a lesson `verified` only when the stated action is supported within its stated scope: for example, a demonstrated failure and checked fix, an accepted review finding corroborated by the affected contract, or an explicit user correction scoped to that user/project. Mark a useful but unresolved explanation `candidate` and say what would verify it. `Verified` is not a claim of universal validity.
6. **Retain only useful evidence.** Link to a test, affected symbol, diff, issue, or accessible task record; preserve a short factual excerpt if the original is ephemeral. Describe whether verification was executed now or supplied as a prior record. Redact sensitive values and store concise observable facts, not hidden reasoning or full transcripts.

Use the [lesson template](../assets/lesson.md), then follow [Persist or revise](memory.md#persist-or-revise). If execution is blocked, retain that gap; do not promote an intended fix to a verified technical lesson.

Example of a bounded lesson from a synthetic episode:

> Two requests both observed one remaining seat and both booked it. A recorded regression failed before the fix and passed after capacity enforcement moved into an atomic database operation. When changing a shared capacity check followed by a write, inspect the database operation's atomicity and add a concurrent boundary scenario. An in-process mutex does not cover separate service instances. This does not prescribe a particular database or locking technique for unrelated systems.

An intentional test failure in TDD already has a workflow purpose. Capture an additional lesson only if it reveals a distinct development mistake or useful discovery, not merely because the test is red.

## Reassess a lesson

For a repeated failure, determine whether the earlier lesson was available, retrieved, applicable, acted on, and supported. These are different problems: missing retrieval calls for an activation/retrieval correction; inapplicable advice calls for tighter scope; a failed recommended action calls for new diagnosis. Repetition alone does not justify making the instruction louder or more general.

Use current requirements, source artifacts, and fresh evidence to resolve conflicting lessons. Do not decide by frequency or timestamp alone. Narrow or correct the affected entry; mark it `retired` when its recommendation has been disproved or superseded, retaining a short reason and replacement link where useful. Historical evidence remains historical; it must not be reported as a fresh check.

Promotion to a broader rule requires evidence for the broader conditions. Updating a shared skill or global instruction is a separate authoring task; a project memory update does not authorize it.

## Route a shared procedure defect

Distinguish a project-specific mistake from a defect in reusable guidance. Check whether the necessary rule/context was available, reachable through navigation, applicable and actually followed. An unavailable runner is not fixed by stronger prose; a local convention does not automatically belong in a shared skill. Do not assume the model is the cause before examining the observed context and tool limitations.

When evidence supports a shared-procedure defect, identify the existing rule owner and propose a narrow correction with a regression scenario. Preserve the raw task, contract and failure evidence; define evaluator-held expectations independently of the proposed wording. Include a new variant and an unaffected control where relevant. Group incidents only when their supported mechanism is shared.

When authoring in the skills source collection, use its shared evaluation procedure (`docs/skill-evaluation.md`). That repository reference is not a dependency of a copied or installed skill; elsewhere, use the project's available evaluation workflow while preserving the evidence and authority boundaries in this route.

Proceed with edits only when shared-skill authoring is already authorized; otherwise return the proposed correction and scenario without modifying shared or installed instructions. An authorized correction must be checked against the affected failure and relevant regressions, with result and execution evidence reported separately. Record missing proof instead of calling a successful explanation a verified workflow fix.

## Evaluate learning

Evaluate changed behavior, not the number of lessons written. Use representative decision checks to verify that recorded lessons change decisions on similar work.

For cross-session retention, give a fresh agent the task and access to the designated memory, without the original conversation or the expected lesson. Observe whether it locates relevant experience, checks applicability, and makes the intended preventive decision. A proposed check proves a decision change; only execution and its result can prove that the check or fix works.

To measure improvement, compare equivalent tasks with and without relevant memory under comparable model, tools, and effort. Include new variants of the failure class and unrelated control tasks. Keep evaluation cases separate from the examples used to write lessons; supply the agent only raw task inputs, not the evaluator's expected outcomes. Track repeated mistakes, successful task completion, inappropriate applications, and added effort. Do not infer causal improvement from a single successful replay or fabricated baseline.

Report evaluation levels separately: document integrity, skill-format validation, agent decision checks, executed engineering checks, and measured improvement. If only the first levels ran, state that practical improvement remains unmeasured.
