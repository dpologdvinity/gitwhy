#!/bin/sh
# Build the signed apt/dnf repos from goreleaser's dist/ with a throwaway key, then install
# gitwhy from them in clean Ubuntu and Fedora containers. Also checks that each package asks for
# git 2.40 or newer, and that apt and dnf reject a tampered repo.
#
# Usage:    scripts/test-repos.sh [dist dir]     (default: dist)
# Needs:    docker, and a dist/ from `goreleaser release --snapshot --clean --skip=publish`.
# CI runs it in the `packages` job of .github/workflows/ci.yml.
set -eu

dist=$(realpath "${1:-dist}")
scripts=$(realpath "$(dirname "$0")")

# Set the cleanup trap before the first temp dir exists, so an early failure leaves nothing behind.
site=
bad=
trap 'rm -rf ${site:+"$site"} ${bad:+"$bad"}' EXIT
site=$(mktemp -d)
bad="$site.bad"
chmod 755 "$site"

echo "== build repos"
# The key has no passphrase; OWNER lets us delete the files the container creates as root.
docker run --rm -v "$scripts:/scripts:ro" -v "$dist:/dist:ro" -v "$site:/site" -e OWNER="$(id -u)" ubuntu:24.04 sh -ec '
  apt-get update -qq && apt-get install -y -qq apt-utils createrepo-c gnupg >/dev/null
  export GNUPGHOME=$(mktemp -d)
  gpg --batch --passphrase "" --quick-generate-key "gitwhy test <test@example.com>" default default never 2>/dev/null
  mkdir /pk && cp /dist/*.deb /dist/*.rpm /pk/
  sh /scripts/build-repos.sh /pk /site
  chown -R "$OWNER" /site; chmod -R a+rX /site'

aptconf='echo "deb [signed-by=/site/gitwhy.asc] file:/site/apt ./" > /etc/apt/sources.list.d/gitwhy.list'
dnfconf='sed -e "s#baseurl=.*#baseurl=file:///site/rpm#" -e "s#gpgkey=.*#gpgkey=file:///site/gitwhy.asc#" /site/gitwhy.repo > /etc/yum.repos.d/gitwhy.repo'

echo "== apt install (Ubuntu)"
# The package must depend on git >= 2.40 (gitwhy refuses older git), and the git apt
# installs must meet it.
docker run --rm -v "$site:/site:ro" ubuntu:24.04 sh -ec "$aptconf"'
  apt-get update -qq
  apt-get install -y -qq gitwhy >/dev/null
  dpkg-query -W -f="\${Depends}\n" gitwhy | grep -F "git (>= 1:2.40)"
  git --version
  gitwhy version'

# Containers mount the repo read-only, so tampering happens on a copy.
cp -r "$site" "$bad"
sed -i 's/^Origin: gitwhy/Origin: evil/' "$bad/apt/InRelease"
echo "tampered" >> "$bad/rpm/repodata/repomd.xml"

echo "== apt rejects a tampered repo"
out=$(docker run --rm -v "$bad:/site:ro" ubuntu:24.04 sh -c "$aptconf"'; apt-get update 2>&1; true')
echo "$out" | grep -qE "BADSIG|not signed|signature" || { echo "FAIL: apt did not reject a tampered InRelease:"; echo "$out"; exit 1; }
echo "$out" | grep -iE "BADSIG|not signed|signature|GPG|verif" | head -2
echo "ok"

echo "== dnf install (Fedora)"
docker run --rm -v "$site:/site:ro" fedora:latest sh -ec "$dnfconf"'
  dnf install -y -q gitwhy >/dev/null 2>&1
  rpm -q --requires gitwhy | grep -F "git >= 2.40"
  git --version
  gitwhy version'

echo "== dnf rejects a tampered repo"
out=$(docker run --rm -v "$bad:/site:ro" fedora:latest sh -c "$dnfconf"'; dnf install -y gitwhy 2>&1; true')
echo "$out" | grep -qiE "bad GPG signature|signature verification error: bad" || { echo "FAIL: dnf did not reject tampered metadata:"; echo "$out"; exit 1; }
echo "$out" | grep -iE "BADSIG|not signed|signature|GPG|verif" | head -2
echo "ok"
echo "all repo checks passed"
