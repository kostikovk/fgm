# Version selection

`fgm --chdir DIR current --json` reports each version with `source`, `source_file`, and (for detected linters) `source_line`. Use it instead of re-deriving the rules below.

## Go

1. `.fgm.toml` `[toolchain] go = "1.26.8"` — explicit version (`source: config`). It must not be lower than the `go` directive.
2. Nearest `go.work`, then `go.mod`: the `toolchain` directive.
3. The `go` directive. With `go = "latest-patch"`, it is raised to the newest released patch of that minor, never below a `toolchain` directive (`source: latest-patch`, `base` holds the directive).
4. The global default (`fgm use go X --global`) outside Go projects.

The `go` directive is a minimum; prefer `fgm --chdir DIR pin go latest-patch` over editing `go.mod` when the goal is "build with the newest patch like CI". A `latest-patch` warning about installed versions means the patch data may be stale.

## golangci-lint

1. `.fgm.toml` `golangci_lint = "vX.Y.Z"` (`source: config`); `"auto"` skips this step.
2. A pin the repository already has (`source: detected`): `.tool-versions`, `mise.toml`, a `go.mod` `tool` directive, `golangci/golangci-lint-action` `with.version`, or `GOLANGCI_LINT_VERSION`/`golangci-lint@vX`/`install.sh … vX` in Makefiles. Disagreeing pins select nothing and produce a warning listing each `file:line`.
3. The newest catalog version compatible with the resolved Go (`source: local` when installed, else `remote`).

`golangci_lint_detect = false` disables step 2. When pins disagree, make every source name the same version rather than adding a third pin; `fgm check` reports the disagreement as `C3`.

## Execution environment

`fgm exec` and the shims run the selected `go` with `GOTOOLCHAIN=local` and without `GOROOT`. `--gotoolchain inherit` (or `FGM_EXEC_GOTOOLCHAIN=inherit`) keeps the caller's values; use it only when the user asks for Go's own toolchain switching.
