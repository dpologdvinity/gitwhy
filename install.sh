#!/bin/sh
# Install the latest gitwhy release into ~/.local/bin (or $GITWHY_BIN_DIR).
# Usage: curl -fsSL https://raw.githubusercontent.com/dpologdvinity/gitwhy/main/install.sh | sh
set -eu

repo="dpologdvinity/gitwhy"
bin_dir="${GITWHY_BIN_DIR:-$HOME/.local/bin}"

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) echo "gitwhy: unsupported OS $(uname -s) (Linux and macOS only for now)" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) echo "gitwhy: unsupported CPU $(uname -m)" >&2; exit 1 ;;
esac

# The "latest" release URL redirects to .../tag/vX.Y.Z.
tag=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest" | sed 's|.*/tag/||')
case "$tag" in v*) ;; *) echo "gitwhy: couldn't find the latest release" >&2; exit 1 ;; esac
version=${tag#v}
file="gitwhy_${version}_${os}_${arch}.tar.gz"
base="https://github.com/$repo/releases/download/$tag"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsSL "$base/$file" -o "$tmp/$file"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt"

want=$(grep " $file\$" "$tmp/checksums.txt" | cut -d' ' -f1)
if command -v sha256sum >/dev/null 2>&1; then
  got=$(sha256sum "$tmp/$file" | cut -d' ' -f1)
else
  got=$(shasum -a 256 "$tmp/$file" | cut -d' ' -f1)
fi
if [ -z "$want" ] || [ "$want" != "$got" ]; then
  echo "gitwhy: checksum mismatch for $file; not installing" >&2
  exit 1
fi

tar -xzf "$tmp/$file" -C "$tmp" gitwhy
mkdir -p "$bin_dir"
mv "$tmp/gitwhy" "$bin_dir/gitwhy"
chmod +x "$bin_dir/gitwhy"
echo "Installed gitwhy $version to $bin_dir/gitwhy"
case ":$PATH:" in
  *":$bin_dir:"*) ;;
  *) echo "Add it to your PATH: export PATH=\"$bin_dir:\$PATH\"" ;;
esac
