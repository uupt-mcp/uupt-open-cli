#!/bin/sh
set -eu

# Install UU跑腿 CLI agent skills from GitHub Releases into agent skill directories.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install-skills.sh | sh
#
# Downloads uupt-skills.zip from GitHub Releases and copies it under each target
# path using agent directory detection (AGENT_DIRS + parent-directory gate),
# with root defaulting to the current directory. Set UUPT_SKILLS_ROOT=$HOME to
# match home-directory layout.
#
# Environment variables (optional):
#   UUPT_SKILLS_VERSION  — release tag (default: latest)
#   UUPT_SKILLS_ROOT     — base path for agent dirs (default: $PWD)

REPO="uupt-mcp/uupt-open-cli"
VERSION="${UUPT_SKILLS_VERSION:-latest}"
SKILL_NAME="uupt"
ROOT="${UUPT_SKILLS_ROOT:-$PWD}"

# ── Helpers ──────────────────────────────────────────────────────────────────

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf '❌ 缺少必要命令: %s\n' "$1" >&2
    exit 1
  fi
}

resolve_version() {
  if [ "$VERSION" = "latest" ]; then
    VERSION="$(curl -fsSI "https://github.com/${REPO}/releases/latest" 2>/dev/null \
      | grep -i '^location:' | sed 's|.*/tag/||;s/[[:space:]]*$//')"
    if [ -z "$VERSION" ]; then
      # 回退到 latest 文件
      VERSION="$(curl -fsSL "https://raw.githubusercontent.com/${REPO}/refs/heads/main/latest" 2>/dev/null | tr -d '\n\r')"
    fi
    if [ -z "$VERSION" ]; then
      printf '❌ 无法确定最新版本，请设置 UUPT_SKILLS_VERSION 环境变量\n' >&2
      exit 1
    fi
  fi
  # 确保 VERSION 以 'v' 开头
  case "$VERSION" in
    v*) ;;
    *) VERSION="v$VERSION" ;;
  esac
}

extract_zip() {
  archive="$1"
  dest="$2"
  if command -v unzip >/dev/null 2>&1; then
    unzip -q "$archive" -d "$dest"
    return 0
  fi
  printf '❌ 缺少必要命令: unzip\n' >&2
  exit 1
}

# One-line summary copy (2nd+ targets).
_copy_skill_summary() {
  _src="$1"
  _dest="$2"
  _label="$3"

  if [ -d "$_dest" ]; then
    rm -rf "$_dest"
  fi

  mkdir -p "$_dest"
  cp -R "$_src/"* "$_dest/" 2>/dev/null || cp -r "$_src/"* "$_dest/"
  file_count="$(find "$_dest" -type f | wc -l | tr -d ' ')"

  printf '  ✅ Skills → %s (%s 个文件)\n' "$_label" "$file_count"
}

# Full copy with top-level listing (1st target).
_copy_skill() {
  _src="$1"
  _dest="$2"
  _label="$3"

  if [ -d "$_dest" ]; then
    rm -rf "$_dest"
  fi

  mkdir -p "$_dest"
  cp -R "$_src/"* "$_dest/" 2>/dev/null || cp -r "$_src/"* "$_dest/"
  file_count="$(find "$_dest" -type f | wc -l | tr -d ' ')"

  printf '  ✅ Skills → %s (%s 个文件)\n' "$_label" "$file_count"

  for entry in "$_dest"/*; do
    entry_name="$(basename "$entry")"
    if [ -d "$entry" ]; then
      sub_count="$(find "$entry" -type f | wc -l | tr -d ' ')"
      printf '     📁 %s/ (%s 个文件)\n' "$entry_name" "$sub_count"
    else
      printf '     📄 %s\n' "$entry_name"
    fi
  done
}

# Install skills into all detected agent directories under a given root.
install_skills_to_root() {
  skill_src="$1"
  root="$2"
  installed=0
  idx=0
  for agent_dir in \
    ".agents/skills" \
    ".claude/skills" \
    ".cursor/skills" \
    ".gemini/skills" \
    ".codex/skills" \
    ".github/skills" \
    ".windsurf/skills" \
    ".augment/skills" \
    ".cline/skills" \
    ".amp/skills" \
    ".kiro/skills" \
    ".trae/skills" \
    ".openclaw/skills" \
    ".hermes/skills" \
    ".qoder/skills"
  do
    base_dir="$root/$agent_dir"
    parent_gate="$(dirname "$base_dir")"
    if [ "$idx" -gt 0 ] && [ ! -e "$parent_gate" ]; then
      idx=$((idx + 1))
      continue
    fi
    dest="$base_dir/$SKILL_NAME"
    if [ "$root" = "$HOME" ]; then
      label="~/$agent_dir/$SKILL_NAME"
    else
      label="$root/$agent_dir/$SKILL_NAME"
    fi
    if [ "$installed" -eq 0 ]; then
      _copy_skill "$skill_src" "$dest" "$label"
    else
      _copy_skill_summary "$skill_src" "$dest" "$label"
    fi
    installed=$((installed + 1))
    idx=$((idx + 1))
  done
  if [ "$installed" -eq 0 ]; then
    if [ "$root" = "$HOME" ]; then
      flabel="~/.agents/skills/$SKILL_NAME"
    else
      flabel="$root/.agents/skills/$SKILL_NAME"
    fi
    _copy_skill "$skill_src" "$root/.agents/skills/$SKILL_NAME" "$flabel"
  fi
}

# ── Main ─────────────────────────────────────────────────────────────────────

main() {
  need_cmd curl
  resolve_version

  printf '\n'
  printf '  ┌──────────────────────────────────────┐\n'
  printf '  │     UU跑腿 Skill 安装器              │\n'
  printf '  │     UU跑腿开放平台 CLI               │\n'
  printf '  └──────────────────────────────────────┘\n'
  printf '\n'

  TMPDIR_WORK="$(mktemp -d)"
  trap 'rm -rf "$TMPDIR_WORK"' EXIT INT TERM

  ASSET_URL="https://github.com/${REPO}/releases/download/${VERSION}/uupt-skills.zip"
  printf '  ⬇  从 GitHub Releases 下载 skills: %s (%s)\n' "$REPO" "$VERSION"
  curl -fsSL "$ASSET_URL" -o "$TMPDIR_WORK/uupt-skills.zip"
  extract_zip "$TMPDIR_WORK/uupt-skills.zip" "$TMPDIR_WORK/extracted"

  SKILL_SRC="$TMPDIR_WORK/extracted"
  if [ -f "$TMPDIR_WORK/extracted/${SKILL_NAME}/SKILL.md" ]; then
    SKILL_SRC="$TMPDIR_WORK/extracted/${SKILL_NAME}"
  fi

  if [ ! -f "$SKILL_SRC/SKILL.md" ]; then
    printf '  ❌ Release 资产中未找到 skill 文件\n' >&2
    exit 1
  fi

  printf '\n'
  printf '  安装到根目录: %s\n' "$ROOT"
  install_skills_to_root "$SKILL_SRC" "$ROOT"

  printf '\n'
  printf '  📖 Skill 包含:\n'
  printf '     • SKILL.md — 主 skill 文件（产品概览与意图路由）\n'
  printf '     • references/ — 详细的产品命令参考文档\n'
  printf '\n'
  printf '  ⚡ 前提条件: uupt-open-cli 已安装并在 $PATH 中\n'
  printf '     安装方式: curl -fsSL https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.sh | sh\n'
  printf '\n'
}

main
