# Task Branches and Worktrees

Before the first edit for a new task in a Git repository, work on a branch associated with that task. This applies to production code, tests, documentation, and configuration. Read-only discovery can establish the task and base first. Reuse an existing task branch/worktree for continuation; do not create another branch for each prompt or commit. An explicit task instruction can override this default.

## Discover the current workspace

Inspect applicable instructions, branch conventions, `git status --short --branch`, local branches, and `git worktree list --porcelain`. Establish which task owns the current branch and any existing changes. A branch name alone is insufficient evidence that unrelated edits belong to the new task.

Identify the base from project instructions or the requested target. It may be `main`, a release branch, or another task branch in a stack; do not assume every task starts from `main`. Use an existing verified local ref and state the chosen base. Fetch when remote freshness matters and access is available; do not automatically pull, rebase, or reset to make a base current. If the required base cannot be verified, report that specific gap before creating from an arbitrary ref. For a repository with no commits, use the project's initial-branch policy and explain that a normal worktree cannot yet be created from a committed base. Do not initialize Git merely to satisfy this skill in a non-Git directory.

## Name the branch

Project conventions, provider constraints, and task-specific instructions take precedence. Without an established convention, use:

```text
<type>/<short-description>
<type>/<task-id>-<short-description>
```

Use an English lowercase kebab-case description that identifies the outcome. Include a real supplied task ID when useful for traceability; preserve its canonical spelling. Choose a meaningful type such as `feat`, `fix`, `refactor`, `docs`, `test`, `build`, `ci`, or `chore`. Avoid invented ticket numbers and generic names such as `work`, `changes`, or `fix-stuff`. Namespaces such as a required agent or team prefix remain project-specific.

Examples: `feat/order-price-preview`, `fix/ENG-142-order-rounding`, `refactor/shared-refund-policy`, and `docs/api-contract-guide`. These names are a local convention, not part of the Conventional Commits specification. A branch describes the task; individual commits can have different types within that coherent task.

Validate the candidate with `git check-ref-format --branch <name>` and check for collisions in local and relevant remote-tracking refs. If the name exists, inspect its purpose and worktree before reusing it; do not overwrite, rename, or reset another task's branch. Choose a distinguishing descriptive suffix for a separate task rather than forcing creation.

## Choose a branch or a separate worktree

- **Current worktree:** use when it is clean or contains only verified changes for this task, and no other lane needs it. Create the task branch with `git switch -c <branch> <verified-base>`. A branch provides a separate line of history; it does not isolate the working directory or index.
- **Separate worktree:** use when the current workspace contains unrelated work, another task is active, or parallel work needs separate files and an index. Create it with `git worktree add -b <branch> <path> <verified-base>`. Choose a writable, unused path following the project's layout, preferably outside the source tree so the checkout does not become an untracked subtree. Do not modify ignore files merely to accommodate a speculative worktree layout.
- **Existing task workspace:** continue there after verifying the branch, path, base context, and changes. If its branch is already checked out in another worktree, use that workspace when authorized and not actively owned by another lane. Do not force the same branch into two worktrees; give independent tasks separate branches.

Uncommitted changes can carry across `git switch` when Git considers the switch safe; a successful switch does not prove isolation. A new worktree starts from committed content and does not copy another worktree's staged, unstaged, untracked, or ignored files. If this task depends on uncommitted work, explicitly account for it rather than assuming it is present. Choose a safe task-owned workspace or transfer only verified task-owned content; do not auto-stash, clean, discard, or commit unrelated changes to make setup succeed.

Worktrees share repository refs and configuration. They do not isolate databases, ports, external services, or runtime resources; inspect those only when the task needs them. Run commands inside the chosen workspace and use its project instructions, dependency setup, and test runner. Do not assume ignored configuration, dependencies, or locally installed integrations were copied into a new worktree.

## Verify setup and preserve it

After creation or selection, verify the actual branch, worktree path, and status. State the task branch and chosen base briefly, then continue the editing workflow. If isolation cannot be established safely, finish useful read-only discovery and report the concrete blocker before edits.

Leave the task workspace available for review until merge. Committing or pushing requires the corresponding task authorization; do not auto-merge into the base branch. Once the task's merge is confirmed, follow [post-merge cleanup](workspace-cleanup.md) to remove eligible local task resources and report anything retained. Use the same procedure for an explicit cleanup request; task completion alone is not merge evidence.

## Sources

[GitHub flow](https://docs.github.com/en/get-started/using-github/github-flow) recommends branches with short, descriptive names for changes. [Git check-ref-format](https://git-scm.com/docs/git-check-ref-format) defines valid branch names; [Git switch](https://git-scm.com/docs/git-switch) and [Git worktree](https://git-scm.com/docs/git-worktree) define creation and checkout behavior. The type/task-ID naming pattern and pre-edit workspace requirement are the user's workflow convention, not universal Git requirements.
