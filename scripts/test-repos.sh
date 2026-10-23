#!/bin/sh
# Build the signed apt/dnf repos from goreleaser's dist/ with a throwaway key, then install
# gitwhy from them in clean Ubuntu and Fedora containers. Needs docker.
# Usage: scripts/test-repos.sh [dist dir]
set -eu

dist=$(realpath "${1:-dist}")
scripts=$(realpath "$(dirname "$0")")
site=$(mktemp -d)
chmod 755 "$site"

echo "== build repos"
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
docker run --rm -v "$site:/site:ro" ubuntu:24.04 sh -ec "$aptconf"'
  apt-get update -qq
  apt-get install -y -qq gitwhy >/dev/null
  gitwhy version'

# Containers mount the repo read-only, so tampering happens on a copy.
bad="$site.bad"
trap 'rm -rf "$site" "$bad"' EXIT
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
  gitwhy version'

echo "== dnf rejects a tampered repo"
out=$(docker run --rm -v "$bad:/site:ro" fedora:latest sh -c "$dnfconf"'; dnf install -y gitwhy 2>&1; true')
echo "$out" | grep -qiE "bad GPG signature|signature verification error: bad" || { echo "FAIL: dnf did not reject tampered metadata:"; echo "$out"; exit 1; }
echo "$out" | grep -iE "BADSIG|not signed|signature|GPG|verif" | head -2
echo "ok"
echo "all repo checks passed"
