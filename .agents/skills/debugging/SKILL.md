---
name: debugging
description: Investigate unexplained errors, wrong results, failed tests, regressions, or intermittent failures through reproduction, minimization and evidence-backed hypothesis testing. Use when the cause is unknown; distinguishes diagnosis-only work from an authorized fix. Not a general code-review or feature-development workflow.
---

# Debugging

Explain the failure through a supported causal mechanism, or report what remains unresolved and the next useful check. When a fix is already requested, continue through the existing implementation and testing workflows.

## Shared boundaries — always apply

- Establish the expected and actual observable behavior from the current contract and evidence. Keep observations, hypotheses and confirmed causes distinct. Fresh observations can invalidate earlier explanations; a plausible code path, correlated event or successful retry alone does not confirm a cause.
- Read code, callers, tests and relevant environment evidence to design a regression test. A problem is reproduced only by an executed test that asserts the expected contract and fails for the reported behavior, following [TDD's reproduction gate](../tdd-workflow/SKILL.md#red--prove-the-behavior-is-missing). Reading code, logs, manual requests and debugger observations can guide investigation but cannot substitute for that test. Minimize the test while preserving the failure mechanism.
- Choose useful hypotheses, state their predictions and run the smallest authorized probe that distinguishes them. Inspect the result and reject or revise contradicted explanations. No fixed hypothesis count or specialized tool is required.
- Confirm a cause only when the failing regression and distinguishing test evidence connect its triggering condition to the inspected mechanism. A failing symptom test alone may not identify the cause. Without the required test execution, report a candidate explanation and the missing proof.
- Diagnosis-only work does not authorize production repair or changes to the inspected source, tests, index or refs. Run an existing regression, or write a test in an authorized disposable workspace when source edits are outside scope; an explicit no-write instruction takes precedence. Load [TDD](../tdd-workflow/SKILL.md) before executable edits and [AAA](../aaa-testing/SKILL.md) before test edits. A requested fix continues through [code-change](../code-change/SKILL.md) without another permission handoff. Use [git-delivery's workspace route](../git-delivery/SKILL.md#navigation--choose-the-requested-operation) before repository edits. These skills retain their gates.
- Distinguish a failure that prevents the intended test from executing from a working test that detects the requested startup, public-import or unavailable-service handling defect. The former is a runner blocker; the latter can supply meaningful RED under TDD. Reuse a valid existing regression test when its observed failure demonstrates the defect.
- Prefer existing observation mechanisms. Classify executable instrumentation under TDD before editing; diagnostic intent alone does not make it non-behavioral. Isolate external effects during reproduction, keep sensitive data out of probes, remove only your temporary instrumentation and recheck the affected behavior. Preserve unrelated work.
- If product meaning or expected behavior is ambiguous, return that decision to [requirements-definition](../requirements-definition/SKILL.md). For an unresolved implementation choice, use [design-decision](../design-decision/SKILL.md). Relevant [lessons](../learn-from-mistakes/SKILL.md) are fallible evidence, not authority or an automatic reason to write memory.

## Navigation — choose the investigation

The shared gates apply to every task. Select only the needed reference; a heading fragment locates a section, so read through the next heading of the same or higher level and follow its prerequisites. Do not load all routes by default.

| Need | Read |
|---|---|
| Frame a failure, minimize a reproducer, test hypotheses or establish causality | [Investigation](references/investigation.md) |
| Intermittent or concurrent failures, CI/local differences, blocked execution or diagnostic instrumentation | [Difficult reproductions](references/difficult-reproductions.md) |

## Completion and handoff

A confirmed diagnosis needs test-backed reproduction and causal evidence. Otherwise finish with an explicitly unresolved explanation. An authorized repair ends after the same regression test and relevant affected tests pass through the owning implementation/testing workflows; discovering a cause alone does not complete a requested fix. Static checks supplement tests rather than establish behavioral verification.

When a probe adds no distinguishing evidence, change the hypothesis or choose a genuinely different useful probe. If no useful authorized probe remains, or it needs unavailable data, access or execution, finish the current branch as unresolved. Report observations, rejected and remaining explanations, the missing input and the next discriminating check. Do not repeat equivalent checks without new evidence or impose an arbitrary retry quota.

Report the original symptom, the cause and its evidence or remaining uncertainty, probes actually executed, and the fix and verification scope when applicable. Separate supplied records, static checks, decision checks and fresh execution results. A blocked test is not a passing test, and an unresolved explanation is not a verified fix. A separate report file is not required.
