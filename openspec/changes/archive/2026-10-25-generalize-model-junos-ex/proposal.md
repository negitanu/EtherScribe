## Why

EtherScribe の統一モデルとプリセットはファイアウォール前提のため、ネットワークスイッチ（まず Juniper EX）や将来の Cisco を載せるには L2・VRF・ロール別出力の語彙が足りない。モデルをベンダー横断に一般化してから Junos EX パーサを入れる。

## What Changes

- 統一モデルにデバイス役割（firewall / switch）、仮想コンテキスト（VDOM/vsys/VRF/routing-instance）、VLAN、LAG、L2 ポート属性、VRF スコープ付き L3 ルーティングを **additive** に追加する（既存 FG/PA JSON キーは維持）。
- スイッチ専用プリセット（`switch-overview` / `switch-l2` / `switch-l3` / `switch-ops` / `switch-all`）を新設する。FW プリセットは現状維持。
- Junos EX の設定入力を受け付ける（**brace 形式と `display set` の両方**）。識別不能・SRX 専用ブロックは失敗または coverage。
- Junos 正規化層（brace / set → 共通ステートメント列）の上に EX（ELS 前提）マッパーを実装し、合成 fixture で回帰する。
- QFX / SRX フル対応、運用 show の取り込み、Cisco 実装本体は本 change の対象外（モデルは Cisco 追加を阻害しない形にする）。

## Capabilities

### New Capabilities

- `switch-fabric`: スイッチ向け L2/L3/VRF の統一モデル語彙とスイッチ専用プリセット／セクション。

### Modified Capabilities

- `vendor-parsing`: Junos EX（ELS）の brace / set 解析を SHALL に追加する。
- `config-ingestion`: Junos テキスト入力の受付と機器識別を拡張する。
- `parameter-sheets`: スイッチプリセット選択時の出力範囲を定義する。

## Impact

- `core/internal/model`、`identify`、`ingest`、`export/sections`、JSON/HTML/Excel。
- 新規 `core/internal/parse/junos`（normalize + EX mapper）、`core/testdata` 合成 fixture。
- IPC `sections.list` / `engine.version` の vendors・presets 表示。GUI はプリセット一覧の追随が必要なら最小限。
- 既存 FortiGate / Palo Alto の回帰テストを維持する。
