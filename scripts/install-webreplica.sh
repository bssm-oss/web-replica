#!/usr/bin/env sh
set -eu

repo="bssm-oss/web-replica"
install_dir="${WEBREPLICA_INSTALL_DIR:-/usr/local/bin}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Required command not found: $1" >&2
    exit 1
  fi
}

need curl
need tar

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m | tr '[:upper:]' '[:lower:]')

case "$os" in
  darwin|linux) ;;
  *)
    echo "Unsupported OS: $os" >&2
    echo "Download manually from: https://github.com/$repo/releases/latest" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $arch" >&2
    echo "Download manually from: https://github.com/$repo/releases/latest" >&2
    exit 1
    ;;
esac

latest_json=$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest")
version=$(printf '%s' "$latest_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)

if [ -z "$version" ]; then
  echo "Could not determine latest release version" >&2
  exit 1
fi

package="webreplica_${version}_${os}_${arch}"
archive="${package}.tar.gz"
url="https://github.com/$repo/releases/download/${version}/${archive}"
tmp_dir=$(mktemp -d 2>/dev/null || mktemp -d -t webreplica)

cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

echo "Downloading $archive"
curl -fL "$url" -o "$tmp_dir/$archive"

tar -xzf "$tmp_dir/$archive" -C "$tmp_dir"

if [ ! -x "$tmp_dir/$package/install.sh" ]; then
  echo "Installer not found in release archive" >&2
  exit 1
fi

WEBREPLICA_INSTALL_DIR="$install_dir" "$tmp_dir/$package/install.sh"
