#!/bin/sh
set -eu

# post-goreleaser.sh — Post-build packaging for skills zip.
#
# Run after `goreleaser release` to create uupt-skills.zip and update checksums.

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
DIST_DIR="${UUPT_PACKAGE_DIST_DIR:-$ROOT/dist}"
PACKAGE_VERSION="${UUPT_PACKAGE_VERSION:-}"

export LANG=C
export LC_ALL=C
export LC_CTYPE=C

say() {
  printf '%s\n' "$*"
}

err() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

sha256_file() {
  target="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$target" | awk '{print $1}'
    return
  fi
  shasum -a 256 "$target" | awk '{print $1}'
}

# ---------- skills zip ----------

create_skills_zip() {
  skills_zip="$DIST_DIR/uupt-skills.zip"
  rm -f "$skills_zip"
  (
    cd "$ROOT/skills"
    env -u LC_ALL -u LC_CTYPE LANG=C LC_ALL=C LC_CTYPE=C zip -qr "$skills_zip" .
  )
}

write_checksums() {
  checksum_path="$DIST_DIR/checksums.txt"
  # Append skills zip checksum to goreleaser's checksums file
  if [ -f "$DIST_DIR/uupt-skills.zip" ]; then
    printf '%s  %s\n' "$(sha256_file "$DIST_DIR/uupt-skills.zip")" "uupt-skills.zip" >> "$checksum_path"
  fi
}

# ---------- main ----------

say "==> Creating skills zip"
create_skills_zip

say "==> Updating checksums"
write_checksums

say ""
say "Post-goreleaser packaging complete:"
say "  skills: $DIST_DIR/uupt-skills.zip"
