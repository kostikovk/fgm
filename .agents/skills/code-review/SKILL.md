---
name: code-review
description: Review a diff, pull request, commit range or working changes for correctness, task fidelity, regressions and relevant contract risks. Establish the reviewed snapshot and report actionable findings ranked by supported impact; review-only requests do not authorize fixes or publication.
---

# Code Review

Explain which supported defects the change introduces or exposes, where they occur and why they matter. Review task fidelity and applicable project standards alongside correctness; combine actionable findings by impact rather than treating either lens as sufficient alone.

## Shared boundaries — always apply

- Establish the requested artifact, comparison semantics and actual content before reviewing. Pin available base/target identities and distinguish committed, staged, unstaged and untracked content. Inspect relevant context and callers beyond the hunks. Missing history or content is a coverage gap, not evidence of an empty or safe change.
- Ground expected behavior in the current user contract and authoritative project sources. Commit prose and familiar patterns are clues, not replacements for that contract. Missing specifications limit fidelity claims; continue the correctness review supported by available evidence.
- A finding needs a concrete trigger, supported mechanism and meaningful consequence tied to the reviewed change or an applicable explicit rule. Distinguish an introduced/exposed regression from an unrelated pre-existing defect. Investigate candidates and suppress contradicted claims; unfamiliar structure, repeated shapes and preferences alone do not establish a defect.
- Verify executable-code issues through regression tests under [TDD's reproduction gate](../tdd-workflow/SKILL.md#red--prove-the-behavior-is-missing). Run an existing test or author an isolated test against the exact reviewed snapshot when authorized; load [AAA](../aaa-testing/SKILL.md) before test edits. A static trace, manual probe or documentation lookup can identify a candidate, but cannot establish that the issue was reproduced. If tests cannot run, report the candidate and verification gap separately from confirmed defects. Text-only rule/link violations use document validation.
- Review-only work preserves the inspected source/tests, index and refs. Isolated reproducer tests and ordinary disposable runner artifacts are acceptable when the task permits temporary writes; an explicit no-write instruction takes precedence. Inspect checks before running them; avoid live mutations, installation, automatic fixes, baseline updates and external comments/approval. State which snapshot a check exercised; worktree results do not prove staged or remote content was tested.
- Load relevant technical owners without starting their implementation workflows: [code-change](../code-change/SKILL.md) for reuse/contracts/principles, [frontend-code](../frontend-code/SKILL.md) for UI behavior, [go-code](../go-code/SKILL.md) for Go, and [AAA](../aaa-testing/SKILL.md) for tests. Unknown mechanisms can use [debugging](../debugging/SKILL.md) investigation within the same review-only boundary. No tool, framework or subagent topology is mandatory.
- When the user also requests fixes, complete review and proceed through [code-change](../code-change/SKILL.md), [TDD](../tdd-workflow/SKILL.md) before executable edits, [AAA](../aaa-testing/SKILL.md) before test edits and [Git workspace preparation](../git-delivery/references/task-branches.md). Review findings alone do not grant that authorization. Commits, merge and publication retain separate delivery boundaries.

## Review workflow

1. Read [Scope and evidence](references/scope-and-evidence.md); establish the change, intent, standards and coverage limits. Preserve separate layer identities even when presenting a combined review.
2. Read [Review method](references/review-method.md); trace changed behavior and affected consumers, investigate the highest-impact plausible failures and run appropriate safe checks when available. Depth follows the changed contracts and risk, not a fixed checklist or finding quota.
3. Read [Findings and reporting](references/findings-and-reporting.md); qualify and deduplicate candidates, rank supported impact and report actual coverage/checks. Recheck scope identity before completion; if content changed, update the affected review or state precisely which snapshot the report covers.

These three references supply the core procedure. Load conditional technical routes only when the affected behavior needs them; follow their prerequisites. A heading fragment locates a bounded section, not an automatic partial read.

| Additional need | Read |
|---|---|
| HTTP/API description and consumer compatibility | [API contracts](../code-change/references/api-contracts.md#keep-api-contracts-synchronized); identify authoritative source and affected consumers |
| Open product meaning or design choice | [Requirements](../requirements-definition/SKILL.md) or [design-decision](../design-decision/SKILL.md); report the missing decision without inventing it |
| Presenting a substantive review | [Developer review communication](../developer-readable-output/references/developer-responses.md#code-reviews); finding validity stays here |

## Completion

Finish when the requested scope has been assessed with enough evidence to qualify findings and describe material limits, or no useful authorized check remains. Findings come first, followed by actual coverage and verification gaps. If none are supported, say so for the inspected scope; do not imply complete correctness. Preserve the user's exact schema when one is required. A review report file is optional.
