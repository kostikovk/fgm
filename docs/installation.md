# Installation and removal

[Home](../README.md) · [Documentation](README.md)

## Homebrew

For macOS and Linux:

```bash
brew install kostikovk/fgm/fgm
```

## Install script

Downloads the latest release, verifies its checksum, and installs to `~/.local/bin` by default. Requires `curl`, a SHA-256 checksum utility, and `tar` (or `unzip` for Windows archives).

```bash
curl -fsSL https://raw.githubusercontent.com/kostikovk/fgm/main/install.sh | sh
```

## Build from source

Requires a Go toolchain compatible with the repository's [go.mod](../go.mod).

```bash
git clone https://github.com/kostikovk/fgm.git
cd fgm
mkdir -p ~/.local/bin
go build -o ~/.local/bin/fgm .
```

Ensure the installation directory is on `PATH`, then verify:

```bash
fgm version
```

If the default directory is outside `PATH`, invoke `~/.local/bin/fgm` directly until you add it to your shell's `PATH`. For a custom install location, pass `INSTALL_DIR` to the script:

```bash
curl -fsSL https://raw.githubusercontent.com/kostikovk/fgm/main/install.sh | INSTALL_DIR=/your/bin sh
```

You can also download an archive and `checksums.txt` from [GitHub Releases](https://github.com/kostikovk/fgm/releases). On Windows, use `fgm exec` to run project commands; automatic `go` shim generation is supported only on macOS and Linux.

## Shell integration

For project commands, [getting started](getting-started.md) uses `fgm exec` and requires no shell profile changes. To route plain `go` commands through FGM on macOS or Linux, first install or import your chosen default Go version:

```bash
fgm install go <version>
fgm use go <version> --global
```

Replace `<version>` with an actual Go version. The global default applies outside Go projects; project metadata takes precedence inside a repository.

For the current Bash or Zsh session:

```bash
eval "$(fgm env)"
```

For Fish:

```fish
fgm env --shell fish | source
```

For persistent integration, `fgm doctor --fix` can add the snippet to your shell profile. [Shell profile behavior](reference.md#shell-profile-setup) lists the affected files and confirmation behavior.

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

**Remove shell integration** — delete the FGM snippet from your shell profile (`.zshrc`, `.bashrc`, etc.), including `eval "$(fgm env)"` for Bash/Zsh or `fgm env --shell fish | source` for Fish.

For script or source installations, remove the FGM binary from the directory where you installed it. Removing the managed store deletes FGM toolchains and state; imported host installations remain intact.

Next: [Set up an existing Go project](getting-started.md).
