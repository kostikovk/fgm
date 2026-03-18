# Fast Go Manager (fgm)

> Set up and manage Go development environments with best-practice tooling — from a single binary.

FGM is a CLI tool that manages Go toolchains, golangci-lint compatibility, and lint configuration generation. It handles smart Go version resolution using native metadata (`go.mod`, `go.work`, `toolchain` directives), shell shims for transparent routing, an embedded golangci-lint compatibility catalog, and project-aware lint config generation.

## Features

- **Smart version resolution** from `go.work`, `go.mod`, and `toolchain` directives — no custom version files
- **Global default** for directories outside Go projects
- **Shell shims** for transparent `go` routing across bash, zsh, fish, and PowerShell
- **golangci-lint compatibility** — embedded catalog maps Go versions to compatible lint versions
- **Lint config generation** via `fgm lint init` with preset and import-ordering support
- **Lint config diagnostics** via `fgm lint doctor`
- **Import existing installs** from common system locations
- **Upgrade workflows** with `--dry-run` preview and optional `--with-lint` matching
- **Health diagnostics** via `fgm doctor`

## Install

**Homebrew** (macOS/Linux):

```bash
brew install koskosovu4/fgm/fgm
```

**Install script:**

```bash
curl -fsSL https://raw.githubusercontent.com/koskosovu4/fgm/main/install.sh | sh
```

**From source:**

```bash
git clone https://github.com/koskosovu4/fgm.git
cd fgm
go build -o ~/.local/bin/fgm .
```

## Quick Start

```bash
# Import existing Go installations
fgm import auto

# Set a global default
fgm use go <version> --global

# Add shell integration (add to your shell profile)
eval "$(fgm env)"

# Verify setup
fgm doctor

# Generate a starter golangci-lint config for the current project
fgm lint init --preset standard
```

Inside a repository, FGM automatically resolves the correct Go version:

```bash
cd /path/to/repo
fgm current        # Show resolved Go and lint versions
fgm exec -- go test ./...
```

## Version Resolution

Inside a repository, FGM resolves Go in this order:

1. Nearest `go.work`
2. Nearest `go.mod` (`toolchain` directive takes precedence over `go`)
3. Global default when no project metadata is found

## Commands

| Command | Description |
|---------|-------------|
| `fgm current` | Show resolved Go and lint versions with source labels |
| `fgm doctor` | Check installation health and environment |
| `fgm env` | Print shell setup (`--shell zsh\|bash\|fish\|powershell`) |
| `fgm exec -- <cmd>` | Run command with resolved toolchain on PATH |
| `fgm import auto` | Auto-import existing Go/lint from system locations |
| `fgm install go <version>` | Install a Go version |
| `fgm install golangci-lint <version>` | Install a golangci-lint version |
| `fgm use go <version> --global` | Set global Go version |
| `fgm versions go [--local\|--remote]` | List available Go versions |
| `fgm versions golangci-lint [--local\|--remote]` | List lint versions (`--go <ver>` to filter) |
| `fgm remove go <version>` | Remove an installed Go version |
| `fgm remove golangci-lint <version>` | Remove an installed lint version |
| `fgm pin golangci-lint <version\|auto>` | Pin lint version in `.fgm.toml` |
| `fgm lint init` | Generate a `.golangci.yml` for the resolved Go version |
| `fgm lint doctor` | Audit an existing `.golangci.yml` for issues |
| `fgm upgrade go` | Upgrade Go (`--global\|--project`, `--dry-run`, `--with-lint`) |
| `fgm version` | Show build info |

> [!TIP]
> `fgm current` shows where each version was resolved from, e.g. `go 1.23.0 (global)` or `golangci-lint v2.1.0 (config)`.

## Configuration

FGM uses `.fgm.toml` for repository-level settings. Create or update it with:

```bash
fgm pin golangci-lint <version>
# or let FGM pick a compatible version automatically
fgm pin golangci-lint auto
```

Example `.fgm.toml`:

```toml
[toolchain]
golangci_lint = "v2.1.0"
```

> [!IMPORTANT]
> golangci-lint compatibility output is based on FGM's generated catalog in [`internal/golangcilint/compatibility.json`](./internal/golangcilint/compatibility.json), not a complete upstream historical matrix.

## Lint Configuration

Use `fgm lint init` to generate a `golangci-lint` v2 config at the project root:

```bash
fgm lint init
fgm lint init --preset strict
fgm lint init --with-imports
fgm lint init --force
```

Available presets are `minimal`, `standard`, and `strict`.

Use `fgm lint doctor` to inspect an existing YAML config:

```bash
fgm lint doctor
```

The doctor command reports missing v2 metadata, unknown linters, incompatible linters for the project's Go version, and known formatter/linter conflicts.

## Storage Layout

FGM stores managed toolchains under `$XDG_DATA_HOME/fgm` (default: `~/.local/share/fgm`):

```text
fgm/
  go/
    1.22.0/bin/go
    1.23.0/bin/go
  golangci-lint/
    v2.1.0/golangci-lint
  shims/
    go
  state/
    global-go-version
```

## Development

```bash
make build        # Build binary to tmp/fgm
make test         # Run all tests
make fmt          # Format Go files
make pre-commit   # Run fix, test, tidy, build
make cover-html   # Generate HTML coverage report
```

Run directly:

```bash
go run . --help
go run . current --chdir .
```
