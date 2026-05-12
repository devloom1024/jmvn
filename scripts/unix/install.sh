#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
cd "$PROJECT_DIR"

if [ ! -f .env ]; then
    echo "请先复制 .env.example 为 .env 并配置 INSTALL_DIR"
    exit 1
fi

set -a
source .env
set +a

if [ -z "${INSTALL_DIR:-}" ]; then
    echo "错误: .env 中未配置 INSTALL_DIR"
    exit 1
fi

"$SCRIPT_DIR/build.sh"

cp bin/jmvn "$INSTALL_DIR/"
echo "已安装到 $INSTALL_DIR"
