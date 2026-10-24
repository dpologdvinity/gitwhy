#!/bin/sh
# Install the Alpine (.apk) and Arch (.pkg.tar.zst) packages from goreleaser's dist/ in clean
# containers and run gitwhy. Also checks that each package asks for git 2.40 or newer.
# (apt and dnf are covered by test-repos.sh.)
#
# Usage:    scripts/test-pkgs.sh [dist dir]     (default: dist)
# Needs:    docker, and a dist/ from `goreleaser release --snapshot --clean --skip=publish`.
# CI runs it in the `packages` job of .github/workflows/ci.yml.
set -eu

dist=$(realpath "${1:-dist}")
apk=$(basename "$(ls "$dist"/*_linux_amd64.apk)")
arch=$(basename "$(ls "$dist"/*_linux_amd64.pkg.tar.zst)")

echo "== apk (Alpine)"
# Our packages are unsigned, hence --allow-untrusted; the git dependency comes from Alpine's repo.
docker run --rm -v "$dist:/pk:ro" alpine:latest sh -ec "
  apk add --no-cache --allow-untrusted /pk/$apk >/dev/null
  apk info -R gitwhy | grep -F 'git>=2.40'
  git --version
  gitwhy version"

echo "== pacman (Arch)"
docker run --rm -v "$dist:/pk:ro" archlinux:latest sh -ec "
  pacman -Sy --noconfirm >/dev/null 2>&1
  pacman -U --noconfirm /pk/$arch >/dev/null 2>&1
  pacman -Qi gitwhy | grep -F 'git>=2.40'
  git --version
  gitwhy version"
echo "all package checks passed"
