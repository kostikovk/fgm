# FGM

Fast Go toolchain management for local machines and repositories.

FGM is a Go CLI for selecting the right Go version for the current directory, switching a global default, importing existing installations, and routing `go` through shims so repo-level and machine-level selection can coexist.

> [!IMPORTANT]
> FGM's Go workflow is implemented today. `golangci-lint` remote compatibility listing is available, while lint install and paired Go plus lint install flows are still planned.

## Why FGM

- Resolve Go from native Go metadata instead of a custom version file.
- Support both repo-level selection and a global machine default.
- Import existing Go installations instead of forcing users to reinstall everything.
- Keep startup and lookup paths simple: local files first, network only for remote version listing and installs.

## Current Features

- Resolve Go from `go.work` or the nearest `go.mod`.
- Prefer `toolchain` over `go` when both exist.
- Fall back to a global Go version outside repos.
- List local and remote Go versions.
- List compatible remote `golangci-lint` versions for a target Go version.
- Install Go versions into an FGM-managed store.
- Import existing Go installs from common locations.
- Select a global default Go version.
- Run commands with the resolved Go version via `fgm exec`.
- Generate shell environment setup with `fgm env`.
- Validate setup with `fgm doctor`.
- Route `go` through FGM shims.

## How Resolution Works

Inside a repo, FGM resolves Go in this order:

1. nearest `go.work`
2. nearest `go.mod`
3. inside either file, `toolchain` first
4. inside either file, `go` second

Outside a repo, FGM falls back to the global version selected with `fgm use go <version> --global`.

This means you can keep a machine-wide default like `1.25.7`, while a repo using `toolchain go1.26.0` will still run with `1.26.0`.

## Install

Build FGM from source:

```bash
git clone https://github.com/koskosovu4/fgm.git
cd fgm
go build -o ~/.local/bin/fgm .
```

Make sure the binary is on `PATH`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Quick Start

### 1. Import existing Go installs

```bash
fgm import auto
fgm versions go --local
```

`import auto` scans common locations such as your current `PATH`, `/usr/local/go`, and Homebrew Go installs, then registers them in the FGM store.

### 2. Pick a global default

```bash
fgm use go 1.26.1 --global
```

### 3. Enable shims in your shell

Preview the shell snippet:

```bash
fgm env
```

Apply it to the current shell:

```bash
eval "$(fgm env)"
```

Persist it in your shell config if you want FGM active in every session.

### 4. Verify setup

```bash
fgm doctor
which go
go version
```

### 5. Use repo-specific versions automatically

```bash
cd /path/to/repo
fgm current
fgm exec -- go test ./...
```

## Common Workflows

### Show the selected Go version

```bash
fgm current
fgm current --chdir /path/to/repo
```

When a compatible `golangci-lint` version is known, `fgm current` also prints the selected lint version.

### List installed Go versions

```bash
fgm versions go --local
```

### List remote Go versions for your platform

```bash
fgm versions go --remote
```

FGM marks the currently resolved version with `*`.

### List compatible remote golangci-lint versions

```bash
fgm versions golangci-lint --local
fgm versions golangci-lint --remote --go 1.25.0
fgm versions golangci-lint --remote --chdir /path/to/repo
```

`--local` lists installed `golangci-lint` versions from the FGM-managed store.

FGM filters the fetched remote releases to versions that match your platform and a curated embedded compatibility manifest, then marks the recommended version with `*`.

This output should be read as:

- known compatible versions in FGM's generated catalog
- not a guaranteed complete list of every historically compatible `golangci-lint` release ever published

The manifest is generated from upstream `golangci-lint` releases, Go support issues, and the Go releases feed. Regenerate it with:

```bash
make update-lint-compat
```

### Install a Go version

```bash
fgm install go 1.25.7
```

FGM downloads the correct archive for your OS and architecture, verifies it, shows download progress, and installs it into the local FGM store.

### Install a golangci-lint version

```bash
fgm install golangci-lint v2.11.2
```

FGM downloads the matching archive for your platform, verifies the archive checksum when available, and installs the binary into the local FGM store.

### Switch the global default

```bash
fgm use go 1.25.7 --global
```

### Run a command with the resolved toolchain

```bash
fgm exec -- go version
fgm exec -- golangci-lint version
fgm exec --chdir /path/to/repo -- go test ./...
```

When a compatible installed `golangci-lint` version is selected, `fgm exec` also prepends its binary directory to `PATH`.

### Remove an installed FGM-managed version

```bash
fgm remove go 1.25.7
fgm remove golangci-lint v2.11.2
```

## Shims

FGM shims are small wrapper scripts placed earlier on `PATH` than your system Go.

They let plain `go` mean:

- the repo-selected version when you are inside a workspace or module
- the global selected version when you are outside a repo

Without shims, FGM still works through `fgm exec`, but your normal `go` command will continue using whichever binary appears first on `PATH`.

> [!NOTE]
> The shim calls `fgm __shim ...`, so the `fgm` binary itself must also be available on `PATH`.

## Storage Layout

FGM stores data in:

- `$XDG_DATA_HOME/fgm` when `XDG_DATA_HOME` is set
- otherwise `~/.local/share/fgm`

Current layout:

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

Imported Go installs are registered into this store, typically via symlinks to existing local installations.

## Command Summary

```text
fgm current
fgm doctor
fgm env [--shell zsh|bash|fish|powershell]
fgm exec -- <command> [args...]
fgm import auto
fgm install go <version>
fgm install golangci-lint <version>
fgm remove go <version>
fgm remove golangci-lint <version>
fgm use go <version> --global
fgm versions go --local
fgm versions go --remote
fgm versions golangci-lint --local
fgm versions golangci-lint --remote [--go <version>]
```

FGM also exposes Cobra-generated shell completion via `fgm completion`.

## Development

Run tests:

```bash
go test ./...
make test
```

Run the CLI directly:

```bash
go run . --help
go run . current --chdir .
make cmd
make cmd CMD='current --chdir .'
```

Build a local binary:

```bash
go build -o ./tmp/fgm .
./tmp/fgm --help
make build
```

The codebase is being built test-first with table-driven unit tests and CLI-level command tests.

### Pre-commit hook

This repo includes a Hooky-managed pre-commit hook in [`.hooky/hooks/pre-commit`](/Users/koskosovu4/projects/fgm/.hooky/hooks/pre-commit).

Install Hooky:

```bash
go install github.com/kostikovk/hooky@latest
```

Sync the hook into `.git/hooks`:

```bash
make hook-install
```

The hook runs:

```bash
make pre-commit
```

That currently executes:

- `make fix`
- `make build`

## Status And Next Work

Implemented now:

- Go resolution from `go.work` and `go.mod`
- local and remote Go version listing
- remote compatible `golangci-lint` version listing
- embedded compatibility manifest for verified known `golangci-lint` support ranges
- Go install, remove, import, global use, doctor, env, exec, and shim support

Planned next:

- paired install flow for Go plus compatible `golangci-lint`
- `golangci-lint` install and local version management
- richer health checks and shell onboarding polish
