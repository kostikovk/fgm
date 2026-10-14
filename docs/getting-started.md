# Set up an existing Go project

[Home](../README.md) · [Documentation](README.md)

This guide assumes [FGM is installed](installation.md) and your repository contains `go.mod` or `go.work`. For automated setup, use the [AI agent prompt](ai-agents.md#set-up-your-project-with-one-prompt).

## 1. Select the project

Open the module or workspace directory:

```bash
cd /path/to/your-go-project
fgm current
```

FGM reads the nearest `go.work`, then `go.mod`, preferring `toolchain` over `go` directives. The output identifies the resolved Go and linter versions and their sources. [Version resolution](reference.md#version-resolution) explains the full selection rules.

If no Go version can be resolved, check that you are inside the intended project. Outside Go projects, a [global default](installation.md#shell-integration) must already be configured.

## 2. Install the selected toolchains

Import usable existing installations, then install the selected versions:

```bash
fgm import auto
fgm install
```

Importing registers symlinks to existing tools. Keep those source installations available. Installing fills the FGM store without changing the global Go default.

## 3. Run project checks

```bash
fgm exec -- go test ./...
fgm exec -- go build ./...
```

Use your repository's existing scripts or Make targets where provided:

```bash
fgm exec -- make test
```

`fgm exec` sets the toolchain `PATH` for that command. It works in interactive shells, CI, and agent sessions without editing shell profiles. In a multi-module workspace, run the repository's workspace checks or run checks in each relevant module directory.

## 4. Set up linting

If the project already pins a linter in `.fgm.toml`, keep that selection. Without an explicit pin, FGM already selects a compatible linter automatically. You can record that policy in `.fgm.toml`:

```bash
fgm pin golangci-lint auto
fgm install
```

Before pinning, check where `.fgm.toml` will be written: FGM updates the nearest existing config, which may belong to an ancestor outside your repository. Preserve shared configs unless changing them is intended.

If `fgm current` reports no selected linter, writing `auto` will not add compatible versions to the catalog. Report the compatibility gap and continue Go checks; run lint checks only after a selected compatible linter is installed.

Audit an existing `.golangci.yml` or `.golangci.yaml`, including configuration inherited from an ancestor directory:

```bash
fgm lint doctor
```

If no config exists, create one:

```bash
fgm lint init --preset standard
```

After confirming the selected linter is installed, verify its actual version against `fgm current`, then run it:

```bash
fgm exec -- golangci-lint version
fgm exec -- golangci-lint run ./...
```

`lint doctor` prints diagnostic findings and can exit successfully even with `ERROR` findings. Read the output. Preserve existing configuration; generation does not migrate it. See [lint configuration](reference.md#lint-configuration) for presets, import ordering, and overwrite behavior.

## 5. Verify the setup

```bash
fgm current
fgm exec -- go version
fgm doctor
```

Inspect any findings. A shell integration finding is separate from running commands through `fgm exec`; follow [shell setup](installation.md#shell-integration) if you want plain `go` commands to use FGM.

## Work from another directory

Pass the target directory explicitly:

```bash
fgm --chdir /path/to/project current
fgm --chdir /path/to/project exec -- go test ./...
```

`FGM_CHDIR` supplies a default directory when `--chdir` is omitted. The explicit flag takes precedence, and `fgm exec` uses the selected directory as the child process's working directory.

Next: [CLI reference](reference.md) for command flags, pins, and upgrade behavior.
