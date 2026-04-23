#!/usr/bin/env sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
source_bin="$script_dir/webreplica"

if [ ! -f "$source_bin" ]; then
  echo "webreplica binary not found next to install.sh" >&2
  exit 1
fi

install_dir="${WEBREPLICA_INSTALL_DIR:-/usr/local/bin}"
install_path="$install_dir/webreplica"

if [ -d "$install_dir" ] && [ -w "$install_dir" ]; then
  install -m 0755 "$source_bin" "$install_path"
elif command -v sudo >/dev/null 2>&1; then
  sudo mkdir -p "$install_dir"
  sudo install -m 0755 "$source_bin" "$install_path"
else
  install_dir="$HOME/.local/bin"
  install_path="$install_dir/webreplica"
  mkdir -p "$install_dir"
  install -m 0755 "$source_bin" "$install_path"
fi

echo "Installed webreplica to $install_path"

case ":$PATH:" in
  *":$install_dir:"*)
    echo "You can now run:"
    echo "  webreplica https://example.com"
    ;;
  *)
    echo "Add this directory to PATH, then run webreplica from anywhere:"
    echo "  export PATH=\"$install_dir:\$PATH\""
    echo "  webreplica https://example.com"
    ;;
esac
