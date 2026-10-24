#!/usr/bin/env bash
# Build EtherScribe desktop (engine + renderer + Electron, optionally packed).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PACK=1

usage() {
  cat <<'EOF'
Usage: ./scripts/build.sh [--no-pack]

  (default)  Build and run electron-builder --dir → desktop/release/
  --no-pack  Build engine + Vite + Electron only (no packager)

EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-pack) PACK=0 ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
  shift
done

ensure_desktop_deps() {
  cd "$ROOT/desktop"
  if [[ ! -d node_modules ]]; then
    echo "desktop/node_modules missing; running npm install..."
    npm install
  fi
}

ensure_desktop_deps
cd "$ROOT/desktop"
if [[ "$PACK" -eq 1 ]]; then
  npm run pack
  echo "packed → $ROOT/desktop/release/"
else
  npm run build
  echo "built (no pack) → engine/resources, desktop/dist, desktop/dist-electron"
fi
