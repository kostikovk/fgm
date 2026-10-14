# Scope and Evidence

## Choose the comparison

Inspect applicable instructions, Git state, the user's request and available PR metadata. Establish which artifact the user wants evaluated before choosing a command. Read-only review does not require switching branches or creating a task branch.

| Requested artifact | Evidence / comparison |
|---|---|
| Change between two exact commits, tags or versions | Resolve both to immutable commit IDs; compare their trees with `git diff <from-id> <to-id>`. An explicit endpoint comparison is not automatically three-dot. |
| A topic's contribution to a PR | Verify target branch/base and topic head from actual metadata/history. Use the appropriate merge-base comparison for contribution (`git diff <base-id>...<head-id>`), or the provider's supplied review snapshot. State the semantics and identities; contribution review alone does not test the eventual merge result. |
| A particular commit | Inspect its parents and intended scope. A merge commit may mean a first-parent change, another parent comparison or merge resolution; inspect/clarify that distinction rather than choosing silently. |
| Staged changes only | `git diff --cached` and index versions, including added/deleted files. Worktree file reads may show different code. |
| Unstaged changes only | `git diff` compares index to worktree; untracked files need separate discovery if included by the request. |
| All local working changes | Inspect status, staged diff, unstaged diff and relevant untracked files separately; include branch commits only when requested. A net `git diff HEAD` is useful context but can conceal an index change canceled in the worktree. |
| Supplied patch or remote snapshot only | Review available content; obtain relevant context through available read access when useful. Do not pretend it is the checked-out tree or assume missing files contain a required safeguard. |

For example, an index removes a guard and the worktree restores it: the staged candidate and current worktree are different artifacts. Cover both for an all-local request, label any staged finding accordingly, and do not claim the restored final worktree still has the defect. A staged-only review must assess the index version.

Record base/head IDs, layer/content identity and the selected scope concisely. Verify refs and available history before invoking comparisons. Missing requested refs, shallow history, ambiguous merge bases or competing plausible scopes require resolving the missing evidence or reporting the gap. Choose a base from instructions/history when unambiguous; do not invent `main` or force a question when inspection answers it. Empty committed diff does not rule out local changes. Do not silently fetch, reset, checkout or rewrite refs to obtain a cleaner review environment.

## Inspect the actual artifact

Start with changed-file/status summaries, then read the relevant complete functions, callers and tests. Account for renames/deletions, executable configuration, generated contracts and lockfile changes when they affect the behavior. Git does not include untracked file content in ordinary diffs: discover relevant new files through status and read them separately. Avoid indiscriminate reads of ignored dependencies, caches, secrets or unrelated artifacts.

For index content, use index-aware reads such as `git show :path`; for committed content, read the pinned tree. A line number in the worktree may not locate the same code in the index. Label finding locations by reviewed artifact; deleted code may need an old-side location and an explanation of the removed safeguard.

When binary files, submodules, truncated patches or unavailable generated sources prevent inspection, state what could and could not be assessed. A submodule pointer alone does not reveal the nested change. Do not report a subsystem reviewed merely because its path appeared in a stat summary.

If the candidate changes during review, refresh the affected evidence and checks or report the exact earlier snapshot; never merge conclusions from unidentified versions into a current verdict.

## Establish intent and applicable standards

Use the current user request, linked requirement/issue/spec, authoritative API contract, project instructions and relevant tests to establish expected outcomes. Track their provenance and conflicts. Commit messages and branch names help discovery but do not override a current task contract. Existing tests describe exercised behavior and can themselves be incomplete or wrong.

When no spec is available, assess reachable correctness and compatibility from supported contracts, and label fidelity to missing product requirements unassessed. When sources disagree materially, describe the decision needed; do not invent omitted functionality or declare a deliberate supported change a regression because it differs from the base.

Apply documented standards within their actual scope. Technical criteria come from the relevant owner, not a generic smell catalog. A documented rule violation can be actionable even without a runtime failure, but cite the rule and affected change; distinguish that from personal taste. Avoid repeating formatter/linter noise when authoritative checks already provide it; retain relevant check failures and consequences in the report.
