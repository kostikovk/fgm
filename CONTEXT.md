# Project Domain & Architecture Context (`CONTEXT.md`)

Fast Go Manager (`fgm`) is a single-binary Go CLI that manages Go toolchains, `golangci-lint` compatibility, and project-aware lint configuration. It resolves Go versions per directory from native Go metadata (`go.work`, `go.mod`, `toolchain` directives), routes `go` invocations through shell shims, and generates or audits `.golangci.yml` configurations.

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
| `.agents/skills/*/SKILL.md` | Task-specific engineering workflows, boundaries, and reference routing |
| `.agents/skills/*/references/*.md` | Focused technical procedures and language/framework guidance |
| [README.md](README.md) | User-facing CLI documentation, installation, and quick-start examples |
| This file | Project purpose, domain terms, layered architecture, package ownership, and invariants |

Treat the working tree, `go.mod`, and `Makefile` as the source of truth for current versions, commands, and file availability.

## Domain glossary

| Term | Definition | Code representation |
|---|---|---|
| **Selection (Resolved Toolchain)** | The active Go and `golangci-lint` versions resolved for a working directory, paired with their resolution sources (`GoSource`: resolved `go.work`/`go.mod` file path or `"global"`; `LintSource`: `"config"`, `"local"`, or `"remote"` — `"auto"` in `.fgm.toml` defers to `"local"` or `"remote"`). | `app.Selection` in `internal/app/app.go`, produced by `internal/resolve` and `internal/currenttoolchain` |
| **Toolchain Store** | The FGM-managed directory under `$XDG_DATA_HOME/fgm` (default `~/.local/share/fgm`) holding installed Go and `golangci-lint` binaries, shims, and global state. | `internal/golocal` and `internal/lintlocal` |
| **Shim** | Lightweight executable wrapper in the FGM store (`shims/go`) that intercepts `go` calls and dispatches them to the resolved Go binary for the current working directory. | `cmd/shim.go`, `internal/shim`, and `internal/golocal` |
| **Compatibility Catalog** | Embedded JSON mapping Go minor versions to compatible and recommended `golangci-lint` versions (`internal/golangcilint/compatibility.json`). | `internal/golangcilint` and generator `scripts/update-lint-compatibility` |
| **Repo Pin Config (`.fgm.toml`)** | Repository-level TOML file pinning `toolchain.golangci_lint` to a specific version or `"auto"`. | `internal/fgmconfig` |
| **Lint Config (`.golangci.yml`)** | Generated or existing `golangci-lint` v2 YAML configuration managed through presets (`minimal`, `standard`, `strict`) and audited for version/linter compatibility. | `internal/lintconfig` (`catalog.go`, `presets.go`, `generate.go`, `imports.go`, `doctor.go`) |
| **Shell Integration** | Shell-specific environment snippet (`eval "$(fgm env)"`) that places FGM shims on `PATH`. | `internal/envsetup` (`renderer.go`) |

## Architecture and package ownership

FGM follows a strict three-layer design:

1. **`cmd/` (CLI Facade Layer)** — Thin Cobra subcommand definitions (`current`, `doctor`, `env`, `exec`, `import`, `install`, `lint`, `pin`, `remove`, hidden `__shim` in `cmd/shim.go`, `upgrade`, `use`, `version`, `versions`). Commands parse flags, bind Viper configuration where needed, delegate to services on `*app.App` (except `cmd/pin.go`, which calls `fgmconfig.SaveNearest` directly, and `cmd/shim.go`, which constructs `shim.New` and invokes `os/exec` directly), and format output. Business logic does not live in `cmd/`.
2. **`internal/app/app.go` (Dependency Injection & Service Contracts)** — Defines the central `App` struct, domain value types (`Selection`, `LintVersion`, `LintFinding`, `GoUpgradeOptions`, `LintConfigOptions`, etc.), and service interfaces consumed by `cmd/`. `main.go` wires concrete implementations into `app.App` and passes it to `cmd.NewRootCmd()`.
3. **`internal/*` (Focused Service & Utility Packages)** — Single-responsibility packages implementing the interfaces in `internal/app/app.go`:

| Package | Responsibility |
|---|---|
| `internal/resolve` | Walks directory ancestors to locate `go.work` or `go.mod`, parses `toolchain` and `go` directives, and falls back to the global Go version |
| `internal/currenttoolchain` | Combines resolved Go version with `.fgm.toml` pinning and `golangci-lint` compatibility into a unified `app.Selection` |
| `internal/golocal` / `internal/lintlocal` | Manage installed Go and `golangci-lint` versions in the FGM store, global Go default state, and shim creation |
| `internal/goreleases` | Fetches remotely available Go releases from the official `go.dev` API |
| `internal/golangcilint` | Serves the embedded compatibility catalog (`compatibility.json`) and queries remote `golangci-lint` releases |
| `internal/goinstall` / `internal/lintinstall` | Download, verify, extract, and register Go and `golangci-lint` archives into the local store |
| `internal/goimport` / `internal/lintimport` | Detect and import existing Go and `golangci-lint` installations from standard system paths |
| `internal/goupgrade` / `internal/lintupgrade` | Upgrade Go (global or project `go.mod`/`go.work`) and `golangci-lint` with `--dry-run` and `--with-lint` support |
| `internal/fgmconfig` | Parse and persist repository-level `.fgm.toml` configuration |
| `internal/lintconfig` | Embedded linter catalog (`linters.json`), presets (`minimal`, `standard`, `strict`), `.golangci.yml` generation, import ordering, and `lint doctor` diagnostics |
| `internal/envsetup` | Render shell setup scripts (`bash`, `zsh`, `fish`, `powershell`) |
| `internal/shim` | Resolve the target Go binary path for shim execution |
| `internal/execenv` | Execute external commands with the resolved toolchain prepended to `PATH` |
| `internal/doctor` | Run environment health checks and report actionable findings |
| `internal/archive` | Archive extraction (`.tar.gz`, `.zip`) and download progress reporting |
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
  state/global-go-version
```

### Build metadata injection

`main.go` declares `buildVersion`, `buildCommit`, and `buildDate`, injected at build time via `-ldflags -X` in `Makefile` and `.goreleaser.yaml`. `release_files_test.go` guards the integrity of `.goreleaser.yaml`, `.github/workflows/release.yml`, `install.sh`, and `Makefile`.

## Architectural invariants

1. **Thin command layer:** `cmd/` contains only CLI flag handling, input wiring, and presentation. Domain rules and filesystem/network orchestration belong in `internal/`.
2. **Interface seam in `internal/app/app.go`:** Service dependencies injected into `cmd/` are declared as interfaces on `app.App` so commands and services remain testable in isolation with lightweight stubs (current direct-call exceptions are `cmd/pin.go` → `fgmconfig.SaveNearest` and `cmd/shim.go` → `shim.New` / `os/exec`).
3. **Generated compatibility data:** `internal/golangcilint/compatibility.json` is generated by `scripts/update-lint-compatibility` (`make update-lint-compat`) rather than edited by hand.
4. **Deterministic, isolated tests:** Unit and command tests use temporary directories (`t.TempDir()`), injected callbacks/interfaces, and in-memory or local fixtures; they must not mutate the user's real `$XDG_DATA_HOME/fgm` store or shell profile.
