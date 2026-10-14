---
name: go-code
description: Write, edit, refactor, or review production Go code with version-aware guidance on errors, context, concurrency, data access, API contracts, and performance. Use for implementation in .go files and Go-specific code questions; test-only tasks follow the project's testing conventions.
---

# Go Code

Produce correct, maintainable Go that fits the target repository and its supported Go version. For review-only requests, preserve the inspected source and tests while following the review evidence boundary below.

## Ground the change

Read applicable project instructions, `go.mod`, relevant callers, neighboring implementation, and tests. Preserve the repository's architecture, public contracts, dependency choices, and user-facing language. Do not import Finance Copilot's package layout, financial types, timezone, read-only mode, or tooling assumptions into another project.

Treat the `go` directive as the minimum supported language and standard-library baseline; the installed toolchain and `toolchain` directive do not raise that baseline. Inspect CI and build constraints too. A task to use a newer API does not by itself authorize a version bump or dependency migration.

Before running the first executable command—including the initial RED test or reproduction run—apply [Toolchain and runner detection](references/tooling.md#toolchain-and-runner-detection) to choose how Go, linter, and build commands are invoked in the workspace.

## Start implementation with tests

For every Go implementation task, load [code-change](../code-change/SKILL.md), [tdd-workflow](../tdd-workflow/SKILL.md), and [aaa-testing](../aaa-testing/SKILL.md), including AAA's Go reference, before executable edits. Apply code-change's workspace preparation and discovery first. Locate these files through the skill catalog or the linked paths when the tool has no skill invocation syntax; missing required guidance is a blocker rather than permission to skip it.

The first implementation step is to write the focused behavioral tests before production code. Reuse and run an existing valid regression test when it already covers the contract; do not duplicate it for ceremony. New or changed behavior requires observed meaningful RED before implementation. For a behavior-preserving refactor, write missing characterization tests and verify relevant coverage is green before production edits, following TDD's refactor branch. Test-only work uses AAA and executed tests. Issue reproduction and verification follow TDD's test-backed evidence rule.

For review-only requests, preserve the inspected production code and tests. Run suitable existing tests or authorized disposable regression tests in an isolated workspace; apply TDD and AAA before creating those tests. An explicit no-write or reading-only restriction forbids new tests. Without executed test confirmation, report a defect candidate and the verification gap rather than a confirmed issue or verified reproduction.

## Navigation — choose the technical concern

Grounding, supported-version constraints, implementation rules, and verification below always apply. A fragment identifies a heading, not an automatic partial read. Locate it and read through the next heading of the same or higher level with bounded reads. If section reads are unavailable, load only the relevant reference file. Follow prerequisite links; do not load every reference by default.

| Change | Reference section |
|---|---|
| Error handling and identities | [Errors](references/errors-and-safety.md#errors) |
| Nil/zero values, slice ownership, JSON null/empty | [Values and ownership](references/errors-and-safety.md#nil-zero-values-and-ownership) |
| Exact arithmetic / money, time, resource release | [Numeric correctness](references/errors-and-safety.md#numeric-correctness-and-money), [Time](references/errors-and-safety.md#time), or [Resources](references/errors-and-safety.md#resources) |
| Database access / transactions | [Database access](references/data-and-io.md#database-access); pgx v5 detail only when that dependency is present |
| HTTP, validation, response contracts | [HTTP and untrusted input](references/data-and-io.md#http-and-untrusted-input) |
| Context / lifecycle or concurrency | [Context and lifecycle](references/data-and-io.md#context-and-lifecycle) or [Concurrency](references/data-and-io.md#concurrency); read both when goroutine lifecycle is affected |
| Package placement / dependency boundaries | [Packages and boundaries](references/design-and-structure.md#packages-and-boundaries) |
| Interfaces, constructors/types, naming, or generics | The matching heading in [Design and Structure](references/design-and-structure.md) |
| Choosing or modernizing Go features | [Version-Aware Go](references/modern-go.md); establish availability before selecting a feature |
| Slow paths, allocations, benchmarks | [Performance](references/performance.md); establish evidence before choosing or measuring a change |
| Selecting the runner/toolchain manager, formatting, lint, CI, or toolchain failures | [Toolchain and runner detection](references/tooling.md#toolchain-and-runner-detection), [Verification choices](references/tooling.md#verification-choices), [Version failures](references/tooling.md#version-failures), or [Handling findings](references/tooling.md#handling-findings) |

## Implement the contract

- Handle errors where their outcome is decided. Add useful context, preserve identities that callers rely on, and distinguish absence, invalid input, and infrastructure failure.
- Pass caller context through blocking operations. Bound external work, release resources on every path, and give goroutines an owner, termination condition, and error policy.
- Preserve slice ownership, JSON null/empty semantics, result ordering, rounding, and time-zone behavior. Change these deliberately with contract tests rather than as incidental cleanup.
- Validate external data before it becomes trusted state. Use exact arithmetic where precision is part of the contract; missing data must remain distinguishable from a legitimate zero.
- Extend existing boundaries and utilities before adding abstractions. Prefer a suitable standard-library API, while retaining working dependencies and older idioms when compatibility or behavior requires them.
- Keep implementation and verification proportional. A refactor preserves observable behavior; a performance change needs a measured benefit and correctness evidence.

## Verify and report

Run focused behavior tests, format changed files, then the applicable repository checks. Run race tests for affected concurrent code and broader checks required by CI. Use [tooling.md](references/tooling.md) to choose commands and diagnose unavailable prerequisites; do not assume PostgreSQL or optional tools exist.

Inspect the final diff for lost error handling, resource leaks, accidental API/version changes, and unrelated edits. Report changed files, observable behavior, checks actually run, and any validation gap. Do not claim a suggested command passed.

## Sources

Adapted from the supplied Finance Copilot `go-code` skill, which credits [spf13/go-skills](https://github.com/spf13/go-skills) and [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang). Reference links to official documentation substantiate technical guidance; project-specific policies are discovered in the target repository.
