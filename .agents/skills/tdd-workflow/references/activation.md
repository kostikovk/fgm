# Persistent invocation

Skill discovery and automatic selection differ between tools. A description or a loaded skill is not a technical interception of every file write. Place a concise explicit rule in the project's or user's standing-instruction file supported by the chosen tool. Make the skill files accessible and preserve existing instructions and runtime markers.

## Portable standing rule

Use this rule with the actual skill locations in the target environment:

```markdown
## Test-first development

Before writing or editing executable code, or verifying an executable-code issue, read and apply tdd-workflow/SKILL.md. Before writing or editing tests, read aaa-testing/SKILL.md and the relevant language reference. Resolve these files through the tool's skill catalog or their explicit accessible paths. Apply the same rule in delegated implementation tasks.

For new or changed behavior, require observed behavioral RED before implementation, minimal GREEN, then REFACTOR with green regression tests. Pure refactors start from verified green coverage; write missing characterization tests before production edits. Test-only and non-behavioral changes use their corresponding branches. Reproduce and verify executable issues only through executed behavioral tests; logs, manual probes, and reading guide investigation but do not establish reproduction. Reuse valid existing regression coverage instead of duplicating it. Do not claim RED, GREEN, or reproduced issues when test execution is blocked.
```

Use ordinary file reads when the tool has no skill command, catalog, or automatic activation. Invocation spelling such as `$tdd-workflow` is a tool-specific convenience; it is not required by the procedure. If a required file is inaccessible, report the missing guidance before executable edits. This repository does not install instructions, adapters, hooks, or permissions in another tool.

## Optional Codex setup example

For a Codex installation using `~/.codex` as its configuration directory, its standing instruction file can reference `~/.codex/skills/tdd-workflow/SKILL.md` and `~/.codex/skills/aaa-testing/SKILL.md` when those files are installed there. Use the configured locations and instruction-discovery behavior of the actual installation. This example is not a global path convention for Claude, Gemini, pi, or another tool.

## Verify activation separately from format

In an isolated workspace, use a plain implementation request that does not mention TDD. Inspect the agent's skill reads and ordered actions: test creation or valid regression reuse, a meaningful failing run, implementation, passing run, then a refactoring assessment. A correctly named green test alone does not prove invocation or test-first ordering. Check a separate issue-investigation scenario for an executed reproducing test before a reproduction claim. Run these checks per model and tool; a result in one integration does not establish activation in another.
