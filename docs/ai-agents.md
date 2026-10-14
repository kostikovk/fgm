# AI coding agents

[Home](../README.md) · [Documentation](README.md)

FGM ships a standalone [Agent Skill](https://agentskills.io/specification) at [`skills/fgm/SKILL.md`](../skills/fgm/SKILL.md) so AI coding agents can inspect, bootstrap, and run project toolchains deterministically in non-interactive shells without mutating host installations or shell profiles.

## Set up your project with one prompt

Open the repository you want to configure in an AI coding agent with terminal, file, and network access. Copy and paste this prompt as one message:

```text
Set up Fast Go Manager (FGM) for this repository. Complete the available
installation, project setup, and verification steps, then report their results.

1. Inspect repository instructions, Git status, go.work/go.mod (including
   toolchain directives), inherited .fgm.toml and lint configs, scripts,
   Makefile, and CI. Identify the target project root and relevant modules;
   use each module/workspace directory as DIR below. If the target or required
   Go version is unclear, report that blocker rather than guessing.
   Preserve existing pins, configs, customized skills, and unrelated changes.
   Before any repository file write, resolve the destination (including symlinks)
   and verify it lies inside the target project root. Read inherited configs
   outside that root without modifying them. Do not change global defaults,
   shell profiles, Go requirements, application code, or commit/push changes.
   FGM itself and its toolchain store may use their documented user-local paths.

2. Read https://github.com/kostikovk/fgm/blob/main/README.md and its installation
   guide. Install FGM if missing; if already installed, record `fgm version`
   and use command help to check the commands/flags required by this workflow.
   Update an unsupported FGM using its documented installation method, or
   report an installation blocker. Preserve host Go and linter installations.
   Use FGM's absolute path in every call if its directory is outside PATH.

3. Obtain skills/fgm from https://github.com/kostikovk/fgm at a fixed commit or
   tagged release that actually contains the skill. Check resource availability;
   do not assume every binary release includes the skill. Record its revision.
   Use the agent-client table at
   https://github.com/kostikovk/fgm/blob/main/docs/ai-agents.md and your client's
   documented discovery rules to choose a project-local skill directory.
   If the source or discovery rules are unavailable, report a skill blocker.
   Copy the whole skill directory without overwriting a customized skill.
   Verify SKILL.md exists and is readable, then read and follow it. If a skill
   already exists, inspect and reuse it if suitable; report its known source.
   Use the client's discovery/invocation mechanism to verify registration where
   available. File presence alone does not prove registration. If reload is
   needed, follow the readable file for this session and report reload pending.

4. Inspect selection and diagnostics with `fgm --chdir DIR current` and
   `fgm --chdir DIR doctor`. If selection succeeds, run `fgm import auto`, then
   `fgm --chdir DIR install`, and recheck selection and installed binaries.
   Preserve the selected Go version and existing linter pin. If no linter is
   selected, report a compatibility blocker: `pin golangci-lint auto` uses the
   same selection logic and cannot repair an empty compatibility set. Do not
   add or change pins as a workaround. If the selected linter is absent or
   incompatible, report the blocker and continue independent Go checks when
   possible; do not substitute a linter from inherited PATH.

5. Audit an existing .golangci.yml/.golangci.yaml with
   `fgm --chdir DIR lint doctor`, including configs in ancestor directories.
   If none exists and a compatible selected linter is installed, determine the
   actual config output directory. Only when it is inside the target project,
   run `fgm --chdir DIR lint init --preset standard`, then audit the new config.
   Do not use --force or migrate existing configs. Inspect diagnostic output;
   successful exit status can still accompany ERROR or WARN findings.

6. Through `fgm --chdir DIR exec -- ...`, run `go version` and, only after
   confirming the selected linter is installed, `golangci-lint version`.
   Compare actual versions with `fgm --chdir DIR current`. A mismatch is a
   toolchain blocker; report it without bypassing FGM or changing Go metadata.
   Run the repository's documented test, build, and lint checks through FGM,
   after inspecting targets for source/config/module rewrites. If no suitable
   targets exist, use `go test ./...` and `go build ./...` through FGM. Cover
   each relevant module. Once the linter version is verified, run a real
   `golangci-lint run ./...` through FGM (with the project's existing flags),
   unless an existing lint target already performed that check. Do not infer
   lint success from doctor output. Continue independent checks after failures.
   Recheck current, doctor, and repository changes, including untracked files.
   Missing global defaults or shell shims alone do not block project commands
   through FGM exec; leave those host settings unchanged.

7. Report FGM's version/path, the skill path/source revision and registration
   or reload state, selected and actually executed Go/linter versions, changed
   files, and exact commands with exit codes and relevant diagnostic findings.
   Give separate PASS/FAIL/BLOCKED/NOT RUN results for toolchain setup, skill
   availability/registration, and each project check. Passing means the stage's
   required evidence was obtained; state why any check was not run. Repository
   test/lint failures must not be hidden by a successful tool installation.
   Do not claim full setup completion while required stages remain incomplete.

Proceed with routine setup within the available permissions. When blocked,
finish independent steps and report the exact remaining action or decision.
Do not claim checks passed unless you executed and inspected them.
```

The prompt authorizes installing or updating FGM, installing the selected toolchains and project skill, and creating a missing lint config inside the project. It preserves existing pins and inherited configuration. Completion depends on permissions, network access, compatible toolchains, and skill discovery; the report records each stage separately.

## Commands for agent sessions

These commands assume Go selection succeeds and the selected linter can be installed. If lint selection is empty or installation fails, keep lint checks blocked and continue independent Go checks.

In non-interactive agent sessions (where `eval "$(fgm env)"` does not persist across separate tool calls), run workspace commands through `fgm --chdir <dir> exec --`:

```bash
# 1. Inspect resolved Go and golangci-lint versions and environment health
fgm --chdir . current
fgm --chdir . doctor

# 2. Import existing system toolchains and install any missing workspace binaries
fgm import auto
fgm --chdir . install

# 3. Verify actual tool versions against current before running checks
fgm --chdir . exec -- go version
# Run the following linter commands only after confirming its selected version is installed
fgm --chdir . exec -- golangci-lint version
fgm --chdir . exec -- go test ./...
fgm --chdir . exec -- golangci-lint run ./...

# 4. Audit existing lint config (inspect output lines; exits 0 even on ERROR findings)
fgm --chdir . lint doctor
```

## Install the skill manually

Copy the `skills/fgm` directory from a tagged release or commit into your project's agent skill directory (having the `fgm` source repository elsewhere on disk does not automatically register the skill in another project):

| Agent Client | Project Skill Directory |
|---|---|
| [Claude Code](https://code.claude.com/docs/en/skills) | `.claude/skills/fgm/` |
| [Codex](https://learn.chatgpt.com/docs/build-skills) | `.agents/skills/fgm/` |
| [Pi](https://raw.githubusercontent.com/badlogic/pi-mono/main/packages/coding-agent/docs/skills.md) | `.agents/skills/fgm/` (supported shared alias) |
| [Cursor](https://cursor.com/docs/skills) | `.cursor/skills/fgm/` or `.agents/skills/fgm/` |
| [Gemini CLI](https://geminicli.com/docs/cli/skills/) | `.gemini/skills/fgm/` or `.agents/skills/fgm/` |

For a new `.agents/skills/fgm/` directory, the current standalone skill consists of `SKILL.md`. Check that the source is available and that the destination does not already contain a skill before using this example. For a fixed revision, replace `main` in the URL with a verified commit or tag containing the skill:

```bash
mkdir -p .agents/skills/fgm
curl -fsSL https://raw.githubusercontent.com/kostikovk/fgm/main/skills/fgm/SKILL.md \
  -o .agents/skills/fgm/SKILL.md
```

> [!NOTE]
> The skill directory paths above reflect each client's documented skill discovery conventions (verified against client documentation on October 14, 2026). Runtime behavior and explicit invocation syntax depend on the installed client version.
