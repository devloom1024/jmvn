#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

usage() {
    echo "用法: ./run.sh <子命令>"
    echo ""
    echo "子命令:"
    echo "  build      编译当前平台"
    echo "  build-all  交叉编译所有平台"
    echo "  test       运行测试"
    echo "  clean      清理构建产物"
    echo "  install    编译并安装到 .env 配置的目录"
    exit 1
}

[ $# -lt 1 ] && usage

CMD="$1"
shift

SCRIPTS="$SCRIPT_DIR/scripts/unix"

case "$CMD" in
    build)      exec "$SCRIPTS/build.sh" "$@" ;;
    build-all)  exec "$SCRIPTS/build-all.sh" "$@" ;;
    test)       exec "$SCRIPTS/test.sh" "$@" ;;
    clean)      exec "$SCRIPTS/clean.sh" "$@" ;;
    install)    exec "$SCRIPTS/install.sh" "$@" ;;
    *)          usage ;;
esac
