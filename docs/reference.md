# CLI and configuration reference

[Home](../README.md) · [Documentation](README.md)

## Commands

Use `fgm --help` or `fgm <command> --help` for flags and examples.

| Command | Description |
|---------|-------------|
| `fgm check [--all] [--json] [--fail-on error\|warn]` | Offline repository consistency gate for CI (Go/lint versions across `go.mod`, `.fgm.toml`, workflows, Dockerfiles, lint config) |
| `fgm current [--json] [--all]` | Show resolved Go and lint versions with source labels (`--json`: machine-readable; `--all`: every module) |
| `fgm doctor [--fix]` | Check installation health and environment (`--fix` adds shell integration to `zsh`/`bash`/`fish` profile) |
| `fgm env` | Print shell setup (`--shell zsh\|bash\|fish\|powershell`) |
| `fgm exec [--install] [--gotoolchain local\|inherit] [--allow-unselected-lint] -- <cmd> [args...]` | Run command with resolved Go (and installed selected lint) on `PATH` and `GOTOOLCHAIN=local` (`--install` installs missing resolved toolchains first) |
| `fgm import auto` | Auto-import existing Go and `golangci-lint` installations from system locations |
| `fgm install` | Install resolved Go (and selected `golangci-lint`, if resolved) for the workspace |
| `fgm install go <version>` | Install a specific Go version |
| `fgm install golangci-lint <version>` | Install a specific `golangci-lint` version |
| `fgm use go <version> --global` | Set global Go version and generate Unix shims |
| `fgm versions go [--local\|--remote]` | List Go versions (defaults to `--local`) |
| `fgm versions golangci-lint [--local\|--remote] [--go <ver>]` | List lint versions (defaults to `--local`; `--go` filters only with `--remote`) |
| `fgm prune [--dry-run]` | Remove unused Go and `golangci-lint` versions not referenced by global default or current workspace |
| `fgm remove go <version>` | Remove an FGM-managed Go version |
| `fgm remove golangci-lint <version>` | Remove an FGM-managed lint version |
| `fgm pin go <version\|latest-patch>` | Pin the Go version policy in `.fgm.toml` |
| `fgm pin golangci-lint <version\|auto>` | Pin lint version in `.fgm.toml` |
| `fgm lint init` | Generate a `.golangci.yml` for the resolved Go version |
| `fgm lint doctor` | Audit an existing `.golangci.yml` or `.golangci.yaml` for issues |
| `fgm upgrade go (--global\|--project)` | Upgrade Go (`--dry-run`, `--to <ver>`, `--with-lint`) |
| `fgm upgrade golangci-lint` | Upgrade `golangci-lint` (`--dry-run`, `--to <ver>`) |
| `fgm version` | Show build info |

> [!TIP]
> `fgm current` shows where each version was resolved from, e.g. `go 1.25.0 (go.mod)` or `golangci-lint v2.11.2 (config)`.

## Machine-readable output

`fgm current --json` prints one object (never styled, regardless of TTY). Paths in `source_file` are relative to `dir`:

```json
{
  "schema": 1,
  "dir": "/abs/path/services/api",
  "module_root": "/abs/path/services/api",
  "go": {"version": "1.26.9", "source": "latest-patch", "source_file": "go.mod",
         "base": "1.26.0", "installed": true, "path": "/…/go/1.26.9/bin/go"},
  "golangci_lint": {"version": "v2.9.0", "source": "detected", "source_file": "Makefile",
         "source_line": 31, "installed": false, "compatible": true},
  "warnings": []
}
```

| Field | Meaning |
|---|---|
| `schema` | Document version. Changes within a schema version are additive only. |
| `dir`, `module_root` | Absolute working directory and the directory of its nearest `go.mod` (`""` when none). |
| `go.source` | `go.work`, `go.mod`, `config` (`.fgm.toml` `go`), `latest-patch` (with `base`), or `global`. |
| `go.installed`, `go.path` | Whether the selected Go is in the FGM store (or is the imported/system Go), and its binary. |
| `golangci_lint` | `null` when no linter is selected. `source` is `config`, `detected` (with `source_line`), `local`, or `remote`. |
| `golangci_lint.compatible` | Catalog verdict for the selected Go; `null` when the catalog does not know the version. |
| `warnings` | Stale `latest-patch` data, conflicting detected lint pins, and similar notes. |
| `error` | Present only in `--all` entries whose module failed to resolve; `go` and `golangci_lint` are then `null`. |

`fgm current --all [--json]` reports every module under the working directory: the `use` entries of the nearest `go.work` that lie under it, otherwise every `go.mod` below it, skipping `vendor/`, `testdata/`, `node_modules/` and dot-directories and never following symlinks out of the root. Modules are listed in lexical order; JSON output is an array, text output is one line per module.

Exit status: `0` when every module resolved, `1` when at least one failed (its entry carries `error`), `2` for an invalid invocation (unknown flag or argument; this applies to every command).

## Repository consistency check

`fgm check` reads repository files only — no network, installs, or writes — so it is safe in CI and pre-commit hooks, and gives identical results offline. `fgm doctor` checks the *host*; `fgm check` checks the *repository*.

| ID | Check | Severity |
|---|---|---|
| C1 | Selected/detected `golangci-lint` cannot run on the resolved Go (catalog `max_go_minor`) | ERROR |
| C2 | Lint config schema does not match the linter major (`version: "2"` with v1, or missing with v2) | ERROR |
| C3 | Linter pin sources disagree (`.fgm.toml`, [detected pins](#detected-golangci-lint-pins)) | ERROR |
| C4 | `actions/setup-go` `go-version` minor lower than the module's `go` directive | ERROR |
| C4b | …different minor from the resolved Go (`go-version-file` is always OK; `${{ … }}` expressions are reported as INFO "not checked") | WARN |
| C5 | Dockerfile `FROM golang:<tag>` older than the `go` directive (a floating `1.26` tag is compared by minor; `ARG`-substituted tags are INFO "not checked") | ERROR |
| C6 | `go.work` `go` version lower than a used module's `go` directive | ERROR |

Output is one line per finding — `ID SEVERITY file:line message (fix: hint)` — or `OK no consistency findings`. `--json` prints the [`current --json`](#machine-readable-output) envelope plus a `findings` array (`id`, `severity`, `file`, `line`, `message`, `hint`; `file` is relative to `dir`); with `--all` it is an array with one entry per module. Workflows map to modules with the same `working-directory` rule as lint pin detection.

Exit status: `0` clean, `1` findings at or above `--fail-on` (default `error`), `2` could not run (invalid invocation or a module that cannot be resolved).

Skip checks for repositories whose files intentionally differ:

```toml
[check]
skip = ["C5"]   # skipping C4 also skips C4b
```

## Terminal output and script compatibility

When writing to an interactive TTY (with `NO_COLOR` and `CI` unset and `TERM != dumb`), `fgm` renders styled help (via [Fang](https://github.com/charmbracelet/fang)) and aligned, status-tagged output (via [Lip Gloss](https://github.com/charmbracelet/lipgloss)):

```text
$ fgm current
go             1.27.2  (go.mod)
golangci-lint  v2.13.0  (local)

$ fgm versions go --local
* 1.27.2  (current)
  1.26.9

$ fgm doctor
✓ OK  global Go version is installed: 1.27.2
✓ OK  shim dir is on PATH: ~/.local/share/fgm/shims
✓ OK  fgm is available on PATH
✓ OK  selected Go version is installed: 1.27.2
✓ OK  selected golangci-lint version is installed: v2.13.0
```

When `stdout` or `stderr` is piped/redirected, or when `NO_COLOR` is non-empty, `CI` is non-empty, or `TERM=dumb`, `fgm` automatically disables ANSI sequences and preserves the exact machine-readable plain-text format (`go 1.27.2 (go.mod)`, `* 1.27.2`, `OK ...` / `WARN ...`). `fgm env`, `fgm version`, `fgm exec`, and `fgm __shim` never add UI styling to their payload streams.

## Execution and imports

### Execution PATH

The resolved Go version must be installed locally. `fgm exec` searches `PATH` in this order:

1. Selected Go installation.
2. Selected `golangci-lint` installation, when installed in the FGM store.
3. Inherited `PATH`.

It does not auto-install missing tools unless `--install` is given. When the selected linter is not installed:

- `fgm exec -- golangci-lint ...` exits non-zero with the `fgm install golangci-lint <version>` hint instead of running another version from the inherited `PATH` (`--install` installs it first);
- any other command (for example `make lint`) still runs, and fgm prints one stderr warning naming the `golangci-lint` binary that `PATH` resolves to. Absolute paths inside scripts cannot be intercepted.

`--allow-unselected-lint` restores the previous behaviour: no error and no warning.

### Go toolchain environment

`fgm exec` and the `go`/`golangci-lint` shims run the child with `GOTOOLCHAIN=local` and without `GOROOT`. The `go` binary on `PATH` already is the selected toolchain, so this keeps "what `fgm current` reports" equal to "what runs" (the official `golang` container images set `GOTOOLCHAIN=local` for the same reason):

- an inherited `GOTOOLCHAIN=go1.X.Y` would make the selected `go` re-exec a different toolchain;
- an inherited `GOROOT` would make it use another installation's standard library and compiler;
- with `GOTOOLCHAIN=auto`, a module or dependency that declares a newer `go`/`toolchain` would silently switch toolchains.

A module that needs a newer Go now fails loudly (`requires go >= 1.26.8 (running go 1.26.0; GOTOOLCHAIN=local)`). Update `go.mod` or the pin instead of the environment.

To keep Go's own switching, pass `fgm exec --gotoolchain inherit` or set `FGM_EXEC_GOTOOLCHAIN=inherit` (the flag wins; shims read only the variable). With `inherit`, fgm prints one warning to stderr when the inherited `GOTOOLCHAIN` is not `local`, `auto`, or empty. `fgm doctor` warns when the shell sets `GOTOOLCHAIN` to a specific toolchain or sets `GOROOT`.

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

It prompts for confirmation when `stdin` is an interactive terminal and applies the change automatically when `stdin` is non-terminal. `fgm env --shell powershell` renders a PowerShell `PATH` snippet, but `shims/go` and `shims/golangci-lint` generation is not implemented on Windows.

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

1. `.fgm.toml` `go = "<version>"` — that version (`config`). It must not be lower than the `go` directive.
2. Nearest `go.work`, then `go.mod`: the `toolchain` directive.
3. The `go` directive — or, with `go = "latest-patch"`, the newest released patch of that minor (never lower than a `toolchain` directive), shown as `go 1.26.9 (latest-patch of go.mod 1.26.0)`.
4. Global default when no project metadata is found.

The `go` directive is a *minimum*: `go 1.26.0` written by `go mod init` selects the first 1.26 release with its known standard-library vulnerabilities, while CI usually runs the newest patch. `latest-patch` matches CI without editing `go.mod` for every consumer. Patch data comes from the embedded catalog, then `go.dev` releases (not used by the shims, which stay offline), then the newest installed patch of that minor, in which case `fgm current` and `fgm doctor` warn that the selection may be stale.

### Working directory

All workspace-aware commands accept `--chdir <dir>` (or the `FGM_CHDIR` environment variable when `--chdir` is omitted; an explicit `--chdir` flag takes precedence over `FGM_CHDIR`). FGM normalizes the working directory to an absolute path before walking ancestor directories for `go.work`, `go.mod`, `.fgm.toml`, and `.golangci.yml`/`.golangci.yaml`, and uses that directory as the child process working directory in `fgm exec`.

## Repository configuration

FGM uses `.fgm.toml` for repository-level settings. Create or update it with:

```bash
fgm pin go latest-patch   # or an explicit version, e.g. 1.26.8
fgm pin golangci-lint <version>
# or let FGM pick a compatible version automatically
fgm pin golangci-lint auto
```

`fgm pin` changes only the key it sets and writes to the nearest existing `.fgm.toml` walking upward from the working directory; if none exists yet, it writes `.fgm.toml` in the nearest ancestor directory containing `go.work`, then `go.mod`, or the working directory itself. Pinning a specific version records that string directly without checking the compatibility catalog upfront (use `fgm doctor` to verify compatibility against the resolved Go version).

Example `.fgm.toml`:

```toml
[toolchain]
go = "latest-patch"        # or "1.26.8"; omit to follow go.work/go.mod
golangci_lint = "v2.11.2"  # or "auto"; omit to use a detected pin
golangci_lint_detect = true  # false ignores pins in Makefile, workflows, .tool-versions, …
```

### Detected `golangci-lint` pins

Without an explicit `golangci_lint` in `.fgm.toml`, FGM looks for the version the repository already pins for CI before falling back to `auto`, and reports where it came from, e.g. `golangci-lint v2.5.0 (detected: Makefile:31)`. Precedence: `.fgm.toml` explicit version → detected pin → `auto`.

Detection is read-only and offline. It searches from the module root up to the repository root (the nearest `.git`, else the `go.work` root):

| Source | Recognised shape |
|---|---|
| `.tool-versions` (asdf, nearest) | `golangci-lint 2.5.0` |
| `mise.toml` / `.mise.toml` (nearest) | `[tools] golangci-lint = "2.5.0"` |
| module `go.mod` (Go 1.24+ tool directive) | `tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint` plus its `require` version |
| `.github/workflows/*.yml` | `uses: golangci/golangci-lint-action@…` with `with.version: v2.5.0` |
| nearest `GNUmakefile`/`makefile`/`Makefile` and `*.mk` | `GOLANGCI_LINT_VERSION ?= v2.5.0` (also `:=`, `::=`, `=`); `golangci-lint@v2.5.0`; `install.sh … v2.5.0` on a line mentioning `golangci-lint` |

- Only exact `X.Y.Z` versions count; `latest` or `v2.5` are not pins. Make variables are matched literally, without evaluation.
- When detected pins disagree, nothing is selected from detection: `fgm current` falls back to `auto` and prints the candidates on stderr, and `fgm doctor` reports a `WARN` with each `file:line`.
- A workflow pin applies to a module when the step's `with.working-directory` (or the job's `defaults.run.working-directory`) equals the module path relative to the repository root, or when the repository has a single module. Reusable workflows called from another repository are not followed.
- Detected versions go through the same compatibility check as explicit pins: `fgm doctor` warns when the catalog says the linter cannot run on the resolved Go.
- Opt out with `golangci_lint_detect = false` under `[toolchain]`.

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
    golangci-lint
  state/
    global-go-version
    update-check.json
```
