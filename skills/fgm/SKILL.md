---
name: fgm
description: Use when a project uses fgm or .fgm.toml, when asked to manage Go or golangci-lint with fgm, or when an fgm-managed project has toolchain, PATH, lint configuration, or Go/linter compatibility failures. Covers non-interactive agent commands, import/install, project upgrades, linter pinning, and diagnostics.
---

# Fast Go Manager (`fgm`)

Manage Go and `golangci-lint` toolchains, execute project commands with the resolved toolchain in non-interactive agent shells, and maintain repository lint configuration without mutating host-wide installations or shell profiles.

## 1. Scope and boundaries

- Use `fgm` to inspect, import, install, pin, upgrade, and execute the workspace's Go and `golangci-lint` toolchains, and to generate or diagnose `.golangci.yml` v2 configurations.
- Preserve the repository's existing build/test entrypoints (`Makefile`, scripts, flags, build tags) and existing `.fgm.toml` or `.golangci.yml`/`.golangci.yaml` policies unless the task explicitly authorizes changing them.
- Keep toolchain management separate from Go coding, TDD, Git delivery, and permission rules owned by other skills.

## 2. Discover the project and selected toolchain

Before running Go, `golangci-lint`, or `make` commands:

1. Identify the target module or workspace directory `DIR` (`go.work`, `go.mod`, and any `.fgm.toml` walking upward from `DIR`).
2. Check that `fgm` is available (`fgm version`) and inspect the resolved workspace selection and health:
   ```bash
   fgm --chdir DIR current
   fgm --chdir DIR doctor
   ```
3. Check locally available versions when needed (`fgm --chdir DIR versions go --local`, `fgm --chdir DIR versions golangci-lint --local`).
4. Note how versions are selected:
   - Go resolves from the nearest `go.work` (`toolchain` over `go`), then nearest `go.mod` (`toolchain` over `go`), then the global default (`state/global-go-version`).
   - `.fgm.toml` pins `golangci-lint` (`[toolchain] golangci_lint = "vX.Y.Z"` or `"auto"`); it never pins Go.

## 3. Run commands in non-interactive sessions

AI agent tool calls execute in separate non-interactive subshells where shell profile edits and prior `eval "$(fgm env)"` calls do not persist into subsequent commands.

- Run every independent `go`, `golangci-lint`, or `make` invocation through `fgm --chdir DIR exec -- <command> [args...]`.
- For shell pipelines or compound commands (`&&`, `||`, `;`, `|`), either wrap each toolchain command with `fgm --chdir DIR exec -- ...` or invoke an explicit child shell under `fgm exec`:
  ```bash
  fgm --chdir DIR exec -- sh -c 'go vet ./... && golangci-lint run ./...'
  ```
- Do not modify `~/.zshrc`, `~/.bashrc`, `~/.bash_profile`, or `~/.config/fish/config.fish` (or run `fgm doctor --fix`) just to run project commands in an agent session unless the user explicitly asked for persistent shell profile setup.
- Do not bypass `fgm` by installing global toolchains with `brew`, `go install`, or ad-hoc `GOTOOLCHAIN` overrides.
- `fgm exec` prepends the resolved Go binary directory (and the selected `golangci-lint` binary directory only when that selected version is installed in the FGM store) ahead of the inherited `PATH`. Absolute binary paths bypass `PATH` routing, and if the selected `golangci-lint` version is not installed in the FGM store, running `fgm exec -- golangci-lint ...` can fall back to an incompatible `golangci-lint` on the inherited `PATH` or fail command lookup—install the selected linter first.

## 4. Choose the command

In the table below, `DIR` is the target module or workspace directory, and `X` (Go version), `V` (`golangci-lint` version), and `TARGET` (target Go version) are concrete versions established from workspace metadata or user instructions.

| Intent / state | Command or procedure |
|---|---|
| Inspect resolved Go and lint versions | `fgm --chdir DIR current` |
| Check environment and toolchain health | `fgm --chdir DIR doctor` |
| Import existing system tools during setup | `fgm import auto`, then `fgm --chdir DIR install` |
| Install the workspace's selected Go and lint set | `fgm --chdir DIR install` |
| Install a specific missing Go version | `fgm install go X` |
| Install a specific missing linter version | `fgm install golangci-lint V` |
| Run tests, build, lint, or Make targets | `fgm --chdir DIR exec -- go test ./...` or `fgm --chdir DIR exec -- make <target>` |
| List remote compatible linters for a target Go version | `fgm versions golangci-lint --remote --go TARGET` |
| Change an authorized repository linter pin | `fgm --chdir DIR pin golangci-lint V` (or `auto`), then `fgm --chdir DIR install` and `fgm --chdir DIR doctor` |
| Generate a new v2 lint config | `fgm --chdir DIR lint init --preset standard`, then `fgm --chdir DIR lint doctor` and a real lint run |
| Audit an existing lint config | `fgm --chdir DIR lint doctor` and inspect every output line |
| Configure an explicit global default Go version | Install/import `X` first, then `fgm use go X --global` (Unix shims only; on Windows use `fgm exec`) |

## 5. Recover missing environment state

- **Workspace selection is known (`fgm --chdir DIR current` succeeds), but binaries are missing:**
  - Run `fgm import auto` first when existing system/Homebrew installations may satisfy the selection via symlinks, then run `fgm --chdir DIR install` (or the exact `fgm install go X` / `fgm install golangci-lint V` command from the CLI hint).
  - `fgm install` populates the local FGM store (`go/X`, `golangci-lint/V`) but does **not** set the global default Go version.
- **No `go.work`/`go.mod` metadata and no global Go selection:**
  - Verify whether `DIR` points to the right module directory (`--chdir DIR`).
  - Only set a global default (`fgm use go X --global`) when the target version `X` is known from project context or user instruction; do not invent an arbitrary `latest` global version.

## 6. Upgrade and lint configuration safely

- **Pinning vs installation:** `fgm --chdir DIR pin golangci-lint V` writes `.fgm.toml` (walking up to the nearest existing `.fgm.toml`, `go.work`, or `go.mod`) without checking compatibility or installing `V`. Always follow a pin change with `fgm --chdir DIR install` and `fgm --chdir DIR doctor`.
- **Upgrading Go or `golangci-lint` with a frozen target:**
  1. Preview first with `--dry-run` (`fgm --chdir DIR upgrade go --project --with-lint --dry-run` or `fgm --chdir DIR upgrade golangci-lint --dry-run`).
  2. Freeze the previewed version with `--to TARGET`, verify compatible linters (`fgm versions golangci-lint --remote --go TARGET`) and any explicit `.fgm.toml` pin, and re-run `--dry-run` with `--to TARGET`.
  3. Execute the live upgrade with the same `--to TARGET` flag. If the pin or target changes, repeat the preview and compatibility check before running without `--dry-run`.
  4. Note that `fgm upgrade go --with-lint` preserves explicit `.fgm.toml` pins (even if incompatible with the new Go version) and is not atomic: if Go installation or metadata rewrite succeeds and the subsequent linter install fails, report the partial state (`go.mod`/`go.work` updated, linter pending) clearly.
- **Generating and auditing lint configuration:**
  - Before running `fgm --chdir DIR lint init`, check for existing `.golangci.yml` or `.golangci.yaml` in `DIR` and ancestor directories (`fgm --chdir DIR lint doctor`). `fgm lint init` checks only for `.golangci.yml` at the project root and overwrites rather than migrating custom or v1 configs when `--force` is passed.
  - `fgm --chdir DIR lint doctor` prints `ERROR`, `WARN`, `INFO`, or `OK` lines and exits with status `0` even when `ERROR` findings are present—always inspect its output text rather than relying on the exit code.

## 7. Verify and report

- Confirm the final state with `fgm --chdir DIR current` and `fgm --chdir DIR doctor`.
- When lint configuration or toolchain versions change, run an actual lint execution (`fgm --chdir DIR exec -- golangci-lint run ./...`) in addition to `fgm --chdir DIR lint doctor`. `fgm doctor` and `fgm lint doctor` validate against FGM's embedded catalog and static rules; only a real `golangci-lint run` proves runtime analyzer compatibility.
- Report the resolved Go and `golangci-lint` versions (with their sources), the exact `fgm` commands executed, and any remaining environment or compatibility gaps.
