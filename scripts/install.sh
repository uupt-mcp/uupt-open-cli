#!/bin/bash
# UU跑腿 CLI 安装脚本

set -e

# 核心变量
# GitHub Release 加速代理。国内优先 ghfast.top，失败再直连和其它代理。
GITHUB_PROXY_FAST="https://ghfast.top/"
BASE_URL="https://github.com/uupt-mcp/uupt-open-cli/releases/download"
LATEST_VERSION_URL="https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/refs/heads/main/latest"
DOWNLOAD_DIR="${TMPDIR:-/tmp}/uupt-open-cli-downloads"
INSTALL_DIR="$HOME/.uupt-open-cli"
BINARY_NAME="uupt-open-cli"

# 颜色输出
RED="\033[0;31m"
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
NC="\033[0m"

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# 参数验证
validate_version() {
  if [[ ! "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    error "无效的版本号格式: $1 (期望格式: X.Y.Z)"
  fi
}

# 平台检测
detect_platform() {
  local os arch

  case "$(uname -s)" in
    Darwin)  os="macos" ;;
    Linux)   os="linux" ;;
    MINGW*|MSYS*|CYGWIN*)
      echo "[ERROR] 检测到 Windows 环境，请使用 PowerShell 安装脚本:"
      echo "  curl.exe -fsSL --tlsv1.2 -o %TEMP%\\uupt-install.ps1 https://cdn.jsdelivr.net/gh/uupt-mcp/uupt-open-cli@main/scripts/install.ps1"
      exit 1
      ;;
    *)       error "不支持的操作系统: $(uname -s)" ;;
  esac

  case "$(uname -m)" in
    x86_64)       arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *)            error "不支持的架构: $(uname -m)" ;;
  esac

  # Apple Silicon: 检查是否通过 Rosetta 2 运行
  if [ "$os" = "macos" ] && [ "$arch" = "amd64" ]; then
    if sysctl -n sysctl.proc_translated 2>/dev/null | grep -q 1; then
      warn "检测到通过 Rosetta 2 运行，将使用 arm64 版本"
      arch="arm64"
    fi
  fi

  PLATFORM="$os"
  ARCH="$arch"
}

# 获取版本
get_version() {
  if [ -n "$1" ]; then
    validate_version "$1"
    VERSION="$1"
    info "使用指定版本: v${VERSION}"
  else
    info "获取最新版本..."
    VERSION=""
    for url in \
      "https://cdn.jsdelivr.net/gh/uupt-mcp/uupt-open-cli@main/latest" \
      "${GITHUB_PROXY_FAST}${LATEST_VERSION_URL}" \
      "https://ghproxy.net/${LATEST_VERSION_URL}" \
      "$LATEST_VERSION_URL"
    do
      if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL --connect-timeout 10 --max-time 30 "$url" 2>/dev/null | tr -d '\n\r')
      elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -q --timeout=30 -O- "$url" 2>/dev/null | tr -d '\n\r')
      else
        error "需要 curl 或 wget 来获取版本号，请先安装其中之一"
      fi
      if [ -n "$VERSION" ]; then
        break
      fi
      warn "获取版本失败: $url"
    done
    if [ -z "$VERSION" ]; then
      error "无法获取最新版本号"
    fi
    info "最新版本: v${VERSION}"
  fi
}

# 下载
download() {
  local filename="uupt-open-cli-${VERSION}-${PLATFORM}-${ARCH}.tar.gz"
  local url="${BASE_URL}/v${VERSION}/${filename}"
  local dest="${DOWNLOAD_DIR}/${filename}"

  mkdir -p "$DOWNLOAD_DIR"

  info "下载 ${filename}..."
  downloaded=0
  for candidate in \
    "${GITHUB_PROXY_FAST}${url}" \
    "$url" \
    "https://ghproxy.net/${url}" \
    "https://mirror.ghproxy.com/${url}"
  do
    info "尝试 $candidate"
    if command -v curl >/dev/null 2>&1; then
      if curl -fSL --connect-timeout 10 --max-time 180 --retry 1 \
        --speed-limit 20480 --speed-time 20 --progress-bar "$candidate" -o "$dest"; then
        downloaded=1
        break
      fi
    elif command -v wget >/dev/null 2>&1; then
      if wget --timeout=180 --progress=bar:force "$candidate" -O "$dest"; then
        downloaded=1
        break
      fi
    else
      error "需要 curl 或 wget 来下载文件，请先安装其中之一"
    fi
    warn "失败: $candidate"
  done
  if [ "$downloaded" -ne 1 ]; then
    error "下载失败: $url"
  fi

  ARCHIVE_PATH="$dest"
  info "下载完成: $dest"
}

# 解压安装
install() {
  local tmp_dir
  tmp_dir=$(mktemp -d)

  info "解压文件..."
  tar -xzf "$ARCHIVE_PATH" -C "$tmp_dir"

  # 查找二进制文件
  local binary
  binary=$(find "$tmp_dir" -name "uupt-open-cli*" -type f ! -name "*.json" | head -1)
  if [ -z "$binary" ]; then
    error "未找到二进制文件"
  fi

  if [ -d "$INSTALL_DIR" ] && [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    # 升级安装：仅替换二进制，保留 configs/config.json 和 logs/
    info "检测到已有安装，执行升级..."
    cp "$binary" "$INSTALL_DIR/$BINARY_NAME"
  else
    # 全新安装：创建目录，复制所有文件
    info "执行全新安装..."
    mkdir -p "$INSTALL_DIR/configs"
    mkdir -p "$INSTALL_DIR/logs"
    cp "$binary" "$INSTALL_DIR/$BINARY_NAME"

    # 复制 configs
    if [ -d "$tmp_dir/configs" ]; then
      cp -r "$tmp_dir/configs/"* "$INSTALL_DIR/configs/"
    fi
  fi

  # 清理临时目录
  rm -rf "$tmp_dir"

  info "安装到: $INSTALL_DIR/$BINARY_NAME"
}

# 设置权限
set_permissions() {
  chmod +x "$INSTALL_DIR/$BINARY_NAME"
  info "已设置执行权限"
}

# 配置 PATH
configure_path() {
  local path_entry="export PATH=\"$INSTALL_DIR:\$PATH\""

  # 检查是否已在 PATH 中
  if echo "$PATH" | grep -q "$INSTALL_DIR"; then
    info "PATH 已包含安装目录"
    return
  fi

  local configured=false

  # bash
  if [ -f "$HOME/.bashrc" ]; then
    if ! grep -q "$INSTALL_DIR" "$HOME/.bashrc"; then
      echo "" >> "$HOME/.bashrc"
      echo "# UU跑腿 CLI" >> "$HOME/.bashrc"
      echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$HOME/.bashrc"
      info "已添加到 ~/.bashrc"
      configured=true
    fi
  fi

  # zsh
  if [ -f "$HOME/.zshrc" ]; then
    if ! grep -q "$INSTALL_DIR" "$HOME/.zshrc"; then
      echo "" >> "$HOME/.zshrc"
      echo "# UU跑腿 CLI" >> "$HOME/.zshrc"
      echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$HOME/.zshrc"
      info "已添加到 ~/.zshrc"
      configured=true
    fi
  fi

  # fish
  local fish_config="$HOME/.config/fish/config.fish"
  if [ -f "$fish_config" ]; then
    if ! grep -q "$INSTALL_DIR" "$fish_config"; then
      echo "" >> "$fish_config"
      echo "# UU跑腿 CLI" >> "$fish_config"
      echo "set -gx PATH $INSTALL_DIR \$PATH" >> "$fish_config"
      info "已添加到 $fish_config"
      configured=true
    fi
  fi

  if [ "$configured" = false ]; then
    # 如果没有找到任何 shell 配置文件，尝试创建 .bashrc
    if [ ! -f "$HOME/.bashrc" ] && [ ! -f "$HOME/.zshrc" ]; then
      echo "# UU跑腿 CLI" >> "$HOME/.bashrc"
      echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$HOME/.bashrc"
      info "已创建并添加到 ~/.bashrc"
    fi
  fi

  warn "请重新打开终端或运行 source 使 PATH 生效"
}

# 验证安装
verify_install() {
  export PATH="$INSTALL_DIR:$PATH"
  if [ ! -x "$INSTALL_DIR/$BINARY_NAME" ]; then
    error "未找到 CLI 可执行文件: $INSTALL_DIR/$BINARY_NAME"
  fi
  if "$INSTALL_DIR/$BINARY_NAME" --version >/dev/null 2>&1; then
    local ver
    ver=$("$INSTALL_DIR/$BINARY_NAME" --version 2>&1)
    info "安装验证成功: $ver"
  else
    error "无法运行 $INSTALL_DIR/$BINARY_NAME --version"
  fi
}

# 主流程
main() {
  echo ""
  echo "========================================="
  echo "  UU跑腿 CLI 安装程序"
  echo "========================================="
  echo ""

  get_version "$1"
  detect_platform
  info "平台: ${PLATFORM}/${ARCH}"

  download
  install
  set_permissions
  configure_path
  verify_install

  echo ""
  info "安装完成！运行 '${BINARY_NAME} --help' 开始使用"
  echo ""
}

main "$@"
