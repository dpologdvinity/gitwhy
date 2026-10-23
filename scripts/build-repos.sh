#!/bin/sh
# Build signed apt and dnf repositories from release packages.
# Usage: build-repos.sh <dir with .deb/.rpm files> <site dir>
# Needs apt-ftparchive (apt-utils), createrepo_c and gpg with a secret key
# (GPG_KEY_ID picks one; GPG_PASSPHRASE unlocks it). The site dir keeps older versions.
set -eu

pkgs=${1:?package dir}
site=${2:?site dir}
mkdir -p "$site/apt" "$site/rpm"

gpg_sign() { # gpg_sign <extra gpg args...>
  gpg --batch --yes --pinentry-mode loopback ${GPG_KEY_ID:+--local-user "$GPG_KEY_ID"} \
    ${GPG_PASSPHRASE:+--passphrase "$GPG_PASSPHRASE"} "$@"
}

for f in "$pkgs"/*.deb; do [ -e "$f" ] && cp "$f" "$site/apt/"; done
for f in "$pkgs"/*.rpm; do [ -e "$f" ] && cp "$f" "$site/rpm/"; done

# apt: a flat repository. Packages and Release list every .deb in the folder.
(
  cd "$site/apt"
  apt-ftparchive packages . > Packages
  gzip -k -f Packages
  apt-ftparchive \
    -o APT::FTPArchive::Release::Origin=gitwhy \
    -o APT::FTPArchive::Release::Label=gitwhy \
    -o APT::FTPArchive::Release::Architectures="amd64 arm64" \
    release . > Release
  rm -f InRelease Release.gpg
  gpg_sign --clearsign --output InRelease Release
  gpg_sign --armor --detach-sign --output Release.gpg Release
)

# dnf: repodata signed with a detached signature (repo_gpgcheck=1).
createrepo_c --quiet "$site/rpm"
rm -f "$site/rpm/repodata/repomd.xml.asc"
gpg_sign --armor --detach-sign --output "$site/rpm/repodata/repomd.xml.asc" "$site/rpm/repodata/repomd.xml"

gpg --armor --export ${GPG_KEY_ID:-} > "$site/gitwhy.asc"

base=${SITE_URL:-https://dpologdvinity.github.io/gitwhy}
cat > "$site/gitwhy.repo" <<REPO
[gitwhy]
name=gitwhy
baseurl=$base/rpm
enabled=1
gpgcheck=0
repo_gpgcheck=1
gpgkey=$base/gitwhy.asc
REPO
