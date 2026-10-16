# Upgrades and lint configuration

## Pinning

`fgm --chdir DIR pin go …` and `fgm --chdir DIR pin golangci-lint V` write only their own key in the nearest `.fgm.toml` (or create it next to the nearest `go.work`/`go.mod`). Pinning neither installs nor checks compatibility: follow every pin change with `fgm --chdir DIR install`, `fgm --chdir DIR doctor`, and `fgm --chdir DIR check`.

## Upgrading with a frozen target

1. Preview: `fgm --chdir DIR upgrade go --project --with-lint --dry-run` or `fgm --chdir DIR upgrade golangci-lint --dry-run`.
2. Freeze the previewed version with `--to TARGET`, check compatible linters (`fgm versions golangci-lint --remote --go TARGET`) and any explicit `.fgm.toml` pin, and preview again with `--to TARGET`.
3. Run the live upgrade with the same `--to TARGET`. If the pin or target changes, repeat the preview and compatibility check first.
4. `fgm upgrade go --with-lint` keeps explicit `.fgm.toml` pins (even incompatible ones) and is not atomic: if Go and `go.mod`/`go.work` were updated but the linter install failed, report that partial state.

After an upgrade, `fgm --chdir DIR check` catches CI files left behind: `setup-go` versions (`C4`/`C4b`), Dockerfile `golang` tags (`C5`), and `go.work` versions (`C6`).

## Lint configuration

- Before `fgm --chdir DIR lint init`, look for an existing `.golangci.yml` or `.golangci.yaml` in `DIR` and its ancestors (`fgm --chdir DIR lint doctor`). `lint init` checks only `.golangci.yml` at the project root, and `--force` overwrites instead of migrating custom or v1 configs.
- A v2 linter needs `version: "2"` in the config and a v1 linter cannot read it; `fgm check` reports a mismatch as `C2`.
- `fgm doctor`, `fgm check`, and `fgm lint doctor` use FGM's embedded catalog and static rules. Only a real run proves analyzer compatibility: `fgm --chdir DIR exec -- golangci-lint run ./...`.
