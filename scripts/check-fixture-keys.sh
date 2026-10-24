#!/usr/bin/env bash
# Regression gate: WallScribe-derived fixtures must expose key JSON fields.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/core"
TMP="$(mktemp)"
go run ./cmd/etherscribe convert -i testdata/fortigate-vdom.conf -o "$TMP" -f json
python3 - <<'PY' "$TMP"
import json,sys
doc=json.load(open(sys.argv[1]))
assert doc.get("deviceInfo",{}).get("hostname")=="FW-VDOM-DEMO"
assert "coverage" in doc
assert doc.get("firewallPolicies")
print("fortigate keys ok")
PY
go run ./cmd/etherscribe convert -i testdata/paloalto-vsys.xml -o "$TMP" -f json
python3 - <<'PY' "$TMP"
import json,sys
doc=json.load(open(sys.argv[1]))
assert doc.get("deviceInfo",{}).get("deviceType")=="PaloAlto"
assert "coverage" in doc or doc.get("interfaces")
print("paloalto keys ok")
PY
go run ./cmd/etherscribe convert -i testdata/junos-ex-els.set -o "$TMP" -f json --preset switch-overview
python3 - <<'PY' "$TMP"
import json,sys
doc=json.load(open(sys.argv[1]))
di=doc.get("deviceInfo",{})
assert di.get("deviceType")=="JunosEX"
assert di.get("role")=="switch"
assert di.get("hostname")=="SW-EX-LAB"
assert doc.get("vlans")
print("junos-ex keys ok")
PY
rm -f "$TMP"
