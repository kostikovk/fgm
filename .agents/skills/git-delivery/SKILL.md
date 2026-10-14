---
name: git-delivery
description: Set up task branches or worktrees, prepare requested commits and pull requests, and clean up completed local task workspaces after a verified merge. Use for Git workspace setup, delivery, or post-merge cleanup; workspace preparation does not authorize commits or publishing.
---

# Git Delivery

Own Git workspace preparation, requested delivery, and post-merge cleanup; implementation and testing retain their existing workflows. Follow repository conventions first. Without established conventions, use task names such as `feat/order-preview` and English Conventional Commits.

## Shared boundaries — always apply

Execute already-authorized steps without asking again. Before repository edits, use a verified task branch/worktree and preserve unrelated working-tree and index changes. Continue the same task's workspace across follow-ups. Read-only requests do not require a new branch. Workspace setup does not authorize a commit or publication; a local commit request does not authorize push, remote PR creation, or history rewriting.

Keep the task workspace for review until merge. After a confirmed merge into `main`, `master`, or the actual integration base, complete safe local cleanup for that task without another confirmation, following [post-merge cleanup](references/workspace-cleanup.md). An explicit instruction to retain it takes precedence. Preserve work when the cleanup checks fail; remote branch deletion requires its own authorization.

## Navigation — choose the requested operation

Read this entrypoint, then load only the route needed below. A fragment identifies a heading; it does not make a file-reading tool select that section automatically. Locate the heading and read through the next heading of the same or higher level using bounded reads. If the tool cannot read sections, load only that reference file. Follow explicit prerequisite links; do not load every reference by default.

| Request | Read | Required context / next step |
|---|---|---|
| Start or continue an editing task; choose branch vs worktree | [Task branches and worktrees](references/task-branches.md) | Read the workspace procedure, then continue implementation; skip delivery procedures |
| Explain or choose a branch name only | [Name the branch](references/task-branches.md#name-the-branch) | Naming guidance only; creating/selecting a workspace requires the full workspace procedure |
| Suggest or format a commit message | [Commit messages](references/commit-messages.md) | Base the message on supplied evidence or inspected changes; do not stage or commit |
| Create a requested local commit | [Local commits](references/local-commits.md) | Inspect all staged and relevant working changes; load message guidance when writing the message |
| Draft or publish a requested PR | [Pull requests](references/pull-requests.md) | A draft needs no publication; remote creation requires the corresponding request |
| Finish a merged task or remove its completed branch/worktree | [Post-merge cleanup](references/workspace-cleanup.md) | Verify merge and task ownership, preserve retained work, then remove only eligible local resources |

For a combined request, follow the needed routes in dependency order: workspace setup before edits, local commit preparation before committing, and PR delivery after its branch/diff is ready. Report the result and actual verification evidence for the requested operation, including any remaining changes or blocker.
