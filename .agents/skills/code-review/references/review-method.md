# Review Method

## Follow the changed contract

Trace the path from input or event through the changed behavior to its observable outcome. Compare base and candidate assumptions: outputs/errors, state ownership/lifetime, side effects, authorization, persistence, transaction/concurrency guarantees and consumer compatibility as relevant. Include affected callers outside the diff when a changed shared mechanism alters their contract.

Review whether the requested behavior is implemented, omitted or contradicted by the change. Report unintended scope only when the contract or a concrete resulting risk supports it; a necessary supporting change is not automatically scope creep. Missing intent is an uncertainty to resolve, not permission to design a new feature.

Choose depth from what changed. Authorization boundaries, destructive operations, concurrent state, public contracts and failure recovery justify targeted tracing; a bounded textual correction usually does not require a system-wide audit. Inspect existing safeguards before claiming their absence. Changes can expose an existing helper defect through a new supported path even when the faulty line is unchanged.

## Use technical owners conditionally

| Concern | Owner / how to apply |
|---|---|
| Reuse, extraction, responsibility boundaries or API predictability | [code-change](../../code-change/SKILL.md) and the relevant principle/contract reference. Compare semantics and consumers; similar syntax alone does not require extraction. |
| UI state, effects, requests, accessibility or browser mechanisms | [frontend-code](../../frontend-code/SKILL.md); select the affected concern and installed framework/runtime route. |
| Go errors, context, concurrency, ownership or data access | [go-code](../../go-code/SKILL.md); select relevant concern/version details. |
| Test contract, independence or assertion quality | [AAA](../../aaa-testing/SKILL.md) and applicable language reference; missing comments alone are not a defect. |
| Unexplained failure needing discriminating evidence | [debugging](../../debugging/SKILL.md); investigation remains within the review's authority. |

Reading an implementation skill supplies criteria; it does not turn review into editing. TDD governs authorized executable changes, not a demand that every reviewed historical patch prove its development sequence. A missing regression test can be a finding when it leaves a specific required behavior or safeguard unsupported; do not request tests solely to mirror the implementation or inflate coverage.

## Investigate candidates before reporting

For each plausible defect, identify the trigger, mechanism and consequence. Trace applicable guards, consumers and error mapping; compare base behavior and the current intended contract. Use static invariants and version-matched primary documentation to design the smallest regression test. For executable-code issues, run a test that asserts the intended contract against the reviewed snapshot and inspect its meaningful failure before reporting reproduction. Static evidence or a manual demonstration alone leaves a candidate unverified; report it separately with the missing test evidence. Text-only documented-rule violations use document checks.

Disprove your candidate where possible: can the input reach this path, is the invariant guaranteed upstream, is a transaction covering the apparent race, or is the behavior deliberately changed? If that evidence contradicts the candidate, drop it. For security/performance claims, establish the relevant access/input boundary or workload and supported consequence; do not escalate severity from a scary label or unsupported possibility.

A pre-existing defect unrelated to the change is outside a change-review finding unless the user requested broader assessment. Mention a material separate concern with its scope if useful; do not attribute it to the patch. If newly reachable behavior exposes an old defect, anchor the finding to the changed entry path and explain that connection. Group manifestations of one fixable cause without hiding independent failures.

## Verify the right snapshot safely

Inspect project scripts/configuration before running focused tests. Reuse an existing regression when it exposes the candidate; otherwise author a test in an authorized disposable copy of the exact snapshot, loading [TDD](../../tdd-workflow/SKILL.md) and [AAA](../../aaa-testing/SKILL.md) first. Tests provide behavioral verification; typecheck/lint/build checks supplement them. Avoid deploy hooks, live writes, downloads, automatic formatting, snapshot updates and commands that alter the candidate. A command named `test` is not proof it is safe. Report missing dependencies instead of letting a package runner install them implicitly.

Record the command, outcome and artifact it exercised. Worktree tests do not execute staged or remote versions by implication. A disposable copy of the intended snapshot can provide appropriate execution when preparation is permitted and available; preserve the reviewed source/index/refs, inspect setup scripts and state any configuration/dependency differences. Do not switch or mutate the user's candidate to manufacture this evidence.

Passing tests support their exercised scenarios, not all contracts. A failure establishes a patch defect only after linking it to the reviewed change and expected outcome; runner/setup problems and unrelated baseline failures remain separate. Compare a relevant baseline when needed and feasible to determine attribution. Do not claim passing checks from supplied logs as fresh execution, or claim TDD RED/GREEN from a review reproducer.

Stop expanding investigation when evidence sufficiently qualifies or rejects the candidates in scope. If a necessary check is unavailable, report the limit and precise unresolved question. Repeating an equivalent probe without new evidence does not improve the review.
