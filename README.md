# Fast Go Manager (fgm)

Manage project Go versions and compatible `golangci-lint` toolchains from one CLI.

FGM reads `go.work`, `go.mod`, and `toolchain` directives to select Go for each repository. It installs missing toolchains, imports existing installations, and generates or audits lint configuration.

- Use the Go version your project already declares.
- Select a compatible linter automatically or keep an explicit project pin.
- Run project commands with the selected tools through `fgm exec`.
- Route plain `go` and `golangci-lint` commands through optional shell shims on macOS and Linux.

## Get started

Choose manual installation or let your AI coding agent set up the project.

### Install manually

With Homebrew on macOS or Linux:

```bash
brew install kostikovk/fgm/fgm
```

Or use the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/kostikovk/fgm/main/install.sh | sh
```

The script installs to `~/.local/bin`; ensure that directory is on `PATH`. See [installation](docs/installation.md) for requirements, release downloads, building from source, shell setup, and removal.

### Set up your project with one prompt

Open your Go repository in an AI coding agent and copy the [setup prompt](docs/ai-agents.md#set-up-your-project-with-one-prompt). It asks the agent to install FGM and its skill, inspect the repository, configure Go and lint tooling, run checks, and report each stage's results or blockers. The agent needs terminal, file, and network access.

## Quick start

From an existing Go module or workspace:

```bash
cd /path/to/your-go-project
fgm import auto                 # Reuse existing Go and linter installations
fgm current                     # Inspect selected versions and their sources
fgm install                     # Install missing selected toolchains
fgm exec -- go test ./...        # Run with the project's Go version
```

FGM uses project metadata before the global default. This workflow requires no shell profile changes. [Getting started](docs/getting-started.md) covers lint setup, verification, and workspace directories.

## Documentation

| I want to… | Read |
|---|---|
| Install, configure my shell, or uninstall FGM | [Installation](docs/installation.md) |
| Set up an existing Go project and run checks | [Getting started](docs/getting-started.md) |
| Set everything up through an AI agent | [AI coding agents and copyable prompt](docs/ai-agents.md) |
| Look up commands, version rules, pins, or lint configuration | [CLI and configuration reference](docs/reference.md) |
| Build or contribute to FGM | [Development](docs/development.md) |

[Website](https://kostikovk.github.io/fgm/) · [Browse documentation](docs/README.md) · [Releases](https://github.com/kostikovk/fgm/releases) · [Report an issue](https://github.com/kostikovk/fgm/issues)

## License

[MIT](LICENSE).
