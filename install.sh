#!/bin/sh
# Install the latest gitwhy release into ~/.local/bin (or $GITWHY_BIN_DIR).
# Usage: curl -fsSL https://dpologdvinity.github.io/gitwhy/install.sh | sh
#
# Linux and macOS, amd64 or arm64. Downloads the release tar.gz, checks its SHA-256 against the
# release's checksums.txt, and installs nothing if it does not match. Windows: use install.ps1.
#
# Needs: a POSIX sh, curl, tar, sed, grep, and sha256sum or shasum.
# Environment: GITWHY_BIN_DIR sets the install folder (default ~/.local/bin).
set -eu

repo="dpologdvinity/gitwhy"
bin_dir="${GITWHY_BIN_DIR:-$HOME/.local/bin}"

# Map uname names to the GOOS and GOARCH names in the release file names.
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

# The "latest" release URL redirects to .../tag/vX.Y.Z. Reading the redirect avoids the GitHub
# API and its rate limit.
tag=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest" | sed 's|.*/tag/||')
case "$tag" in v*) ;; *) echo "gitwhy: couldn't find the latest release" >&2; exit 1 ;; esac
version=${tag#v}
# goreleaser's default archive name: the version has no leading "v".
file="gitwhy_${version}_${os}_${arch}.tar.gz"
base="https://github.com/$repo/releases/download/$tag"

# Download into a temporary folder that is removed however the script exits.
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsSL "$base/$file" -o "$tmp/$file"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt"

# Verify before installing: the line for our archive in checksums.txt must match its SHA-256.
# Linux has sha256sum; macOS has shasum.
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

# Take only the binary from the archive; the licenses and README stay behind.
tar -xzf "$tmp/$file" -C "$tmp" gitwhy
mkdir -p "$bin_dir"
mv "$tmp/gitwhy" "$bin_dir/gitwhy"
chmod +x "$bin_dir/gitwhy"
echo "Installed gitwhy $version to $bin_dir/gitwhy"
# Only print a hint: a piped installer should not edit the user's shell startup files.
case ":$PATH:" in
  *":$bin_dir:"*) ;;
  *) echo "Add it to your PATH: export PATH=\"$bin_dir:\$PATH\"" ;;
esac
