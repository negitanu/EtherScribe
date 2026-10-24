#!/usr/bin/env bash
# Launch EtherScribe desktop (dev by default, or production build).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MODE="dev"
LAUNCH_APP=0
VITE_PORT="${VITE_PORT:-5173}"
VITE_PID=""

usage() {
  cat <<'EOF'
Usage: ./scripts/run.sh [--prod] [--app]

  (default)  Dev: build engine + Electron main, start Vite, open Electron
  --prod     Build then launch Electron against desktop/dist
  --app      After build (implies --prod), open packed app under desktop/release/

Environment:
  VITE_PORT  Dev server port (default 5173)

EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prod) MODE="prod" ;;
    --app)
      MODE="prod"
      LAUNCH_APP=1
      ;;
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

cleanup() {
  if [[ -n "$VITE_PID" ]] && kill -0 "$VITE_PID" 2>/dev/null; then
    kill "$VITE_PID" 2>/dev/null || true
    wait "$VITE_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

ensure_desktop_deps() {
  cd "$ROOT/desktop"
  if [[ ! -d node_modules ]]; then
    echo "desktop/node_modules missing; running npm install..."
    npm install
  fi
}

wait_for_vite() {
  local url="http://127.0.0.1:${VITE_PORT}"
  local i
  for i in $(seq 1 60); do
    if curl -sf "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.25
  done
  echo "Vite did not become ready at $url" >&2
  return 1
}

find_packed_app() {
  local release="$ROOT/desktop/release"
  if [[ ! -d "$release" ]]; then
    return 1
  fi
  # Prefer macOS .app, then linux/win unpacked binaries from electron-builder --dir
  local app
  app="$(find "$release" -maxdepth 3 -name 'EtherScribe.app' -type d 2>/dev/null | head -n 1 || true)"
  if [[ -n "$app" ]]; then
    printf '%s\n' "$app"
    return 0
  fi
  app="$(find "$release" -maxdepth 3 -type f \( -name 'EtherScribe' -o -name 'etherscribe' -o -name 'EtherScribe.exe' \) 2>/dev/null | head -n 1 || true)"
  if [[ -n "$app" ]]; then
    printf '%s\n' "$app"
    return 0
  fi
  return 1
}

ensure_desktop_deps

# Integrated terminals often export ELECTRON_RUN_AS_NODE=1, which makes
# `electron .` behave like plain Node and breaks require("electron").app.
unset ELECTRON_RUN_AS_NODE

if [[ "$MODE" == "dev" ]]; then
  "$ROOT/scripts/build-engine.sh"
  cd "$ROOT/desktop"
  npm run build:electron
  npx vite --port "$VITE_PORT" --strictPort &
  VITE_PID=$!
  wait_for_vite
  export VITE_DEV_SERVER_URL="http://127.0.0.1:${VITE_PORT}"
  ./node_modules/.bin/electron .
  exit $?
fi

# --prod / --app
if [[ "$LAUNCH_APP" -eq 1 ]]; then
  "$ROOT/scripts/build.sh"
  APP_PATH="$(find_packed_app)" || {
    echo "packed app not found under desktop/release/" >&2
    exit 1
  }
  echo "opening $APP_PATH"
  if [[ "$APP_PATH" == *.app ]]; then
    open "$APP_PATH"
  else
    "$APP_PATH"
  fi
  exit 0
fi

"$ROOT/scripts/build-engine.sh"
cd "$ROOT/desktop"
npm run build:electron
npx vite build
export VITE_DEV_SERVER_URL=""
./node_modules/.bin/electron .
