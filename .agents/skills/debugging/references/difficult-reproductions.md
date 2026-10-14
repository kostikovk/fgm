# Difficult Reproductions

Use this reference when the ordinary [investigation loop](investigation.md) cannot produce reliable execution evidence. Preserve its authorization, causal-evidence and stop gates.

## Contents

1. [Intermittent failures and concurrency](#intermittent-failures-and-concurrency)
2. [CI and local differences](#ci-and-local-differences)
3. [Setup failures and unavailable execution](#setup-failures-and-unavailable-execution)
4. [Reproduction with external effects](#reproduction-with-external-effects)

## Intermittent failures and concurrency

Capture inputs and ordering, not just failure frequency. Investigate clock/randomness, shared-state ownership, concurrent actors, resource lifetime and cancellation when the symptoms suggest them. Record seeds, event order or request identifiers where available without exposing secrets.

Use existing clock/randomness seams, barriers, hooks or controlled schedules inside a regression test to force a suspected interleaving when practical. Assert the promised invariant and run the test; a two-actor trace alone only suggests the mechanism. A test in which both actors observe capacity before either writes can distinguish a shared-state race from a bad single-thread calculation.

Repeated runs can characterize frequency and gather evidence; green reruns alone do not prove absence or establish a fix. Avoid adding arbitrary sleeps as causal proof. A single-process schedule may demonstrate a mechanism while leaving cross-process behavior unverified; state that limit.

Executable test or instrumentation edits still use [TDD](../../tdd-workflow/SKILL.md) and [AAA](../../aaa-testing/SKILL.md). Do not introduce a new production seam merely to make investigation convenient without an authorized, tested change.

## CI and local differences

Compare the failing CI command and actual configuration with local execution: relevant language/dependency versions, operating system, working directory, paths/case sensitivity, locale/timezone, environment settings, test selection/order, caches and resource isolation. Inspect only differences relevant to the observed failure; avoid dumping credentials or an entire environment.

A local pass establishes success under those local conditions, not invalidity of the CI report. Preserve the CI artifact and its provenance. Reproduce one candidate difference at a time when safe; a configuration delta is a hypothesis until a probe connects it to the failure.

If CI artifacts are missing, identify the exact log, command or configuration needed for the next distinguishing comparison. Do not change unrelated code to silence an unavailable CI failure.

## Setup failures and unavailable execution

Identify which boundary failed: runner discovery, test setup, execution of the subject, or assertion. A missing runner dependency with no discovered test blocks observed behavioral RED. A working test asserting startup or public-import success may expose an application startup defect. A working test asserting controlled handling of an unavailable dependency may expose an uncaught exception; the dependency error is then part of the subject's behavior, not necessarily a runner problem.

Perform safe, authorized recovery using existing project instructions; do not assume permission to install dependencies or change global configuration. A local environment defect can itself explain a reported environment failure, but it does not prove a separate reported application defect.

When test execution is unavailable, inspect source and supplied artifacts and label conclusions as provisional. Do not claim the problem reproduced, the cause confirmed or behavior verified from logs, manual probes or source alone. Report supplied test results with their provenance rather than as your fresh execution. Keep the requested behavior unimplemented when the required TDD gate remains blocked after safe recovery, following [the existing gate](../../tdd-workflow/SKILL.md#red--prove-the-behavior-is-missing).

Stop the current branch as unresolved when the useful distinguishing observation requires inaccessible trace/data/access. State what is missing and the next probe; do not fabricate output or repeat equivalent unavailable checks.

## Reproduction with external effects

Determine whether the proposed path reaches real charges, emails, destructive writes or another live effect. Diagnostic scope does not authorize that effect. Use an existing isolated substitute, disposable local data, dry-run interface or project test environment when it preserves the relevant mechanism and is authorized.

A test using a substitute proves only what it models. For example, an assertion against an isolated payment stub can reproduce duplicate invocation while leaving provider idempotency unverified. Calling the stub manually is not a regression test. Report that boundary instead of claiming complete production reproduction.

If no authorized isolated path can answer the question, stop that probe and describe the missing safe environment or artifact. Continue other useful read-only probes when available.
