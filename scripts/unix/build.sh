#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
cd "$PROJECT_DIR"

VERSION="${VERSION:-1.0.0}"

case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*)
        OUTPUT="bin/jmvn.exe"
        ;;
    *)
        OUTPUT="bin/jmvn"
        ;;
esac

echo "构建 jmvn ${VERSION} ..."
go build -ldflags "-s -w -X main.version=${VERSION}" -o "$OUTPUT" .
echo "完成: $OUTPUT"
