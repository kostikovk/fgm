# Findings and Reporting

## Qualify a finding

A finding should give the author enough information to recognize and correct a supported problem:

- **Location and artifact:** precise path and a small relevant line span/symbol in the reviewed version; anchor to the changed cause or newly exposed path. Label index/old-side/remote locations when they differ from the worktree. Do not invent inspected line numbers.
- **Trigger:** reachable input, state, ordering or consumer condition that activates it.
- **Mechanism and consequence:** explain what the changed code does and the resulting incorrect outcome, broken contract or applicable documented-rule violation. Cite the contract/evidence near the claim.
- **Correction direction:** the outcome or safeguard needed, when useful. A finding need not prescribe a rewrite or supply a patch.

Use a concise problem title and explanation. For example: “Older search failure replaces the current query's results: after query B succeeds, query A's late rejection still sets the shared error state because the removed effect lifetime check no longer blocks publication.” Attach an actual inspected location in a real review; this example does not establish one.

For executable-code defects, include the reproducing test command, reviewed snapshot and relevant assertion failure. Link supplied test records with their provenance and distinguish them from fresh execution. Without executed test evidence, keep the candidate and missing verification separate from confirmed findings; static reasoning alone does not qualify behavioral reproduction. Text-only violations retain their document evidence. Do not pad the report with every theoretical failure. Confidence is the strength of evidence, while severity is the consequence if the supported scenario occurs.

## Rank and consolidate

Follow the project's or user's requested convention first. Otherwise use this impact-based priority vocabulary:

| Priority | Meaning |
|---|---|
| P0 — critical | Immediate release/operation blocker with demonstrated broad destructive or severe security impact; the evidence supports urgent handling. |
| P1 — high | Serious supported correctness, security or availability failure that should be fixed before delivery; state any triggering conditions. |
| P2 — medium | Bounded meaningful defect or contract violation that needs correction under the demonstrated conditions. |
| P3 — low | Minor actionable problem or applicable documented-rule violation with limited impact; preference alone does not qualify. |

Do not treat priority as a confidence score or force every review to include high severity. State uncertainty rather than promoting it into an alarming confirmed finding. Use the demonstrated reach and consequences to order findings; correctness and fidelity belong in the same ranked result.

Deduplicate notes that share one cause and correction; include their distinct supported consequences in that finding. Keep independent defects separate even if they occur in one file. Report all material findings without a fixed item/word quota. Suggestions and open questions can be separate when requested/useful; they must not dilute or masquerade as defects.

## Report coverage honestly

Put findings first, then summarize the actual artifact/comparison, assessed paths/contracts, checks and material limits. Preserve the required fields/schema when the user supplies one; presentation guidance in [developer-readable-output](../../developer-readable-output/SKILL.md) does not override it. Review comments, approval, commits and fixes require the corresponding authorization; a local report does not imply external publication.

If no finding qualifies, state “No supported findings in the inspected scope” with the scope and meaningful verification gaps. Do not infer deployment readiness, approval or complete safety from absence of findings. A partial or blocked review should clearly name what was unassessed instead of appearing complete.

Separate supplied logs, static reasoning, fresh checks and proposed checks. Include failing checks with attribution/remaining uncertainty, and blocked checks without calling them passed. If a separate verdict is requested, apply its defined meaning and support it with the scope/evidence; do not manufacture an unconditional approval convention.
