---
name: code-change
description: Ground production code changes in existing behavior before implementing a feature, fixing a bug, or refactoring. Find reusable functions, components, hooks, services, and library capabilities; choose reuse, compatible extension, extraction, or new code using practical KISS, YAGNI, DRY, and SOLID criteria. Use for implementation tasks; design-only decisions and review-only requests keep their own workflows.
---

# Code Change

Make the smallest maintainable change that fulfills the requested contract and fits the repository. This skill supplies implementation decisions; `tdd-workflow` supplies development order, `aaa-testing` supplies test structure, and applicable language/framework skills supply technical details.

## Establish the contract

Read applicable project instructions, the affected implementation, callers, nearby tests, and relevant dependency/version constraints. State the observable outcome being added or preserved. Separate known behavior, intended behavior, and assumptions that could change the solution.

Keep discovery proportional to the change: trace the affected path and a useful precedent. Clarify only missing requirements that materially change the result. Use the repository's vocabulary and architecture.

When product intent, behavioral rules or scope remain unclear, use [requirements-definition](../requirements-definition/SKILL.md) to establish the affected contract. When the request needs multiple deliverable scenarios or coordinated technical tasks, use [task-decomposition](../task-decomposition/SKILL.md) for that breakdown and the task's readiness/dependency checks. Continue directly for a clear bounded change; do not require a separate PRD or backlog. A stale or blocked task contract needs reassessment before its dependent implementation.

Before the first repository edit, including tests or documentation, apply the [task branch/worktree policy](../git-delivery/references/task-branches.md) owned by `git-delivery`. Reuse a verified workspace for the same task or create one from the appropriate base; then run discovery and checks in that workspace. This preparation does not authorize commits or publication.

## Find existing behavior before adding code

1. Search for the capability by domain terms, related symbols, and behavior. Check relevant functions, components, hooks, services, policies, and existing library APIs. Use available symbol/graph tools when helpful, with ordinary repository search as a fallback.
2. Read promising implementations, their callers, and tests. Similar names or signatures do not establish equivalent behavior. Compare input assumptions, output and error semantics, state ownership, side effects, dependency direction, authorization, and transaction/concurrency guarantees where relevant.
3. Choose the smallest sufficient option:
   - **Reuse:** the existing contract satisfies the requirement.
   - **Extend:** a bounded change preserves existing callers' contracts and adds the required capability.
   - **Extract:** existing and new callers share a coherent operation or rule that belongs behind one contract.
   - **Create:** the inspected candidates cannot satisfy the requirement without mixing independent responsibilities or breaking useful boundaries.
4. Record the deciding evidence briefly: actual path/symbol, relevant behavior, and the chosen option. If no suitable candidate was found, state the inspected area and the concrete gap. Inaccessible code is not evidence of absence.

An unsafe precedent needs correction, not blind reuse. Consider all affected callers when changing a shared mechanism; an endpoint-only fix may leave the same defect elsewhere.

## Navigation — choose the needed detail

A fragment identifies a heading, not an automatic partial read. Locate it and read through the next heading of the same or higher level with bounded reads. If section reads are unavailable, load only the relevant reference file. Follow prerequisite links; do not load every reference by default.

| Situation after discovery | Read | Prerequisite |
|---|---|---|
| A shared operation needs extraction or a wrapper may be redundant | [Extract shared knowledge deliberately](references/reuse-and-principles.md#extract-shared-knowledge-deliberately) | Compare existing implementations and callers first |
| Choosing or evaluating implementation structure | [Apply principles to the observed problem](references/reuse-and-principles.md#apply-principles-to-the-observed-problem) | Apply relevant criteria to the observed change; read extraction guidance too when extracting |
| Adding or changing an API | [Keep API contracts synchronized](references/api-contracts.md#keep-api-contracts-synchronized) | Identify routes, contract source and consumers before changing behavior |

Use the applicable principle criteria for every implementation choice; the section is separate so its catalog need not be loaded for discovery or workspace setup alone. Sources for these criteria remain in [reuse-and-principles.md](references/reuse-and-principles.md#sources).

## Implement through the appropriate test branch

Load `tdd-workflow` before executable edits and `aaa-testing` before test edits. Locate them through the available skill catalog or alongside this skill. Load the relevant language/framework skill when available. If a required testing skill cannot be found, report that dependency before executable edits.

Use tests to reproduce issues and verify executable behavior through TDD's gates. Source inspection, logs and manual probes can guide the test; they cannot replace an executed regression or establish a verified fix. Static checks supplement behavioral tests. Documentation-only changes use document validation.

When extraction is necessary before a feature, verify relevant coverage is green, add missing characterization coverage, extract while preserving behavior, and rerun the affected tests. Then use observed behavioral RED for the new capability and minimal GREEN. A behavior-changing extension to a shared function needs RED for that change and regression coverage for existing consumers. Follow the testing skills for blocked execution and other change classifications; do not duplicate their full protocol here.

After GREEN, simplify the touched scope if useful. Keep new behavior and structural cleanup in distinct, verified steps. Run the repository checks appropriate to the affected contract, and inspect the final changes for accidental public API, dependency, or unrelated behavior changes.

## Report the result

Briefly report what existing mechanism was reused, extended, or extracted, or why new code was necessary; the observable result; and checks actually run with any validation gap. Evidence belongs in the working update or final response; a separate report file is not required. Commits, PRs, and project configuration remain separate tasks unless requested.
