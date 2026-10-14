# CLI and configuration reference

[Home](../README.md) · [Documentation](README.md)

## Commands

Use `fgm --help` or `fgm <command> --help` for flags and examples.

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

## Execution and imports

### Execution PATH

The resolved Go version must be installed locally. `fgm exec` searches `PATH` in this order:

1. Selected Go installation.
2. Selected `golangci-lint` installation, when installed in the FGM store.
3. Inherited `PATH`.

It does not auto-install missing tools. If the selected linter is missing, unrelated commands still run, but `golangci-lint` may fall back to a different version on the inherited `PATH` or fail to start. Run `fgm install` before relying on the selected linter.

### Imported installations

`fgm import auto` registers existing tools through symlinks:

- Go: `go/<version>` points to the external `GOROOT` directory.
- Linter: `golangci-lint/<version>/golangci-lint` points to the external binary.

Imported versions depend on those source installations remaining on disk. `fgm remove` removes only the symlink inside the FGM store.

## Shell profile setup

When the shim directory is missing from `PATH`, `fgm doctor --fix` appends a shell integration snippet to the corresponding profile:

| Shell | Profile |
|---|---|
| Zsh | `~/.zshrc` |
| Bash on macOS | `~/.bash_profile` |
| Bash on Linux | `~/.bashrc` |
| Fish | `~/.config/fish/config.fish` |

It prompts for confirmation when `stdin` is an interactive terminal and applies the change automatically when `stdin` is non-terminal. `fgm env --shell powershell` renders a PowerShell `PATH` snippet, but `shims/go` generation is not implemented on Windows.

## Upgrade behavior

Preview an upgrade before applying it:

```bash
fgm upgrade go --project --with-lint --dry-run
fgm upgrade golangci-lint --dry-run
```

- Go upgrades require exactly one scope: `--global` or `--project`.
- Both commands support `--to <version>` and `--dry-run`.
- `--with-lint` is available only for Go upgrades.
- Neither command rewrites `.fgm.toml` pins.
- An explicit `--to` linter version is installed directly without compatibility filtering.

After choosing a target, preview again with `--to <version>` and use that same target for the live upgrade. Verify explicit linter pins against the target Go version. See the [agent skill](../skills/fgm/SKILL.md#6-upgrade-and-lint-configuration-safely) for handling partial upgrade results.

## Version resolution

Inside a repository, FGM resolves Go in this order:

1. Nearest `go.work` (`toolchain` directive takes precedence over `go`)
2. Nearest `go.mod` (`toolchain` directive takes precedence over `go`)
3. Global default when no project metadata is found

### Working directory

All workspace-aware commands accept `--chdir <dir>` (or the `FGM_CHDIR` environment variable when `--chdir` is omitted; an explicit `--chdir` flag takes precedence over `FGM_CHDIR`). FGM normalizes the working directory to an absolute path before walking ancestor directories for `go.work`, `go.mod`, `.fgm.toml`, and `.golangci.yml`/`.golangci.yaml`, and uses that directory as the child process working directory in `fgm exec`.

## Repository configuration

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
> Remote `golangci-lint` compatibility resolution combines FGM's generated catalog in [`internal/golangcilint/compatibility.json`](../internal/golangcilint/compatibility.json) with live GitHub release lookups, rather than a complete offline historical matrix.

## Lint configuration

Use `fgm lint init` to generate a `golangci-lint` v2 config at the project root:

```bash
fgm lint init
fgm lint init --preset strict
fgm lint init --with-imports
```

Available presets are `minimal`, `standard`, and `strict`. `fgm lint init` writes `.golangci.yml` in the nearest ancestor directory containing `go.work` or `go.mod` (or the working directory when neither exists) and checks only for an existing `.golangci.yml` unless `--force` is set.

Use `fgm lint doctor` to inspect an existing YAML config:

```bash
fgm lint doctor
```

`fgm lint doctor` walks upward from the working directory looking for `.golangci.yml` or `.golangci.yaml` (preferring `.golangci.yml` when both exist in the same directory). It reports missing `version: "2"` metadata (`ERROR`), unknown linters (`WARN`), linters incompatible with the project's Go version (`WARN`), formatter/linter conflicts (`WARN`), and `INFO` recommendations for linters from the `standard` preset that are not yet enabled.

`--force` overwrites rather than migrates existing `.golangci.yml` files. Inspect both YAML filenames, including ancestor directories, before generating a new config. `lint doctor` exits with status 0 when parsing succeeds, even if it prints `ERROR` findings; inspect its output and run real lint checks to verify analyzer compatibility.

## Storage layout

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
