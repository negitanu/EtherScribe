## Why

MVP 骨格は揃ったが、運用系（HA・ログ）とネットワーク系（DHCP・OSPF/BGP・ポリシールート）および SSL-VPN はプリセットに枠がある一方、モデルへ未展開のままである。WallScribe 仕様の基本カテゴリを埋め、パラメータシートの実用密度を上げる。

## What Changes

- FortiGate: HA（`system ha`）、Syslog / SNMP / FortiAnalyzer、SSL-VPN、DHCP、OSPF/OSPF6、BGP、ポリシールートを統一モデルへ載せる。
- Palo Alto: vsys 配下の NAT ルールと IPsec トンネルの最小解析を追加する。
- 既存プリセット `ha` / `logging` / `network` / `vpn` のエクスポートを上記データで埋める。
- 秘匿（HA パスワード、SNMP コミュニティ等）は生値を出力しない。
- 合成 fixture と Go テストで回帰を担保する。

## Capabilities

### New Capabilities

なし（既存 capability の解析範囲拡張）。

### Modified Capabilities

- `vendor-parsing`: FortiGate の HA・ログ・VPN(SSL)・DHCP・動的ルーティング、Palo Alto の NAT/VPN を SHALL に含める。
- `parameter-sheets`: `ha` / `logging` セクションおよび拡張 `network` / `vpn` の出力要件を明文化する。

## Impact

- `core/internal/model` に HA / Logging / DHCP / Routing / SSLVPN フィールド追加。
- `core/internal/parse/fortigate`・`paloalto` のコンバータ追加。
- `core/internal/export`（JSON/HTML/Excel）と `core/testdata`、`docs/reference/wallscribe-essence.md`。
- CLI / Electron IPC 契約のメソッド追加は不要（既存 convert がモデルを透過利用）。
