# Project Domain & Architecture Context (`CONTEXT.md`)

Fast Go Manager (`fgm`) is a single-binary Go CLI that manages Go toolchains, `golangci-lint` compatibility, and project-aware lint configuration. It resolves Go versions per directory from native Go metadata (`go.work`, `go.mod`, `toolchain` directives), routes `go` and `golangci-lint` invocations through shell shims, and generates or audits `.golangci.yml` configurations.

## Contents

1. [Document responsibilities](#document-responsibilities)
2. [Domain glossary](#domain-glossary)
3. [Architecture and package ownership](#architecture-and-package-ownership)
4. [Version resolution and storage](#version-resolution-and-storage)
5. [Architectural invariants](#architectural-invariants)

## Document responsibilities

| Document | Responsibility |
|---|---|
| [AGENTS.md](AGENTS.md) | Repository operating instructions, task-to-skill navigation, verified CLI commands, and engineering boundaries |
| [`skills/fgm/SKILL.md`](skills/fgm/SKILL.md) | Distributable standalone `fgm` skill for AI coding agents (mirrored locally in `.agents/skills/fgm/SKILL.md`) |
| `.agents/skills/*/SKILL.md` | Task-specific engineering workflows, boundaries, and reference routing |
| `.agents/skills/*/references/*.md` | Focused technical procedures and language/framework guidance |
| [README.md](README.md) | User-facing overview, installation entrypoints, quick-start example, and documentation navigation |
| [Documentation](docs/README.md) | Installation and project setup guides, AI agent setup prompt, CLI/configuration reference, and development instructions |
| `site/` | Static landing page (`Next.js 16`, `TypeScript 7`, `output: "export"`, `basePath: "/fgm"`) and Playwright E2E test suite (`site/e2e/landing.spec.ts`) deployed via `.github/workflows/pages.yml` |
| This file | Project purpose, domain terms, layered architecture, package ownership, and invariants |

Treat the working tree, `go.mod`, and `Makefile` as the source of truth for current versions, commands, and file availability.

## Domain glossary

| Term | Definition | Code representation |
|---|---|---|
| **Selection (Resolved Toolchain)** | The active Go and `golangci-lint` versions resolved for a working directory, paired with their resolution sources (`GoSource`: resolved `go.work`/`go.mod` file path or `"global"`; `LintSource`: `"config"`, `"local"`, or `"remote"` — `"auto"` in `.fgm.toml` defers to `"local"` or `"remote"`). | `app.Selection` in `internal/app/app.go`, produced by `internal/resolve` and `internal/currenttoolchain` |
| **Toolchain Store** | The FGM-managed directory under `$XDG_DATA_HOME/fgm` (default `~/.local/share/fgm`) holding installed Go and `golangci-lint` binaries, shims, and global state. | `internal/golocal` and `internal/lintlocal` |
| **Shim** | Lightweight Unix executable wrapper in the FGM store (`shims/go`, `shims/golangci-lint`) that intercepts `go` and `golangci-lint` calls and dispatches them to the resolved binaries for the current working directory (`EnsureShims` is supported on macOS/Linux and rejects Windows). | `cmd/shim.go`, `internal/shim`, and `internal/golocal` |
| **Compatibility Catalog** | Embedded JSON mapping Go minor versions to compatible and recommended `golangci-lint` versions (`internal/golangcilint/compatibility.json`), combined with remote GitHub release queries in `internal/golangcilint`. | `internal/golangcilint` and generator `scripts/update-lint-compatibility` |
| **Repo Pin Config (`.fgm.toml`)** | Repository-level TOML file pinning `toolchain.golangci_lint` to a specific version or `"auto"`. `SaveNearest` updates the nearest existing `.fgm.toml` upward, or creates one in the nearest ancestor `go.work` root, then `go.mod` root, then the working directory. Explicit pins are stored without upfront compatibility catalog filtering. | `internal/fgmconfig` |
| **Lint Config (`.golangci.yml`)** | Generated or existing `golangci-lint` v2 YAML configuration managed through presets (`minimal`, `standard`, `strict`) and audited for version/linter compatibility. | `internal/lintconfig` (`catalog.go`, `presets.go`, `generate.go`, `imports.go`, `doctor.go`) |
| **Shell Integration** | Shell-specific environment snippet (`eval "$(fgm env)"` or `fgm env \| source`) for `bash`, `zsh`, `fish`, and `powershell`, plus profile auto-registration (`zsh`, `bash`, `fish`) via `app.ProfileInstaller` (interactive confirmation on terminal `stdin`, automatic application on non-terminal `stdin`). | `internal/envsetup` (`renderer.go`, `profile.go`) and `cmd/doctor.go` |

## Architecture and package ownership

FGM follows a strict three-layer design:

1. **`cmd/` (CLI Facade Layer)** — Thin Cobra subcommand definitions (`current`, `doctor`, `env`, `exec`, `import`, `install`, `lint`, `pin`, `prune`, `remove`, hidden `__shim` in `cmd/shim.go`, `upgrade`, `use`, `version`, `versions`) plus terminal presentation helpers in `cmd/ui.go` (`charm.land/fang/v2` for help/error rendering and `charm.land/lipgloss/v2` for TTY-only styled command output, preserving plain-text output whenever piped/redirected or when `NO_COLOR`, `CI`, or `TERM=dumb` is set). Commands parse flags, normalize the effective working directory (`--chdir` / `FGM_CHDIR`, with explicit `--chdir` taking precedence) to an absolute path via `resolveWorkDir` in `cmd/root.go`, delegate to services on `*app.App` (except `cmd/pin.go`, which calls `fgmconfig.SaveNearest` directly, and `cmd/shim.go`, which constructs `shim.New` and invokes `os/exec` directly), and format output (for example, `versions go` and `versions golangci-lint` default to `--local`, while `--go` filters `golangci-lint` only with `--remote`). Business logic does not live in `cmd/`.
2. **`internal/app/app.go` (Dependency Injection & Service Contracts)** — Defines the central `App` struct, domain value types (`Selection`, `LintVersion`, `LintFinding`, `GoUpgradeOptions`, `LintConfigOptions`, etc.), and service interfaces (`SelectionResolver`, `GoStore`, `LintStore`, `GoRemoteVersionProvider`, `LintRemoteVersionProvider`, `GoInstaller`, `LintInstaller`, `GoImporter`, `LintImporter`, `GoUpgrader`, `LintUpgrader`, `DoctorService`, `Executor`, `EnvRenderer`, `ProfileInstaller`, `LintConfigGenerator`, `LintDoctor`) consumed by `cmd/`. `main.go` wires concrete implementations into `app.App` and passes it to `cmd.NewRootCmd()`.
3. **`internal/*` (Focused Service & Utility Packages)** — Single-responsibility packages implementing the interfaces in `internal/app/app.go`:

| Package | Responsibility |
|---|---|
| `internal/resolve` | Walks directory ancestors to locate `go.work` or `go.mod`, parses `toolchain` and `go` directives, and falls back to the global Go version (wired directly as `App.GoResolver` for `cmd/shim.go` so `go` shims never resolve `golangci-lint`) |
| `internal/currenttoolchain` | Combines resolved Go version with `.fgm.toml` pinning and `golangci-lint` compatibility (using the embedded compatibility catalog via `LintCatalogProvider` for offline local selection, or `LintRemoteProvider` when configured) into a unified `app.Selection` |
| `internal/golocal` / `internal/lintlocal` | Manage installed Go and `golangci-lint` versions in the FGM store (including symlinked imported installations, which depend on the external source remaining on disk), global Go default state, and Unix shim creation |
| `internal/goreleases` | Fetches remotely available Go releases from the official `go.dev` API |
| `internal/golangcilint` | Serves the embedded compatibility catalog (`compatibility.json`) and queries remote `golangci-lint` releases on GitHub |
| `internal/goinstall` / `internal/lintinstall` | Download, verify, extract, and register Go and `golangci-lint` archives into the local store |
| `internal/goimport` / `internal/lintimport` | Detect and import existing Go (`GOROOT` directory symlinks) and `golangci-lint` (binary symlinks) installations from standard system paths |
| `internal/goupgrade` / `internal/lintupgrade` | Upgrade Go (requiring exactly one of `--global` or `--project`, with `--dry-run`, `--to`, and Go-only `--with-lint`) and `golangci-lint` (`--dry-run` and unvalidated `--to`, without rewriting `.fgm.toml` pins) |
| `internal/fgmconfig` | Parse and persist repository-level `.fgm.toml` configuration (`LoadNearest`, `ResolvePinnedLint`, `SaveNearest`) |
| `internal/lintconfig` | Embedded linter catalog (`linters.json`), presets (`minimal`, `standard`, `strict`), `.golangci.yml` generation at the nearest `go.work`/`go.mod` root (checking only `.golangci.yml` unless `--force`), import ordering, and upward `lint doctor` diagnostics across `.golangci.yml`/`.golangci.yaml` (`.yml` preferred over `.yaml` in the same directory; emits `ERROR`, `WARN`, and `standard`-preset `INFO` findings) |
| `internal/envsetup` | Render shell setup scripts (`bash`, `zsh`, `fish`, `powershell`) and install shell integration into `zsh`, `bash`, and `fish` profiles (`ProfileInstaller`) |
| `internal/shim` | Resolve the target Go and `golangci-lint` binary paths for shim execution |
| `internal/execenv` | Execute external commands in the resolved working directory with `PATH` ordered as `selected Go` (required) → `installed selected golangci-lint` (prepended only when installed locally) → `inherited PATH`, without auto-installing missing toolchains |
| `internal/doctor` | Run environment health checks and report actionable findings (including `fgm` release update checks and `shell_profile` fix hints handled by `fgm doctor --fix`) |
| `internal/fgmupdate` | Check and cache the latest `fgm` release on GitHub (`state/update-check.json`, 24h TTL) and provide non-blocking cached update notices for interactive TTY sessions (suppressed when `FGM_NO_UPDATE_NOTIFIER` is set) |
| `internal/toolhint` | Format exact CLI recovery and self-upgrade command hints (`fgm install go <version>`, `fgm install golangci-lint <version>`, `fgm exec -- <command>`, `brew update && brew upgrade --cask fgm`) shared across stores, commands, and doctor diagnostics |
| `internal/archive` | Archive extraction (`.tar.gz`, `.zip`), cross-filesystem atomic move fallback (`MoveAtomically` on `EXDEV`), and download progress reporting |
| `internal/versionutil` | Semantic/Go version parsing, comparison, normalization, and binary discovery helpers |
| `internal/testutil` | Shared command-execution test helpers for `cmd/` and integration tests |

## Version resolution and storage

### Go version resolution order

Inside a workspace, `internal/resolve` selects the Go version by priority:

1. Nearest `go.work` (`toolchain` directive preferred over `go` directive).
2. Nearest `go.mod` (`toolchain` directive preferred over `go` directive).
3. Global default (`state/global-go-version`) when no project metadata exists.

### Store layout

Managed toolchains and state live under `$XDG_DATA_HOME/fgm` (default `~/.local/share/fgm`):

```text
fgm/
  go/<version>/bin/go
  golangci-lint/<version>/golangci-lint
  shims/go
  shims/golangci-lint
  state/global-go-version
  state/update-check.json
```

### Build metadata injection

`main.go` declares `buildVersion`, `buildCommit`, and `buildDate`, injected at build time via `-ldflags -X` in `Makefile` and `.goreleaser.yaml`. `release_files_test.go` guards the integrity of `.goreleaser.yaml`, `.github/workflows/release.yml`, `install.sh`, and `Makefile`.

## Architectural invariants

1. **Thin command layer:** `cmd/` contains only CLI flag handling, input wiring, and presentation. Domain rules and filesystem/network orchestration belong in `internal/`.
2. **Interface seam in `internal/app/app.go`:** Service dependencies injected into `cmd/` are declared as interfaces on `app.App` so commands and services remain testable in isolation with lightweight stubs (current direct-call exceptions are `cmd/pin.go` → `fgmconfig.SaveNearest` and `cmd/shim.go` → `shim.New` / `os/exec`).
3. **Generated compatibility data:** `internal/golangcilint/compatibility.json` is generated by `scripts/update-lint-compatibility` (`make update-lint-compat`) rather than edited by hand.
4. **Deterministic, isolated tests:** Unit and command tests use temporary directories (`t.TempDir()`), injected callbacks/interfaces, and in-memory or local fixtures; they must not mutate the user's real `$XDG_DATA_HOME/fgm` store or shell profile.
