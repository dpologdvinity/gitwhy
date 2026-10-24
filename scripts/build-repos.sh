#!/bin/sh
# Build signed apt and dnf repositories from release packages.
#
# Usage:    build-repos.sh <dir with .deb/.rpm files> <site dir>
# Needs:    apt-ftparchive (apt-utils), createrepo_c and gpg with a secret key.
# Env:      GPG_KEY_ID      which secret key signs (default: gpg's default key)
#           GPG_PASSPHRASE  unlocks the key (optional; fed on stdin, never on the command line)
#           SITE_URL        where the site is served (default: https://dpologdvinity.github.io/gitwhy)
# The site dir keeps older versions, so users can still install them.
# The release workflow runs it in the `repos` job; scripts/test-repos.sh runs it with a test key.
set -eu

# Both arguments are required; ${1:?} stops with a message when one is missing.
pkgs=${1:?package dir}
site=${2:?site dir}
mkdir -p "$site/apt" "$site/rpm"

# gpg_sign <extra gpg args...>. The passphrase goes through a pipe from printf (a shell builtin),
# so it never shows up in the process list the way a --passphrase argument would.
gpg_sign() {
  if [ -n "${GPG_PASSPHRASE:-}" ]; then
    printf '%s' "$GPG_PASSPHRASE" | gpg --batch --yes --pinentry-mode loopback \
      ${GPG_KEY_ID:+--local-user "$GPG_KEY_ID"} --passphrase-fd 0 "$@"
  else
    gpg --batch --yes --pinentry-mode loopback ${GPG_KEY_ID:+--local-user "$GPG_KEY_ID"} "$@"
  fi
}

# Add this release's packages next to the older ones. The -e test skips the pattern itself
# when there are no files of that kind.
for f in "$pkgs"/*.deb; do [ -e "$f" ] && cp "$f" "$site/apt/"; done
for f in "$pkgs"/*.rpm; do [ -e "$f" ] && cp "$f" "$site/rpm/"; done

# apt: a flat repository. Packages and Release list every .deb in the folder.
(
  cd "$site/apt"
  apt-ftparchive packages . > Packages
  # Keep both: apt may fetch either Packages or Packages.gz.
  gzip -k -f Packages
  apt-ftparchive \
    -o APT::FTPArchive::Release::Origin=gitwhy \
    -o APT::FTPArchive::Release::Label=gitwhy \
    -o APT::FTPArchive::Release::Architectures="amd64 arm64" \
    release . > Release
  # Remove the old signatures first, so a failed signing never leaves one that matches the old
  # Release. Newer apt reads InRelease (signed inline); older apt reads Release.gpg.
  rm -f InRelease Release.gpg
  gpg_sign --clearsign --output InRelease Release
  gpg_sign --armor --detach-sign --output Release.gpg Release
)

# dnf: repodata signed with a detached signature (repo_gpgcheck=1). createrepo_c rebuilds the
# metadata from every .rpm in the folder.
createrepo_c --quiet "$site/rpm"
rm -f "$site/rpm/repodata/repomd.xml.asc"
gpg_sign --armor --detach-sign --output "$site/rpm/repodata/repomd.xml.asc" "$site/rpm/repodata/repomd.xml"

# The public key users download to trust the repos.
gpg --armor --export ${GPG_KEY_ID:+"$GPG_KEY_ID"} > "$site/gitwhy.asc"

# A ready-made .repo file for dnf users to download into /etc/yum.repos.d/.
# gpgcheck=0 because the .rpm files themselves are not signed. repo_gpgcheck=1 checks the signed
# repodata instead, and that lists the SHA-256 of every package, so tampering is still caught.
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
