#!/bin/sh
set -eu

REPO="${GITHUB_REPOSITORY:-kostikovk/fgm}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  darwin) os="darwin" ;;
  linux) os="linux" ;;
  msys*|mingw*|cygwin*) os="windows" ;;
  *) echo "unsupported operating system: $os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac

api_url="https://api.github.com/repos/$REPO/releases/latest"
version="$(curl -fsSL "$api_url" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)"

if [ -z "$version" ]; then
  echo "failed to resolve latest release from $api_url" >&2
  exit 1
fi

archive="fgm_${version#v}_${os}_${arch}.tar.gz"
if [ "$os" = "windows" ]; then
  archive="fgm_${version#v}_${os}_${arch}.zip"
fi

base_url="https://github.com/$REPO/releases/download/$version"
archive_path="$TMP_DIR/$archive"
checksums_path="$TMP_DIR/checksums.txt"

curl -fsSL "$base_url/$archive" -o "$archive_path"
curl -fsSL "$base_url/checksums.txt" -o "$checksums_path"

expected="$(grep " $archive\$" "$checksums_path" | awk '{print $1}')"
if [ -z "$expected" ]; then
  echo "failed to find checksum for $archive" >&2
  exit 1
fi

if command -v shasum >/dev/null 2>&1; then
  actual="$(shasum -a 256 "$archive_path" | awk '{print $1}')"
elif command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$archive_path" | awk '{print $1}')"
else
  echo "missing shasum or sha256sum for checksum verification" >&2
  exit 1
fi

if [ "$actual" != "$expected" ]; then
  echo "checksum mismatch for $archive" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"

if [ "$os" = "windows" ]; then
  unzip -q "$archive_path" -d "$TMP_DIR"
  cp "$TMP_DIR/fgm.exe" "$INSTALL_DIR/fgm.exe"
else
  tar -xzf "$archive_path" -C "$TMP_DIR"
  cp "$TMP_DIR/fgm" "$INSTALL_DIR/fgm"
  chmod +x "$INSTALL_DIR/fgm"
fi

echo "Installed fgm $version to $INSTALL_DIR"
