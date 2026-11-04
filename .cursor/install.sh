#!/usr/bin/env bash
# Idempotent dependency install for EtherScribe on the Cursor cloud Linux VM.
# Go modules are vendored (core/third_party); desktop deps come from the npm lockfile.
# Re-running this script must succeed without rewriting lockfiles.
set -euo pipefail

SOURCE="${BASH_SOURCE[0]:-}"
if [[ -f "$SOURCE" && "$(basename "$SOURCE")" == "install.sh" ]]; then
  ROOT="$(cd "$(dirname "$SOURCE")/.." && pwd)"
else
  ROOT="$(pwd)"
fi
cd "$ROOT"

GO_VERSION="1.27.3"
GO_SHA256="65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40"
# Matches the OpenSpec skills committed in this repo (generatedBy 1.3.1).
OPENSPEC_VERSION="1.3.1"

if [[ "$(uname -s)" != "Linux" || "$(uname -m)" != "x86_64" ]]; then
  echo "EtherScribe cloud install supports Linux x86_64 only." >&2
  exit 1
fi

if ! command -v node >/dev/null 2>&1 || ! command -v npm >/dev/null 2>&1; then
  echo "Node.js 22 and npm are required (GitHub Actions uses Node 22)." >&2
  exit 1
fi
node_major="$(node -p 'process.versions.node.split(".")[0]')"
if [[ "$node_major" -lt 22 ]]; then
  echo "Node.js 22+ is required, found $(node -v)." >&2
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for scripts/check-fixture-keys.sh and scripts/audit-release.py." >&2
  exit 1
fi

if ! /usr/local/go/bin/go version 2>/dev/null | grep -q "go${GO_VERSION} "; then
  tmp="$(mktemp -d)"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o "$tmp/go.tgz"
  echo "${GO_SHA256}  $tmp/go.tgz" | sha256sum -c -
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "$tmp/go.tgz"
  rm -rf "$tmp"
fi
sudo ln -sfn /usr/local/go/bin/go /usr/local/bin/go
sudo ln -sfn /usr/local/go/bin/gofmt /usr/local/bin/gofmt
export PATH="/usr/local/bin:${PATH}"
hash -r

# `node` on the cloud image is /exec-daemon/node, so npm's default global
# prefix is not writable. Install the CLI under the user prefix and symlink it.
openspec_prefix="${HOME}/.local"
openspec_version="$(${openspec_prefix}/bin/openspec --version 2>/dev/null || true)"
if [[ "$openspec_version" != *"${OPENSPEC_VERSION}"* ]]; then
  mkdir -p "$openspec_prefix"
  npm install -g --prefix "$openspec_prefix" "@fission-ai/openspec@${OPENSPEC_VERSION}"
fi
sudo ln -sfn "${openspec_prefix}/bin/openspec" /usr/local/bin/openspec

npm ci --no-audit --no-fund --prefix "$ROOT/desktop"
