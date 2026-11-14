#!/bin/bash
# Build a fnOS (飞牛) installable .fpk package for ezBookkeeping (ebk-fnos).
#
# Usage:
#   ./build-fnos-fpk.sh                 # build backend + frontend, then pack
#   ./build-fnos-fpk.sh --skip-frontend # reuse existing dist/ (CI / second run)
#   ./build-fnos-fpk.sh --no-pack       # assemble build/fnos only
#   ./build-fnos-fpk.sh --min-os-version 1.2.0401 # set manifest os_min_version
#
# Requirements: Go (>= 1.27 toolchain auto-download), GCC, Node.js + npm.
# Optional: fnpack (official fnOS packaging tool). Without it the script falls
# back to a tar.gz archive named .fpk, which fnOS can also install in most cases.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT_DIR}"

SKIP_FRONTEND=0
NO_PACK=0
MIN_OS_VERSION=

while [ $# -gt 0 ]; do
    case "$1" in
        --skip-frontend) SKIP_FRONTEND=1; shift ;;
        --no-pack) NO_PACK=1; shift ;;
        --min-os-version)
            if [ $# -lt 2 ] || [ -z "${2:-}" ]; then
                echo "--min-os-version requires a version value" >&2
                exit 1
            fi
            MIN_OS_VERSION="$2"
            shift 2
            ;;
        -h|--help)
            sed -n '2,12p' "$0" | sed 's/^# \{0,1\}//'
            exit 0
            ;;
        *)
            echo "未知参数：$1" >&2
            exit 1
            ;;
    esac
done

VERSION="$(grep '"version"' package.json | head -n 1 | cut -d ':' -f 2 | tr -d ' ' | tr -d ',' | tr -d '"')"
BUILD_DIR="${ROOT_DIR}/build/fnos"
OUT_DIR="${ROOT_DIR}/build"
FPK_NAME="ebk-fnos-${VERSION}-x86.fpk"

echo "==> 版本：${VERSION}"

# ------------------------------------------------------------------
# 1. 编译后端（CGO 开启，SQLite 依赖 cgo）
# ------------------------------------------------------------------
echo "==> 编译后端 ezbookkeeping"
mkdir -p "${BUILD_DIR}/app"
CGO_ENABLED=1 go build -trimpath \
    -ldflags "-w -s -X main.Version=${VERSION}" \
    -o "${BUILD_DIR}/app/ezbookkeeping" .

chmod +x "${BUILD_DIR}/app/ezbookkeeping"

# ------------------------------------------------------------------
# 2. 构建前端静态资源
# ------------------------------------------------------------------
if [ "${SKIP_FRONTEND}" -eq 1 ]; then
    echo "==> 跳过前端构建（复用 dist/）"
else
    if [ ! -d node_modules ]; then
        echo "==> 安装前端依赖"
        npm ci
    fi
    echo "==> 构建前端（内存不足时请调高 NODE_OPTIONS）"
    NODE_OPTIONS="${NODE_OPTIONS:---max-old-space-size=4096}" npm run build
fi

if [ ! -f dist/desktop.html ] || [ ! -f dist/mobile.html ]; then
    echo "错误：dist/ 下缺少 desktop.html / mobile.html，前端未构建成功。" >&2
    echo "      请在有足够内存的机器上执行 npm run build，或加 --skip-frontend 复用已有 dist。" >&2
    exit 1
fi

echo "==> 复制静态资源与模板"
rm -rf "${BUILD_DIR}/app/public" "${BUILD_DIR}/app/templates" "${BUILD_DIR}/app/conf"
cp -R dist "${BUILD_DIR}/app/public"
cp -R templates "${BUILD_DIR}/app/templates"
cp -R conf "${BUILD_DIR}/app/conf"
rm -rf "${BUILD_DIR}/app/public/.vite" 2>/dev/null || true

# ------------------------------------------------------------------
# 3. 组装 fpk 目录结构
# ------------------------------------------------------------------
echo "==> 组装应用包目录"
cp -R fnos/. "${BUILD_DIR}/"
cp LICENSE "${BUILD_DIR}/LICENSE" 2>/dev/null || true

# 与 package.json 中的版本保持一致
sed -i "s|^version = .*|version = ${VERSION}|" "${BUILD_DIR}/manifest"

# 允许打包时指定最低系统版本（manifest os_min_version）
if [ -n "${MIN_OS_VERSION}" ]; then
    if grep -q '^os_min_version' "${BUILD_DIR}/manifest"; then
        sed -i "s|^os_min_version = .*|os_min_version = ${MIN_OS_VERSION}|" "${BUILD_DIR}/manifest"
    else
        printf 'os_min_version = %s\n' "${MIN_OS_VERSION}" >> "${BUILD_DIR}/manifest"
    fi
    echo "==> 最低系统版本：${MIN_OS_VERSION}"
fi

chmod +x "${BUILD_DIR}/cmd/"* "${BUILD_DIR}/app/scripts/write_config.sh" "${BUILD_DIR}/app/ezbookkeeping"

echo "==> 包结构："
find "${BUILD_DIR}" -maxdepth 2 -type d | sed "s|${BUILD_DIR}|  .|"

if [ "${NO_PACK}" -eq 1 ]; then
    echo "==> 已跳过打包，目录位于 ${BUILD_DIR}"
    exit 0
fi

# ------------------------------------------------------------------
# 4. 打包
# ------------------------------------------------------------------
mkdir -p "${OUT_DIR}"
rm -f "${OUT_DIR:?}/${FPK_NAME}"

if command -v fnpack >/dev/null 2>&1; then
    echo "==> 使用 fnpack 打包"
    ( cd "${BUILD_DIR}" && fnpack build )
    built="$(find "${BUILD_DIR}" -maxdepth 2 -name '*.fpk' | head -n 1)"
    if [ -n "${built}" ]; then
        mv "${built}" "${OUT_DIR}/${FPK_NAME}"
    else
        echo "错误：fnpack 未产出 .fpk 文件" >&2
        exit 1
    fi
else
    # fpk = tar.gz(包根文件 + app.tgz)，app.tgz 是 app/ 目录内容的压缩包，
    # manifest.checksum = md5(app.tgz)。直接把 app/ 目录放进外层会导致
    # 安装器报「解压 app.tgz 失败」。
    echo "==> 生成 app.tgz"
    cp LICENSE "${BUILD_DIR}/app/LICENSE" 2>/dev/null || true
    tar -czf "${BUILD_DIR}/app.tgz" -C "${BUILD_DIR}/app" $(cd "${BUILD_DIR}/app" && ls -A)
    CHECKSUM="$(md5sum "${BUILD_DIR}/app.tgz" | awk '{print $1}')"

    if grep -q '^checksum' "${BUILD_DIR}/manifest"; then
        sed -i "s|^checksum.*|checksum = ${CHECKSUM}|" "${BUILD_DIR}/manifest"
    else
        printf '\nchecksum = %s\n' "${CHECKSUM}" >> "${BUILD_DIR}/manifest"
    fi

    echo "==> 使用 tar.gz 打包（app.tgz 模式，checksum=${CHECKSUM}）"
    rm -rf "${BUILD_DIR}/app" 2>/dev/null || true
    tar -czf "${OUT_DIR}/${FPK_NAME}" -C "${BUILD_DIR}" manifest ICON.PNG ICON_256.PNG app.tgz cmd config wizard
fi

echo "==> 完成：${OUT_DIR}/${FPK_NAME}"
ls -lh "${OUT_DIR}/${FPK_NAME}"
