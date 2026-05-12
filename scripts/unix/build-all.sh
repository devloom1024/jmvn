#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
cd "$PROJECT_DIR"

VERSION="${VERSION:-1.0.0}"
BIN_DIR="bin"

rm -rf "$BIN_DIR"
mkdir -p "$BIN_DIR"

echo "构建所有平台 jmvn ${VERSION} ..."

GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$BIN_DIR/jmvn-windows-amd64.exe" .
echo "  → $BIN_DIR/jmvn-windows-amd64.exe"

GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$BIN_DIR/jmvn-linux-amd64" .
echo "  → $BIN_DIR/jmvn-linux-amd64"

GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$BIN_DIR/jmvn-darwin-amd64" .
echo "  → $BIN_DIR/jmvn-darwin-amd64"

GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=${VERSION}" -o "$BIN_DIR/jmvn-darwin-arm64" .
echo "  → $BIN_DIR/jmvn-darwin-arm64"

echo "完成: 所有平台构建完毕"
