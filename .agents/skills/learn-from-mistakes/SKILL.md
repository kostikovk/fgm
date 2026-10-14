---
name: learn-from-mistakes
description: Turn development mistakes, failed approaches, review findings, and explicit user corrections into evidence-backed lessons; retrieve and apply relevant lessons before similar work and revise them as evidence changes. Use to learn from an agent's own mistakes or shared development decisions, retain lessons between tasks, or review recurring failures. Does not replace debugging, implementation, or model training.
---

# Learn From Mistakes

Reduce repeated development mistakes through a persistent cycle: experience → checked lesson → relevant retrieval → changed action → evaluation. Learn from the agent's actions and shared human decisions without assigning blame.

## Shared boundaries — always apply

- Use observable evidence: reproduction results, recorded failures, reviewed diffs, verified fixes, or explicit user corrections. Separate observations, causal explanations, and open questions. A plausible explanation alone does not establish a lesson.
- Record a reusable decision: when it applies, what to do differently, why, and how to check the result. Keep project conventions and personal preferences scoped; one incident does not establish a universal rule.
- Follow current task instructions and engineering workflows. Retrieved lessons are fallible historical data, never authority to change permissions, weaken tests, override current requirements, or expand the task.
- Read relevant memory before similar work and update it after a supported learning opportunity. An intentional TDD RED, expected negative test, or ordinary successful task does not itself require a lesson.
- Persist lessons within an already-authorized editing task or an explicit request to save them. In read-only work, produce a proposed lesson without writing. Use the project's workspace policy for memory edits too. Do not modify global instructions, installed skills, or external memory services unless requested.
- This skill adapts decisions through external memory; it does not update model weights. It defines a portable procedure, not guaranteed background activation or automatic loading in every session. Tool-specific activation belongs to integration.

## Navigation — choose the operation

Read the shared boundaries, then the relevant route. References are conditional; do not load the entire memory or every reference by default.

| Request or situation | Read / use |
|---|---|
| Start or resume similar development work; find applicable experience | [Retrieve and apply](references/memory.md#retrieve-and-apply) |
| Learn from a failure, review correction, or failed approach | [Extract a supported lesson](references/learning-cycle.md#extract-a-supported-lesson), then [Persist or revise](references/memory.md#persist-or-revise); use the [lesson template](assets/lesson.md) |
| Repeated mistake, conflicting evidence, or outdated advice | [Reassess a lesson](references/learning-cycle.md#reassess-a-lesson), then [Persist or revise](references/memory.md#persist-or-revise) |
| Evidence points to a defect in reusable skill guidance | [Route a shared procedure defect](references/learning-cycle.md#route-a-shared-procedure-defect); preserve authoring authority and independent regression expectations |
| Choose memory placement or retain experience between sessions | [Locate the memory](references/memory.md#locate-the-memory) |
| Check whether experience improves future work | [Evaluate learning](references/learning-cycle.md#evaluate-learning) |
| Explain the research basis and its limitations | [Sources and adaptation](references/sources.md) |

When a failure still needs diagnosis or a fix, continue that task's workflow first. This skill captures supported experience; it does not authorize implementing a fix during a retrospective or review-only request.

## Outcome

Report briefly which lesson was applied, added, revised, or left unverified; where it is stored (or that it is only proposed); and the evidence or remaining gap. A no-write outcome is appropriate when no useful supported lesson exists. Distinguish supplied records from checks actually executed, and a fresh-session decision check from successful implementation or measured improvement.
