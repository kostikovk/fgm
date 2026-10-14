---
name: tdd-workflow
description: Apply test-driven development before writing or changing executable code in any language. Use for features, bug fixes, behavior changes, and refactoring with AAA tests named should ... when ..., observed RED, minimal GREEN, then REFACTOR. Also classify test-only and non-behavioral edits before proceeding.
---

# TDD Workflow

Deliver verified behavior through small **RED → GREEN → REFACTOR** cycles. Load this skill before code edits or executable-issue verification; apply the appropriate branch below. This is a development discipline, not a replacement for the task's planning, review, or authorization workflow.

## Contents

1. [Before editing code](#before-editing-code)
2. [Test-backed issue verification](#test-backed-issue-verification)
3. [Navigation — optional supporting detail](#navigation--optional-supporting-detail)
4. [RED — prove the behavior is missing](#red--prove-the-behavior-is-missing)
5. [GREEN — implement the tested behavior](#green--implement-the-tested-behavior)
6. [REFACTOR — improve structure while preserving behavior](#refactor--improve-structure-while-preserving-behavior)
7. [Evidence and recovery](#evidence-and-recovery)
8. [Sources](#sources)

## Before editing code

1. Inspect the requirement, existing implementation, nearby tests, and actual test commands. Identify one observable behavior and its expected outcome. Keep a short list of remaining scenarios to guide the next cycles. Reuse the project's runner and test infrastructure; do not introduce a new dependency just to perform TDD.
2. Load **`$aaa-testing`** before writing or editing tests. Locate its `SKILL.md` through the available skills catalog or beside this skill in the installed skills directory, then read only the relevant language reference. AAA governs test structure and mandatory `should <observable behavior> when <specific condition>` scenario titles; this skill governs the order of development. For another language, apply AAA's shared rules using that project's native runner. If the required AAA skill is unavailable, report that missing dependency before code edits rather than silently skipping it.
3. Run the relevant existing tests to establish a baseline. Separate pre-existing failures from the behavior being changed. An unrelated broken baseline does not establish RED for the requested behavior.
4. Classify the change:
   - **New or changed behavior, including a bug fix:** follow the cycle below. An executed regression test must reproduce the bug before the fix.
   - **Behavior-preserving refactor:** first confirm relevant coverage is green. Add missing characterization tests before the refactor; these can pass against existing behavior. Do not manufacture a failing test for behavior that already exists. Then use the REFACTOR gate.
   - **Tests only:** use AAA and verify the tests; do not change production behavior unless requested. A new test for existing behavior may already be green.
   - **Documentation, skill instructions, formatting, or declarative configuration without executable behavior:** use the relevant validation instead of an artificial TDD cycle. A build script, migration, or configuration change that alters executable behavior belongs in the behavior branch and may require an integration or smoke test.

## Test-backed issue verification

For executable code, reproduce and verify an issue through an executed test with an independent assertion of the observable contract. Write a focused regression test, or reuse a valid existing one, and inspect its result against the affected behavior. Logs, code reading, manual probes, and user reports can guide investigation and test inputs; they do not establish verified reproduction or a verified fix.

A failing test must expose the reported defect rather than an unrelated environment or setup failure. An issue that cannot be reproduced by a test remains unverified; report the observed evidence and test-execution blocker without claiming reproduction. A passing scenario can show that the tested condition satisfies its contract, but does not disprove an issue under untested conditions. Diagnosis or review does not authorize production changes; test creation must also stay within the requested scope, and strict read-only work reports the missing execution evidence when no suitable existing test is available.

## Navigation — optional supporting detail

Classification and the applicable RED/GREEN/REFACTOR gates below are mandatory; navigation only selects supporting examples or setup guidance. A fragment identifies a heading, not an automatic partial read. Locate it and read through the next heading of the same or higher level with bounded reads. If section reads are unavailable, load only the relevant reference file. Follow prerequisite links; do not load every reference by default.

| Need | Read |
|---|---|
| Concrete cycle example and edge cases | [A small TDD cycle](references/red-green-refactor.md) |
| Revisit what qualifies as execution evidence | [What counts as evidence](references/red-green-refactor.md#what-counts-as-evidence); the phase gates below still apply |
| Configure persistent invocation in project instructions | [Persistent invocation](references/activation.md); classify runtime configuration changes before editing |
| Verify the origin of the cycle guidance | [Sources and adaptation](references/red-green-refactor.md#sources-and-adaptation) |

## RED — prove the behavior is missing

- Write one small test for the next behavior **before implementing that behavior**. Arrange independent inputs and dependencies, perform one conceptual Act, and assert the observable contract. Derive the expected result from the requirement, not the current algorithm.
- An existing failing test can supply RED when it precisely reproduces the requested defect. Run it and inspect the failure before the fix; do not duplicate valid regression coverage merely to create a new test.
- Run the focused test and inspect its failure. RED requires the test to be discovered and fail because the requested outcome is absent or incorrect. Syntax errors, wrong imports, missing packages, unavailable services, or unrelated setup failures are not behavioral RED.
- If a new subject cannot compile or be imported, write the test first, then add only the smallest signature, export, or explicit unimplemented placeholder needed to run it. Do not implement the requested outcome in this scaffold. Rerun and obtain the meaningful failure before GREEN.
- If the test already passes, check the assertion and input. Existing behavior may need no production change; otherwise strengthen the scenario to expose the missing requirement. Do not arbitrarily change the expectation to force RED.
- Capture the focused command and the relevant expected/actual failure in the work record or progress update. Do not claim RED merely because a test was written or because the process exited nonzero.

**Gate:** no implementation of the requested behavior before observed, meaningful RED. If execution remains unavailable after safe recovery attempts, report the blocker and leave the behavior unimplemented; elapsed time or an assumed failure does not satisfy this gate.

## GREEN — implement the tested behavior

- Write the smallest maintainable production change that satisfies the failing scenario. Avoid adding untested features or starting structural cleanup during GREEN.
- Rerun the same focused test, then the relevant existing tests. Fix regressions before moving on. Do not weaken assertions, skip the test, or mock away the subject to obtain a green result.
- Keep changes scoped to the requested contract. Each additional behavior needs its own RED cycle; do not write an entire implementation after the first failing test.

**Gate:** the new test and relevant regression tests pass before REFACTOR. A test that passes only in isolation while affected tests fail is not a completed GREEN phase.

## REFACTOR — improve structure while preserving behavior

- After GREEN, inspect for duplication, unclear names, or unnecessarily complex code in the touched scope. If no useful simplification exists, record that decision; do not create changes to demonstrate a phase.
- When cleanup is useful, state a small cleanup plan and keep the tested contract and assertions intact. Refactor production code or tests in small steps, rerunning relevant tests after each step.
- If an intended refactor changes observable behavior, return to RED for that behavior. If a refactor causes a regression, restore GREEN before continuing.
- Repeat the cycle for the next behavior. Keep the relevant suite green at cycle boundaries; complete applicable formatting, lint, typecheck, build, or integration checks before delivery.

## Evidence and recovery

Report the behaviors covered, files changed, actual RED failure, GREEN command/results, refactoring performed or deemed unnecessary, and final checks. Keep this concise; retain useful runner output rather than creating a mandatory report file. Distinguish commands run from suggestions and environmental blockers from test failures.

If implementation was started before RED, acknowledge the missed gate. Preserve existing user work and unrelated edits; do not automatically delete or rewrite code to manufacture a test-first history. Isolate or safely reverse only your own unverified behavior change, observe the test against the pre-change behavior, then continue. If that cannot be done safely, report the gap instead of claiming a TDD cycle.

## Sources

The RED/GREEN/REFACTOR sequence follows the primary sources linked in [references/red-green-refactor.md](references/red-green-refactor.md). AAA and the scenario naming convention are supplied by the user's `aaa-testing` skill. The persistent invocation rule implements the user's preferred development flow.
