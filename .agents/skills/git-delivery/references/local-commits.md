# Local Commits

Use this procedure for a requested local commit. Preserve unrelated work; do not infer permission to push, publish a PR, or rewrite history.

## Inspect the delivery scope

Read applicable project instructions and identify the requested commit scope. Inspect the branch, working-tree status, staged diff, unstaged diff, and relevant untracked files. Before the first commit, inspect the index directly; there is no `HEAD` to compare against. A branch comparison alone omits local changes.

If additional repository edits are needed, first apply [task workspace preparation](task-branches.md). Committing existing verified changes does not require creating a second task branch.

## Choose a coherent commit boundary

Group changes by one purpose that can be explained, reviewed, and reverted together. Include the tests, contracts, generated artifacts, and documentation needed for that purpose. Do not split a behavior change from its regression coverage merely to separate file types, or combine unrelated changes just because they are present.

For an initial repository import, a single bootstrap commit can be appropriate. Inspect the included files and omit local runtime state, caches, credentials, and accidental generated output. Follow existing ignore rules; add a narrow ignore entry when relevant rather than hiding arbitrary files.

Stage explicit paths or selected hunks after inspection. Review pre-existing staged changes as well as your own edits. Preserve unrelated user work in both the working tree and index; do not silently include, unstage, reset, or discard it. When mixed changes cannot be separated safely, report the specific ambiguity before committing. Avoid broad `git add .` or `git add -A` when unrelated work may be present.

## Verify the staged content

Review `git diff --cached --stat` and the full staged diff, including new files, deletions, and generated artifacts. Confirm the index represents the intended commit, and run `git diff --cached --check` for whitespace errors.

Run the checks appropriate to the staged change and project instructions. Code changes use the repository's relevant tests and required checks; skill instructions and documentation use format, link, and content validation. Existing results are useful only when they still cover the final staged content. State environmental blockers and pre-existing failures explicitly; never claim a command ran when it did not.

Write the message using [commit message guidance](commit-messages.md). Message conventions have one owner; do not duplicate them here.

## Commit and check the result

Commit the reviewed index once the requested scope is verified. Avoid passing file paths to `git commit` after selecting hunks: that form can commit the current file contents instead of the selected index contents. Respect hooks and signing requirements; do not bypass them with `--no-verify`, disable signing, or alter identity to force success. If a hook modifies files or fails, inspect the result, make authorized corrections, and revalidate the affected content before retrying. A retry must not silently expand the commit.

After a successful commit, inspect its hash, subject, and included paths, then check the remaining working-tree/index status. Report the commit and relevant checks; explain any remaining changes instead of assuming the repository is clean. Do not amend, reset, rebase, or force-push unrelated or shared history as part of routine delivery.

## Sources

[Git add](https://git-scm.com/docs/git-add), [Git diff](https://git-scm.com/docs/git-diff), and [Git commit](https://git-scm.com/docs/git-commit) document index inspection, selective staging, message files, and hooks. Atomic boundaries and proportional verification encode the user's workflow preferences.
