# Fast Go Manager (fgm)

> Set up and manage Go development environments with best-practice tooling — from a single binary.

FGM is a CLI tool that manages Go toolchains, golangci-lint compatibility, and lint configuration generation. In future releases it will expand further into broader Go project setup: scaffolding, import organization workflows, and CI templates. The goal is a single command that takes a Go project from zero to production-ready best practices.

**Current scope** focuses on smart Go version resolution using native metadata (`go.mod`, `go.work`, `toolchain` directives), shell shims for transparent routing, an embedded golangci-lint compatibility catalog, and project-aware lint config generation.

## UI Direction

FGM will adopt [`go-tui`](https://github.com/grindlemire/go-tui) as the main interactive UI layer for future releases. The existing Cobra CLI remains the automation and scripting surface, while the TUI becomes the primary guided experience for discovery, diagnostics, installs, upgrades, and project setup workflows.

The intended split is:

- **Cobra commands** remain the stable non-interactive interface for scripts, CI, and power users
- **Internal services** continue to own all business logic and state changes
- **go-tui** orchestrates those services into a richer terminal UX with menus, forms, progress views, and guided setup flows

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
fgm use go 1.26.1 --global

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
> `fgm current` shows where each version was resolved from, e.g. `go 1.26.1 (global)` or `golangci-lint v2.11.2 (config)`.

## Configuration

FGM uses `.fgm.toml` for repository-level settings. Create or update it with:

```bash
fgm pin golangci-lint v2.11.2
# or let FGM pick a compatible version automatically
fgm pin golangci-lint auto
```

Example `.fgm.toml`:

```toml
[toolchain]
golangci_lint = "v2.11.2"
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
    1.25.7/bin/go
    1.26.1/bin/go
  golangci-lint/
    v2.11.2/golangci-lint
  shims/
    go
  state/
    global-go-version
```

## Roadmap

FGM is growing from a version manager into a complete Go project setup tool. Here's what's planned next:

### UI Foundation — `go-tui` Adoption Plan

To make FGM easier to use interactively, the next UI milestone is adopting `go-tui` as the main terminal interface.

1. Add a dedicated TUI entry command such as `fgm ui` that launches the interactive app
2. Keep all core logic in `internal/` services so the TUI stays thin and testable
3. Start with read-heavy flows: current toolchain view, doctor results, installed versions, compatibility browsing
4. Expand to guided actions: install, use, import, upgrade, and lint config generation
5. Add progress and confirmation screens for long-running or destructive operations
6. Keep feature parity by routing TUI actions through the same service layer used by Cobra commands

### MVP-2 — Project Scaffolding

- `fgm init` — scaffold a new Go project with recommended structure, `Makefile`, `.golangci.yml`, and `.gitignore`
- Pre-commit hook setup (`fgm hooks init`)
- Editorconfig and formatting defaults

### MVP-3 — CI & Workflow Templates

- `fgm ci init` — generate GitHub Actions workflows for Go (test, lint, release)
- GoReleaser config generation
- Dependency update automation templates (Dependabot / Renovate)

### Future

- Static analysis profile management (`go vet`, `staticcheck`)
- Project-level tool pinning beyond lint (e.g., `buf`, `mockgen`, `sqlc`)
- Team-shared configuration via remote presets

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
