# Pull Requests

Use this reference for a requested PR description or remote PR delivery. A description draft requires no publication.

## Draft the description

Inspect the intended branch diff and use the repository template. Lead with the concrete problem and resulting behavior; include relevant validation and material limitations. Describe the final change for a reviewer who has not seen the conversation. Keep scope, title, and body aligned; omit abandoned approaches unless they explain a consequential trade-off.

For a draft based on supplied excerpts, label the evidence limit rather than claiming the entire branch was inspected. Do not invent issue links, testing evidence, or release effects.

## Publish when requested

When remote creation is requested, use the existing provider/tooling, inspect the target base and branch diff, and publish only the authorized scope. Prefer structured body arguments or a body file for multiline content. Do not infer a request to merge, deploy, or rewrite shared history.

If local edits are necessary, follow [task workspace preparation](task-branches.md); if a local commit is requested as part of delivery, follow [local commits](local-commits.md). Do not load those procedures merely to write a description for an already-prepared branch.

Report the created PR link or completed draft and actual checks, with any publication blocker stated explicitly.

Creating a PR does not complete the task workspace lifecycle. When its merge is confirmed during this task or a later continuation, follow [post-merge cleanup](workspace-cleanup.md). A closed, unmerged PR or an absent remote branch does not establish a merge.
