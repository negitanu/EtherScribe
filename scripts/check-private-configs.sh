#!/usr/bin/env bash
# Opt-in local diagnostics: aggregate output only, no generated reports.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
if [[ "${1:-}" == "--help" ]]; then
  echo 'Usage: ./scripts/check-private-configs.sh [private-input-directory]'
  echo 'Defaults to .data/ or data/. Only anonymous counts and validation status are printed.'
  exit 0
fi
PRIVATE_INPUT_DIR="${1:-${ETHERSCRIBE_PRIVATE_DATA_DIR:-}}"
if [[ -z "$PRIVATE_INPUT_DIR" ]]; then
  if [[ -d "$ROOT/.data" ]]; then PRIVATE_INPUT_DIR="$ROOT/.data"
  elif [[ -d "$ROOT/data" ]]; then PRIVATE_INPUT_DIR="$ROOT/data"
  else echo 'Private input directory is unavailable.' >&2; exit 2
  fi
fi
if [[ ! -d "$PRIVATE_INPUT_DIR" ]]; then echo 'Private input directory is unavailable.' >&2; exit 2; fi
PRIVATE_INPUT_DIR="$(cd "$PRIVATE_INPUT_DIR" && pwd)"
cd "$ROOT/core"
ETHERSCRIBE_PRIVATE_DATA_DIR="$PRIVATE_INPUT_DIR" go test ./internal/pipeline -run '^TestPrivateConfigurations$' -count=1 -v
