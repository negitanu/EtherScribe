#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="$ROOT/desktop/resources/engine"
mkdir -p "$OUT_DIR"
cd "$ROOT/core"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
BIN="etherscribe"
if [[ "$GOOS" == "windows" ]]; then BIN="etherscribe.exe"; fi
GOPROXY="${GOPROXY:-off}" CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags='-s -w' -o "$OUT_DIR/$BIN" ./cmd/etherscribe
echo "built $OUT_DIR/$BIN"
