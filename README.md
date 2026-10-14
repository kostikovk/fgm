# Fast Go Manager (fgm)

> Set up and manage Go development environments with best-practice tooling — from a single binary.

FGM is a CLI tool that manages Go toolchains, golangci-lint compatibility, and lint configuration generation. It handles smart Go version resolution using native metadata (`go.mod`, `go.work`, `toolchain` directives), shell shims for transparent routing, an embedded golangci-lint compatibility catalog, and project-aware lint config generation.

## Features

- **Smart version resolution** from `go.work`, `go.mod`, and `toolchain` directives — no custom version files
- **Global default** for directories outside Go projects
- **Shell environment & Unix shims** — `fgm env` renders setup snippets for `bash`, `zsh`, `fish`, and `powershell`, while `go` shim script generation (`shims/go`) is supported on macOS and Linux
- **golangci-lint compatibility** — combines an embedded catalog with remote GitHub release lookups to pair Go versions with compatible lint versions
- **Lint config generation** via `fgm lint init` with preset and import-ordering support
- **Lint config diagnostics** via `fgm lint doctor`
- **Import existing installs** from common system locations via symlinks
- **Upgrade workflows** with `--dry-run` preview, `--to` target selection, and optional `--with-lint` matching
- **Health diagnostics** via `fgm doctor` (including `--fix` for shell profile integration on `zsh`, `bash`, and `fish`)

## Install

**Homebrew** (macOS/Linux):

```bash
brew install kostikovk/fgm/fgm
```

**Install script:**

```bash
curl -fsSL https://raw.githubusercontent.com/kostikovk/fgm/main/install.sh | sh
```

**From source:**

```bash
git clone https://github.com/kostikovk/fgm.git
cd fgm
go build -o ~/.local/bin/fgm .
```

## Uninstall

**Homebrew:**

```bash
brew uninstall fgm
brew untap kostikovk/fgm
```

**Remove managed toolchains and state:**

```bash
rm -rf "${XDG_DATA_HOME:-$HOME/.local/share}/fgm"
```

**Remove shell integration** — delete the `eval "$(fgm env)"` line from your shell profile (`.zshrc`, `.bashrc`, etc.).

## Quick Start

```bash
# Import existing Go and golangci-lint installations
fgm import auto

# Set a global default (and generate Unix shims)
fgm use go <version> --global

# Add shell integration (or run: fgm doctor --fix)
eval "$(fgm env)"

# Verify setup
fgm doctor

# Install the resolved Go and golangci-lint versions for the current repo
fgm install

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

1. Nearest `go.work` (`toolchain` directive takes precedence over `go`)
2. Nearest `go.mod` (`toolchain` directive takes precedence over `go`)
3. Global default when no project metadata is found

### Working Directory (`--chdir` and `FGM_CHDIR`)

All workspace-aware commands accept `--chdir <dir>` (or the `FGM_CHDIR` environment variable when `--chdir` is omitted; an explicit `--chdir` flag takes precedence over `FGM_CHDIR`). FGM normalizes the working directory to an absolute path before walking ancestor directories for `go.work`, `go.mod`, `.fgm.toml`, and `.golangci.yml`/`.golangci.yaml`, and uses that directory as the child process working directory in `fgm exec`.

## Commands

| Command | Description |
|---------|-------------|
| `fgm current` | Show resolved Go and lint versions with source labels |
| `fgm doctor [--fix]` | Check installation health and environment (`--fix` adds shell integration to `zsh`/`bash`/`fish` profile) |
| `fgm env` | Print shell setup (`--shell zsh\|bash\|fish\|powershell`) |
| `fgm exec -- <cmd> [args...]` | Run command with resolved Go (and installed selected lint) on `PATH` |
| `fgm import auto` | Auto-import existing Go and `golangci-lint` installations from system locations |
| `fgm install` | Install resolved Go (and selected `golangci-lint`, if resolved) for the workspace |
| `fgm install go <version>` | Install a specific Go version |
| `fgm install golangci-lint <version>` | Install a specific `golangci-lint` version |
| `fgm use go <version> --global` | Set global Go version and generate Unix shims |
| `fgm versions go [--local\|--remote]` | List Go versions (defaults to `--local`) |
| `fgm versions golangci-lint [--local\|--remote] [--go <ver>]` | List lint versions (defaults to `--local`; `--go` filters only with `--remote`) |
| `fgm remove go <version>` | Remove an FGM-managed Go version |
| `fgm remove golangci-lint <version>` | Remove an FGM-managed lint version |
| `fgm pin golangci-lint <version\|auto>` | Pin lint version in `.fgm.toml` |
| `fgm lint init` | Generate a `.golangci.yml` for the resolved Go version |
| `fgm lint doctor` | Audit an existing `.golangci.yml` or `.golangci.yaml` for issues |
| `fgm upgrade go (--global\|--project)` | Upgrade Go (`--dry-run`, `--to <ver>`, `--with-lint`) |
| `fgm upgrade golangci-lint` | Upgrade `golangci-lint` (`--dry-run`, `--to <ver>`) |
| `fgm version` | Show build info |

> [!TIP]
> `fgm current` shows where each version was resolved from, e.g. `go 1.25.0 (go.mod)` or `golangci-lint v2.11.2 (config)`.

### Execution, Import, and Upgrade Notes

- **`fgm exec` `PATH` contract:** The resolved Go version must be installed locally. If the workspace also selects a `golangci-lint` version and that binary is installed in the FGM store, its directory is prepended after Go (`selected Go` → `installed selected golangci-lint` → `inherited PATH`). If the selected `golangci-lint` version is not installed locally, `fgm exec` does not auto-install it and does not block unrelated commands; running `fgm exec -- golangci-lint ...` in that state will fall back to another `golangci-lint` on the inherited `PATH` or fail with a command lookup error if none is found.
- **Importing existing installations (`fgm import auto`):** Imported Go installations are registered as directory symlinks (`go/<version>` → external `GOROOT`) and imported `golangci-lint` installations as binary symlinks (`golangci-lint/<version>/golangci-lint` → external binary). Imported versions remain valid only while the external source installation exists on disk; `fgm remove` removes only the symlink inside the FGM store.
- **Shell profile auto-setup (`fgm doctor --fix`):** When the FGM shim directory is missing from `PATH`, `fgm doctor --fix` appends the shell integration snippet to `~/.zshrc` (`zsh`), `~/.bash_profile` on macOS or `~/.bashrc` on Linux (`bash`), or `~/.config/fish/config.fish` (`fish`). It prompts for confirmation when `stdin` is an interactive terminal and applies the change automatically when `stdin` is non-terminal. Note that `fgm env --shell powershell` renders a PowerShell `PATH` snippet, but shim generation (`shims/go`) is not implemented on Windows.
- **Upgrading Go and `golangci-lint`:** `fgm upgrade go` requires exactly one scope flag (`--global` or `--project`) and supports `--to <version>`, `--dry-run`, and `--with-lint` (`--with-lint` belongs only to `fgm upgrade go`). `fgm upgrade golangci-lint` supports `--dry-run` and `--to <version>`. Neither upgrade command rewrites `.fgm.toml` pins, and an explicit `--to <version>` on `fgm upgrade golangci-lint` is installed directly without catalog compatibility filtering.

## Configuration

FGM uses `.fgm.toml` for repository-level settings. Create or update it with:

```bash
fgm pin golangci-lint <version>
# or let FGM pick a compatible version automatically
fgm pin golangci-lint auto
```

`fgm pin` writes to the nearest existing `.fgm.toml` walking upward from the working directory; if none exists yet, it writes `.fgm.toml` in the nearest ancestor directory containing `go.work`, then `go.mod`, or the working directory itself. Pinning a specific version records that string directly without checking the compatibility catalog upfront (use `fgm doctor` to verify compatibility against the resolved Go version).

Example `.fgm.toml`:

```toml
[toolchain]
golangci_lint = "v2.11.2"
```

> [!IMPORTANT]
> Remote `golangci-lint` compatibility resolution combines FGM's generated catalog in [`internal/golangcilint/compatibility.json`](./internal/golangcilint/compatibility.json) with live GitHub release lookups, rather than a complete offline historical matrix.

## Lint Configuration

Use `fgm lint init` to generate a `golangci-lint` v2 config at the project root:

```bash
fgm lint init
fgm lint init --preset strict
fgm lint init --with-imports
fgm lint init --force
```

Available presets are `minimal`, `standard`, and `strict`. `fgm lint init` writes `.golangci.yml` in the nearest ancestor directory containing `go.work` or `go.mod` (or the working directory when neither exists) and checks only for an existing `.golangci.yml` unless `--force` is set.

Use `fgm lint doctor` to inspect an existing YAML config:

```bash
fgm lint doctor
```

`fgm lint doctor` walks upward from the working directory looking for `.golangci.yml` or `.golangci.yaml` (preferring `.golangci.yml` when both exist in the same directory). It reports missing `version: "2"` metadata (`ERROR`), unknown linters (`WARN`), linters incompatible with the project's Go version (`WARN`), formatter/linter conflicts (`WARN`), and `INFO` recommendations for linters from the `standard` preset that are not yet enabled.

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
