# Using FGM in CI

[Home](../README.md) · [Documentation](README.md)

## Pinned installation

`install.sh` installs the latest release unless `FGM_VERSION` names an exact tag. It always verifies the archive against the release `checksums.txt` and fails closed — nothing is extracted or installed on a mismatch. Fetch the script from the same tag you install, so the installer itself is pinned too:

```bash
curl -fsSL https://raw.githubusercontent.com/kostikovk/fgm/v0.4.0/install.sh | FGM_VERSION=v0.4.0 sh
```

| Variable | Default | Meaning |
|---|---|---|
| `FGM_VERSION` | latest release | Exact release tag to install (`v` prefix optional) |
| `INSTALL_DIR` | `~/.local/bin` | Where the `fgm` binary is placed |
| `FGM_RELEASE_BASE_URL` | `https://github.com/kostikovk/fgm/releases/download` | Mirror laid out as `<base>/<tag>/<file>` |

## `setup-fgm` action

The composite action in [`setup-fgm/`](../setup-fgm/action.yml) installs a pinned fgm, resolves the repository's toolchain with `fgm current --json`, installs it (restoring the FGM store from `actions/cache` when possible), puts the selected Go and `golangci-lint` directories on `PATH`, and sets `GOTOOLCHAIN=local` for later steps.

| Input | Default | Meaning |
|---|---|---|
| `version` | — (required) | Exact fgm release tag, e.g. `v0.4.0` |
| `working-directory` | `.` | Directory whose toolchain is resolved |
| `install` | `true` | Install missing resolved toolchains |
| `cache` | `true` | Cache the resolved toolchains; key: OS, arch, Go and linter versions |
| `token` | `github.token` | GitHub token for `golangci-lint` release lookups (unauthenticated GitHub API calls are limited to 60/hour per runner IP) |
| `fgm-binary` | — | Use a prebuilt fgm instead of downloading (used to test fgm itself) |

Outputs: `go-version`, `go-source`, `golangci-lint-version`, `golangci-lint-source`, `cache-hit`.

The action uses `jq`, which GitHub-hosted Linux and macOS runners provide.

## Example: consistency job without linting

`fgm check` reads repository files only, so this job needs no Go toolchain download (`install: false`, `cache: false`):

```yaml
jobs:
  toolchain-consistency:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: kostikovk/fgm/setup-fgm@v0.4.0
        with:
          version: v0.4.0
          install: false
          cache: false
      - run: fgm check --all
```

## Example: feed the resolved Go into existing `setup-go` steps

Keep `actions/setup-go` (and its module cache) and let fgm decide the version, so `.fgm.toml` policies such as `go = "latest-patch"` apply in CI too:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - id: fgm
        uses: kostikovk/fgm/setup-fgm@v0.4.0
        with:
          version: v0.4.0
          install: false
          cache: false
      - uses: actions/setup-go@v7
        with:
          go-version: ${{ steps.fgm.outputs.go-version }}
      - run: go test ./...
      - uses: golangci/golangci-lint-action@v8
        if: steps.fgm.outputs.golangci-lint-version != ''
        with:
          version: ${{ steps.fgm.outputs.golangci-lint-version }}
```
