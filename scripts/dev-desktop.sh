#!/usr/bin/env bash
# Compatibility wrapper — prefer ./scripts/run.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
exec "$ROOT/scripts/run.sh" --prod "$@"
