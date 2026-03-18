# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build          # Build binary to tmp/fgm (injects version/commit/date via ldflags)
make test           # Run all tests: go test ./...
make fmt            # Format: go fmt ./...
make fix            # Apply go fix + format
make tidy           # Clean up go.mod/go.sum
make pre-commit     # Run fix, test, tidy, build (also the pre-commit hook)
make cover-html     # Generate HTML coverage report at tmp/coverage.html
```

Run a single test:

```bash
go test ./internal/resolve/ -run TestResolveName
```

Run the CLI during development:

```bash
go run . current --chdir .
make cmd CMD='current --chdir .'
```

## Project Vision

FGM is a Go toolchain manager with golangci-lint compatibility and lint configuration support. The codebase covers version management, shell shims, an embedded lint compatibility catalog, and project-aware lint config generation/diagnostics.

## Architecture

FGM is a single-binary CLI that manages Go toolchain versions, golangci-lint compatibility, and generated lint configuration. It resolves the correct Go version per-directory using native Go metadata (`go.mod`, `go.work`, `toolchain` directives) and routes `go` through shell shims.

### Layered Design

**`cmd/`** — Thin Cobra command layer. Each file defines one subcommand, receives the `app.App` struct, and delegates to internal services. No business logic lives here.

**`internal/app/app.go`** — Central dependency injection. The `App` struct holds interfaces for every service (resolver, stores, installers, importers, etc.). All interfaces are defined in this file. `main.go` wires concrete implementations into `App` and passes it to `cmd.NewRootCmd()`.

**`internal/`** — Focused packages, each owning one responsibility:

| Package | Role |
|---------|------|
| `resolve` | Version resolution: walks up to find `go.work`/`go.mod`, reads `toolchain`/`go` directives, falls back to global |
| `currenttoolchain` | Combines resolver + lint compatibility into a single "current versions" view |
| `golocal` / `lintlocal` | Manage installed versions in the FGM store, track global default |
| `goreleases` | Fetch available Go versions from go.dev API |
| `golangcilint` | Embedded compatibility catalog (`compatibility.json`) + remote lint version provider |
| `goinstall` / `lintinstall` | Download and install toolchain versions |
| `goimport` / `lintimport` | Auto-import existing installations from common system locations |
| `goupgrade` | Global and project-level Go upgrade logic with dry-run support |
| `fgmconfig` | Parse/save `.fgm.toml` (repo-level lint pinning) |
| `lintconfig` | Generate and audit `.golangci.yml` based on resolved Go/lint toolchains |
| `envsetup` | Render shell setup snippets (bash/zsh/fish/powershell) |
| `shim` | Shim resolution — routes `go` to the correct toolchain |
| `execenv` | Execute commands with resolved toolchain on PATH |
| `doctor` | Health check diagnostics |

### Version Resolution Order

Inside a repository, Go version is resolved by priority:
1. Nearest `go.work`
2. Nearest `go.mod` — `toolchain` directive preferred over `go` directive
3. Global default (for directories with no project metadata)

### Storage

Toolchains are stored under `$XDG_DATA_HOME/fgm` (default `~/.local/share/fgm`) with subdirectories for `go/`, `golangci-lint/`, `shims/`, and `state/`.

### Build Injection

`main.go` declares `buildVersion`, `buildCommit`, `buildDate` variables populated via `-ldflags -X` at build time (see Makefile and `.goreleaser.yaml`).

## Conventions

- Commands in `cmd/` are thin facades — put logic in `internal/` packages.
- All service dependencies are defined as interfaces in `internal/app/app.go`.
- Config file is `.fgm.toml` (TOML), managed by `internal/fgmconfig/`.
- Generated lint config is `.golangci.yml` (YAML), managed by `internal/lintconfig/`.
- The lint compatibility catalog at `internal/golangcilint/compatibility.json` is generated — regenerate with `make update-lint-compat`.
