# Fast Go Manager (fnm)

Fast Go toolchain management for local machines and repositories.

FGM is a Go CLI that resolves the right Go version for the current directory, manages a global default for work outside repositories, imports existing local installs, and routes `go` through shims so project-specific and machine-wide selection can coexist. It also manages compatible `golangci-lint` binaries through an embedded compatibility catalog.

> [!IMPORTANT]
> `golangci-lint` compatibility output is based on FGM's generated catalog in [`internal/golangcilint/compatibility.json`](./internal/golangcilint/compatibility.json), not a complete upstream historical matrix.

## Technology Stack

- Go `1.26.1`
- Cobra for CLI command structure
- Viper for environment-backed configuration
- `go-toml/v2` for `.fgm.toml` parsing
- Standard library `net/http` for remote release discovery and downloads

Primary module metadata lives in [`go.mod`](./go.mod).

## Project Architecture

FGM is a single-binary CLI with a thin Cobra command layer over focused internal packages:

- `cmd/` defines the user-facing CLI commands and flags.
- `internal/app` wires shared services into the command layer.
- `internal/resolve` and `internal/currenttoolchain` resolve the active Go and `golangci-lint` selection.
- `internal/goreleases`, `internal/goinstall`, `internal/goupgrade`, and `internal/goimport` handle Go release discovery, install, upgrade, and import flows.
- `internal/golangcilint`, `internal/lintinstall`, `internal/lintimport`, and `internal/lintlocal` provide the equivalent lint-toolchain workflow.
- `internal/envsetup`, `internal/execenv`, `internal/doctor`, and `internal/shim` cover shell integration, command execution, setup diagnostics, and shim behavior.

High-level flow:

```text
CLI command
  -> resolve current context from go.work / go.mod / global state
  -> locate or install toolchain in the FGM store
  -> execute, report, or update state
```

Inside a repository, FGM resolves Go in this order:

1. nearest `go.work`
2. nearest `go.mod`
3. `toolchain` directive before `go`
4. global default when no project metadata is found

## Getting Started

### Prerequisites

- Go `1.26.1` or newer to build from source
- A shell such as `zsh`, `bash`, `fish`, or PowerShell if you want shim support

### Install

```bash
git clone https://github.com/koskosovu4/fgm.git
cd fgm
go build -o ~/.local/bin/fgm .
export PATH="$HOME/.local/bin:$PATH"
```

### Quick Start

Import existing Go installations and inspect what FGM knows about:

```bash
fgm import auto
fgm versions go --local
```

Set a global default for directories that do not contain Go project metadata:

```bash
fgm use go 1.26.1 --global
```

Preview and apply shell setup:

```bash
fgm env
eval "$(fgm env)"
```

Verify the environment:

```bash
fgm doctor
which go
go version
```

Use the resolved toolchain in a repository:

```bash
cd /path/to/repo
fgm current
fgm exec -- go test ./...
```

## Project Structure

```text
fgm/
  cmd/                            Cobra command entrypoints
  internal/app/                   service wiring
  internal/currenttoolchain/      current Go + lint selection
  internal/doctor/                health checks
  internal/envsetup/              shell integration rendering
  internal/execenv/               command execution with resolved PATH
  internal/fgmconfig/             .fgm.toml parsing
  internal/goimport/              existing Go installation import
  internal/goinstall/             Go download and install
  internal/golangcilint/          lint compatibility catalog and provider
  internal/golocal/               local Go store management
  internal/goreleases/            remote Go release provider
  internal/goupgrade/             global and project Go upgrades
  internal/lintimport/            existing golangci-lint import
  internal/lintinstall/           golangci-lint download and install
  internal/lintlocal/             local lint store management
  internal/resolve/               project/global resolution rules
  internal/shim/                  shim resolution logic
  scripts/update-lint-compatibility/
  main.go
  Makefile
```

## Key Features

- Resolve Go versions from native Go metadata instead of a custom version file.
- Prefer `toolchain` over `go` when both directives exist.
- Fall back to a global Go version outside repositories.
- List installed and remote Go versions for the current platform.
- Install Go versions into an FGM-managed local store.
- Import existing Go and `golangci-lint` installations from common locations.
- Show the active Go and compatible `golangci-lint` selection with `fgm current`.
- Pin repo-level `golangci-lint` behavior in `.fgm.toml`.
- Run commands with the resolved toolchain via `fgm exec`.
- Generate shell setup snippets with `fgm env`.
- Validate installation and PATH setup with `fgm doctor`.
- Upgrade global or project Go versions with `fgm upgrade go`.
- Optionally install matching `golangci-lint` versions during Go upgrades.

## Common Commands

```text
fgm current
fgm doctor
fgm env [--shell zsh|bash|fish|powershell]
fgm exec -- <command> [args...]
fgm import auto
fgm install
fgm install go <version>
fgm install golangci-lint <version>
fgm pin golangci-lint <version|auto>
fgm remove go <version>
fgm remove golangci-lint <version>
fgm upgrade go --global
fgm upgrade go --project
fgm upgrade go --global --dry-run
fgm upgrade go --project --to <version>
fgm upgrade go --project --with-lint
fgm use go <version> --global
fgm versions go --local
fgm versions go --remote
fgm versions golangci-lint --local
fgm versions golangci-lint --remote [--go <version>]
```

## Development Workflow

Local development is centered on standard Go commands plus a small Makefile wrapper:

```bash
make build
make test
make fmt
make fix
make tidy
make pre-commit
```

Useful direct workflows:

```bash
go run . --help
go run . current --chdir .
make cmd CMD='current --chdir .'
```

The repository also includes Hooky-managed hooks in [`.hooky/hooks`](./.hooky/hooks). Sync them into `.git/hooks` with:

```bash
make hook-install
```

## Coding Standards

The codebase follows standard Go conventions reinforced by the current build workflow:

- Keep packages focused and responsibility-driven.
- Use Cobra commands as a thin CLI layer over internal services.
- Prefer explicit error returns over hidden side effects.
- Format code with `go fmt`.
- Apply automatic source fixes with `go fix` where appropriate.
- Keep command and service behavior covered by table-driven unit tests.

## Testing

Tests live alongside production code across both `cmd/` and `internal/` packages. The current project testing approach is:

- command-level tests for CLI behavior
- package-level unit tests for resolver, installer, importer, upgrade, and diagnostic services
- standard Go test execution through `go test ./...`

Run the full suite with:

```bash
go test ./...
```

Or use:

```bash
make test
```

## Contributing

Contributions should preserve the existing shape of the project:

1. Add or adjust focused tests next to the affected package.
2. Keep command wiring in `cmd/` and business logic in `internal/`.
3. Run `make pre-commit` before opening a change.
4. Regenerate the lint compatibility catalog with `make update-lint-compat` if you modify the compatibility generator or catalog inputs.

If you are changing shell behavior, install the local hooks and verify `fgm env`, `fgm doctor`, and shim flows manually.

## Storage Layout

FGM stores its managed toolchains under:

- `$XDG_DATA_HOME/fgm` when `XDG_DATA_HOME` is set
- otherwise `~/.local/share/fgm`

Typical layout:

```text
fgm/
  go/
    1.25.7/
      bin/go
    1.26.1/
      bin/go
  shims/
    go
  state/
    global-go-version
```

Imported Go installs are typically registered into this store via symlinks to existing local installations.

## License

No license file is currently present in this repository. Add a project license before distributing the tool outside private or internal use.
