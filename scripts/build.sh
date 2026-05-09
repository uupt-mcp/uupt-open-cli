#!/bin/bash
# UU跑腿 CLI 跨平台编译脚本

set -e

# 版本信息
VERSION=$(cat latest)
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# 输出目录
DIST_DIR="dist"
rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

# 目标平台
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

echo "=== UU跑腿 CLI 构建 ==="
echo "版本: ${VERSION}"
echo "提交: ${COMMIT}"
echo "日期: ${DATE}"
echo ""

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS="${PLATFORM%/*}"
  GOARCH="${PLATFORM#*/}"

  # 平台名称映射
  PLATFORM_NAME="${GOOS}"
  if [ "${GOOS}" = "darwin" ]; then
    PLATFORM_NAME="macos"
  fi

  # 二进制文件名
  BINARY_NAME="uupt-open-cli-${VERSION}-${PLATFORM_NAME}-${GOARCH}"
  if [ "${GOOS}" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
  fi

  echo "编译 ${PLATFORM_NAME}/${GOARCH}..."

  # 编译
  GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
    -ldflags "${LDFLAGS}" \
    -o "${DIST_DIR}/${BINARY_NAME}" \
    ./cmd/uupt-open-cli

  # 打包
  ARCHIVE_NAME="uupt-open-cli-${VERSION}-${PLATFORM_NAME}-${GOARCH}.tar.gz"
  STAGING_DIR="${DIST_DIR}/staging-${PLATFORM_NAME}-${GOARCH}"
  mkdir -p "${STAGING_DIR}/configs"

  # 复制文件到暂存目录
  cp "${DIST_DIR}/${BINARY_NAME}" "${STAGING_DIR}/"
  cp "configs/defaults.json" "${STAGING_DIR}/configs/"

  # 创建 tar.gz
  tar -czf "${DIST_DIR}/${ARCHIVE_NAME}" -C "${STAGING_DIR}" .

  # 清理暂存目录
  rm -rf "${STAGING_DIR}"

  echo "  -> ${DIST_DIR}/${ARCHIVE_NAME}"
done

echo ""
echo "=== 构建完成 ==="
echo "输出目录: ${DIST_DIR}/"
ls -lh "${DIST_DIR}"/*.tar.gz
