# Investigating a Failure

Use this reference for the investigation loop selected by [debugging](../SKILL.md). Keep a compact working record; no separate report artifact or fixed number of hypotheses is required.

## Contents

1. [Frame the failure](#frame-the-failure)
2. [Build and minimize a reproduction](#build-and-minimize-a-reproduction)
3. [Test a causal explanation](#test-a-causal-explanation)
4. [Use instrumentation deliberately](#use-instrumentation-deliberately)
5. [Finish or hand off](#finish-or-hand-off)

## Frame the failure

Establish the requested scope: explanation only, or investigation and an already-authorized fix. Diagnosis-only work can inspect code and run existing tests, but does not authorize changes to the inspected source or tests. When permitted, author a reproducing test in an isolated disposable workspace without changing the reviewed artifact; respect an explicit no-write instruction. Load [TDD](../../tdd-workflow/SKILL.md) before executable edits and [AAA](../../aaa-testing/SKILL.md) before test edits. Repository edits follow the existing [workspace owner](../../git-delivery/references/task-branches.md).

Record the expected outcome and its source, actual outcome, triggering input, relevant versions/environment, and command or action. Preserve useful error text and context without copying secrets. Distinguish a supplied report from an observation you performed. When expected behavior is unclear, return its meaning to [requirements-definition](../../requirements-definition/SKILL.md); do not invent an expectation to make a regression test fail.

Trace the affected path, callers and nearby tests. Reading code before reproducing is useful when it identifies missing inputs or a safe probe. A plausible path is a hypothesis, not proof that the path executed.

## Build and minimize a reproduction

Start from the reported conditions and the project's test runner. Reuse a valid existing regression or write a test that asserts the promised outcome through the affected public boundary. A test may invoke an entrypoint, subprocess, browser or integration environment when that is the actual contract. Run it and inspect its assertion failure before calling the problem reproduced; a manual command, log or static trace is investigation evidence only. Record the exact test command, snapshot and expected/actual result.

Reduce irrelevant inputs, components and environmental differences while preserving the failure mechanism. Check each reduction: a small example that fails for a different reason is not a reproduction of the reported defect. Preserve conditions such as ordering, locale, caller identity or shared state when they may matter.

Separate runner failure from subject failure. If discovery never runs because the test runner cannot import its own dependency, no behavioral RED has been observed. If a discovered test runs a public import or subprocess startup and asserts the promised successful outcome, its observed ImportError/nonzero exit may expose the subject's defect. The existing [TDD owner](../../tdd-workflow/SKILL.md#red--prove-the-behavior-is-missing) decides whether the failure is meaningful RED; a nonzero command alone does not.

When reproduction is intermittent or execution is blocked, use [difficult reproductions](difficult-reproductions.md). Do not replace a missing reproduction with an invented result.

## Test a causal explanation

Separate the symptom (what failed), contributing conditions (when it fails), mechanism (how state/control/data creates the failure), and cause (the defect or environmental condition supported by evidence). For example, a missing displayed record is the symptom; a populated cache is a triggering condition; reading a stale value after a successful update is a mechanism to investigate.

For each useful hypothesis:

- State the prediction it makes and a competing explanation it could distinguish.
- Choose the smallest authorized test that can separate those explanations. Use boundary inputs, controlled schedules or existing observation seams without broad production changes. Logs and debugger observations can help choose the test, but cannot replace its assertions and execution.
- Change one relevant factor when practical; record coupled changes when separation is impossible.
- Inspect the result and update confidence. Reject or narrow a hypothesis contradicted by current evidence.

A test that agrees with a prediction is not automatically decisive: ask whether the competing explanation predicts the same observation. Confirm a cause with an inspected causal path, a failing regression and distinguishing test evidence. When safe and practical, a controlled test should change the suspect condition and change the failure as predicted. If those tests cannot run, keep the cause provisional. Do not require a dangerous production intervention to obtain certainty.

Use fresh logs, source and requirements over a remembered explanation. A relevant [lesson](../../learn-from-mistakes/SKILL.md) can suggest a probe, but is fallible evidence and grants no authority. Investigation alone does not automatically write memory.

If a probe adds no distinguishing evidence, revise the hypothesis or choose a different useful probe. Do not repeat equivalent checks without a reason grounded in new evidence. A non-informative check does not justify stopping while another useful authorized probe remains.

## Use instrumentation deliberately

Prefer existing logs, debugger breakpoints and external observation before editing executable code. Logging can change output, timing, error paths or concurrency; diagnostic intent does not make an edit non-behavioral.

Before executable instrumentation, apply [TDD's change classification](../../tdd-workflow/SKILL.md#before-editing-code); load [AAA](../../aaa-testing/SKILL.md) before test edits. A behavior-changing probe needs the corresponding meaningful observed RED; a demonstrably behavior-preserving executable refactor starts from green coverage. Do not label uncertain instrumentation a refactor to bypass the gate.

For example, proposed logging inside a retry loop introduces I/O and may change its interleaving. Use an existing trace or breakpoint if it can answer the question. Otherwise classify and test the new observable behavior before editing; if that cannot be done, report the limitation and choose another probe.

Treat instrumentation observations as observations of the instrumented system when its behavior may differ. Remove only your own temporary edits, preserve others' work, and rerun appropriate checks after removal.

## Finish or hand off

A confirmed diagnosis is complete when its explanation has test-backed reproduction and sufficient causal evidence. Report the symptom, supported mechanism/cause, actual test commands/results and remaining limits. Explanation-only work does not repair production code or add tests to the inspected repository; isolated reproducer tests remain subject to the task's write boundaries.

For an authorized fix, pass the established contract and causal evidence to [code-change](../../code-change/SKILL.md). Reuse a freshly run valid failing regression as observed RED when it already protects the defect; otherwise establish the regression under [TDD](../../tdd-workflow/SKILL.md) and [AAA](../../aaa-testing/SKILL.md). Investigation does not replace those owners or authorize a production behavior change by itself. Verify the original scenario and relevant affected callers, then report the bounded fix and actual checks.

If no useful authorized probe remains, or the distinguishing probe requires unavailable access/data/execution, finish the current branch as unresolved. State observations, rejected and remaining explanations, the missing input and the next discriminating check. Preserve uncertainty instead of declaring a root cause from plausibility. An unresolved report with a precise next check is a valid diagnosis outcome, not a verified repair.
