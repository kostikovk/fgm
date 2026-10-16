# Commands

`DIR` is the target module or workspace directory; `X` (Go), `V` (`golangci-lint`), and `TARGET` (target Go) are concrete versions from workspace metadata or user instructions.

| Intent / state | Command or procedure |
|---|---|
| Inspect resolved versions (machine-readable) | `fgm --chdir DIR current --json` (`--all` for every module) |
| Check repository consistency (CI-safe, offline) | `fgm --chdir DIR check` (`--all`, `--json`, `--fail-on warn`) |
| Check host and toolchain health | `fgm --chdir DIR doctor` |
| Import existing system tools during setup | `fgm import auto`, then `fgm --chdir DIR install` |
| Install the workspace's selected Go and linter | `fgm --chdir DIR install` |
| Install a specific missing version | `fgm install go X` / `fgm install golangci-lint V` |
| Run tests, build, lint, or Make targets | `fgm --chdir DIR exec -- go test ./...` / `fgm --chdir DIR exec -- make <target>` |
| Run and install missing toolchains first | `fgm --chdir DIR exec --install -- go test ./...` |
| Remove unused Go and linter versions | `fgm --chdir DIR prune` (`--dry-run` to preview) |
| List remote linters compatible with a Go version | `fgm versions golangci-lint --remote --go TARGET` |
| Change an authorized Go policy | `fgm --chdir DIR pin go latest-patch` (or `X`), then `fgm --chdir DIR install` |
| Change an authorized linter pin | `fgm --chdir DIR pin golangci-lint V` (or `auto`), then `install`, `doctor`, and `check` |
| Set a global default Go | Install/import `X`, then `fgm use go X --global` (Unix shims only; on Windows use `fgm exec`) |

`fgm current --all` and `fgm check` exit `1` when a module fails or a finding reaches `--fail-on`, and `2` for an invalid invocation or an unresolvable module. `fgm lint doctor` exits `0` even with `ERROR` lines: read its output.

## Recover missing environment state

- **Selection is known but binaries are missing:** run `fgm import auto` when system or Homebrew installations may satisfy it, then `fgm --chdir DIR install` (or the exact `fgm install …` command from the error hint). `fgm install` fills the FGM store but does not set the global default.
- **`exec -- golangci-lint` fails with "selected golangci-lint … is not installed":** install it as above. Use `--allow-unselected-lint` only when the user explicitly accepts a different linter version.
- **No `go.work`/`go.mod` and no global Go:** check that `DIR` is the module directory. Set a global default only when the version is known from project context or the user; never invent one.
- **`fgm exec` warns that `golangci-lint` on `PATH` resolves to another binary:** a script calls `golangci-lint` through `PATH` while the selected version is missing; install the selected version.
