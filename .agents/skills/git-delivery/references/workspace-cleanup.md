# Post-Merge Workspace Cleanup

Complete the local lifecycle of a merged task: return to a retained workspace, remove its eligible linked worktree, and delete its completed local branch. Apply this after confirming the task's merge or for an explicit cleanup request. Do not sweep unrelated branches or worktrees merely because they appear merged.

## Establish the exact task and merge

Inspect `git worktree list --porcelain`, local branch refs, and the task's merge evidence. Identify the task branch, its current commit, the exact worktree path, and the actual integration target. Use `main` or `master` only when it is the target; a stacked task may merge elsewhere. Preserve protected branches and workspaces still needed by another task or active agent. Honor instructions to retain the task workspace.

Use a verified target ref containing the merge. Fetch from the relevant remote when freshness matters and access is available; do not pull, reset, or rebase unrelated work to refresh it. Record the task and target commit IDs so later checks concern the same revisions.

For a normal merge or fast-forward, verify that the current task tip is contained in the target with `git merge-base --is-ancestor <task-tip> <target-tip>`. Exit status 0 proves ancestry; any other result needs investigation before deletion. An old merged commit does not prove that newer task commits are included.

Squash and rebase merges can integrate changes without retaining the original commit IDs. If ancestry fails, use authoritative merge records to verify the correct repository, source branch and merged source revision, target, and completed merge, then check for later local commits. A PR title, closed status, missing remote branch, or an empty file diff alone is insufficient. If integration or retained commits cannot be established, keep the branch/worktree and report the missing evidence. A normal deletion refusal is not permission to force-delete a squash/rebase branch.

## Check what removal would discard

Before changing the selected workspace, inspect:

```text
git -C <task-worktree> status --short --branch --untracked-files=all
git -C <task-worktree> status --short --ignored
git worktree list --porcelain
```

Preserve staged, unstaged, and untracked files. Inspect ignored files too: dependencies or generated output may be disposable, but local configuration, credentials, and other retained files must not disappear with the directory. Preserve the worktree when their disposition is uncertain. Do not auto-stash, clean, discard, or commit files to make removal possible.

Verify that no task or agent still uses the workspace, that it is not locked, and that the branch has no commits that must be retained. For a detached task worktree, verify its actual `HEAD` against the target rather than relying on a former branch name. Recheck the branch tip, target ref, worktree identity, and cleanliness immediately before removal; if they changed, reassess eligibility.

## Remove only eligible local resources

1. Choose a retained workspace and leave the branch being deleted. In the primary worktree, switch to the verified integration branch when clean and available, preserving the checkout directory. If the target is checked out in another worktree, use that retained workspace when available and inactive; do not force duplicate checkouts or disrupt an active owner. For a linked worktree being removed, move command execution outside its directory first.
2. Remove only the exact eligible linked task worktree with `git worktree remove <task-worktree>`. Never remove the primary worktree or use `--force` or recursive directory deletion to bypass a refusal. A removal failure is a reason to retain and report it.
3. Confirm that the task branch is no longer checked out in any worktree, then delete it with `git branch -d <task-branch>`. Keep the explicit target-integration check even if Git would allow deletion based on the branch's upstream. If deletion refuses, retain the branch and report why; do not fall back to `-D` without explicit authorization covering that exact branch and the commits it would discard.

For a branch-only task, skip linked worktree removal. If the worktree or branch is already absent, treat that resource as complete after checking the remaining task resources. Do not recreate it to perform cleanup.

Remote branch deletion is separate from local cleanup. Perform it only when requested or covered by an established repository policy and after verifying the exact remote and branch. Do not infer remote deletion authority from a local merge or deletion request.

## Verify and report

Re-read `git worktree list --porcelain`, the exact local branch ref, and `git status --short --branch` in the retained workspace. Confirm that removed resources are absent and the retained checkout is on the intended branch with its changes preserved. Report the merge evidence, removed branch/worktree, current workspace, and anything retained with its reason. Do not report complete cleanup if a resource remains blocked.
