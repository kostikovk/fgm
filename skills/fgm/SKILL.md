---
name: fgm
description: Use when a project uses fgm or .fgm.toml, when asked to manage Go or golangci-lint with fgm, or when an fgm-managed project has toolchain, PATH, lint configuration, or Go/linter compatibility failures. Covers non-interactive agent commands, import/install, project upgrades, Go and linter pinning, CI consistency checks, and diagnostics.
---

# Fast Go Manager (`fgm`)

## Contract

1. Run every `go`, `golangci-lint`, and `make` call as `fgm --chdir DIR exec -- <command>`; wrap compound commands as `fgm --chdir DIR exec -- sh -c '...'`. Agent shells do not keep profile or `eval "$(fgm env)"` changes.
2. Read versions with `fgm --chdir DIR current --json` (`--all` for every module); never parse the text output.
3. Never bypass fgm with `GOTOOLCHAIN`, `GOROOT`, `brew`, or `go install`. `fgm exec` already sets `GOTOOLCHAIN=local`; a "requires go >= X" error means update `go.mod` or the pin, not the environment.
4. A missing selected linter is a blocker, not a fallback: install it (`fgm --chdir DIR install`, or `exec --install`) instead of using another `golangci-lint`. Scripts that call `$(GOPATH)/bin/golangci-lint` or another absolute path bypass `fgm exec` (`fgm check` C7).
5. Keep the repository's `.fgm.toml`, lint config, and build entrypoints unless the task authorizes changes. Do not edit shell profiles or run `fgm doctor --fix` unless asked.
6. Before reporting, run `fgm --chdir DIR check` and `fgm --chdir DIR doctor`; report versions with their sources and every fgm command you ran.

## Details

- [Version selection](references/selection.md): Go policy (`go = "latest-patch"`), detected linter pins, sources.
- [Commands](references/commands.md): command table and recovery of missing toolchains.
- [Upgrades and lint configuration](references/upgrades.md): frozen-target upgrades, pinning, `lint init`/`lint doctor`.
