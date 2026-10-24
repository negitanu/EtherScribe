## Context

EtherScribe は FortiGate / Palo Alto の設定からパラメータシートを出す。統一モデル（`model.Config`）とプリセットは FW 語彙が中心。ユーザー決定により Juniper **EX シリーズ**を最初のスイッチ対象とし、入力は **brace と display set の両方**、**L3 と VRF（routing-instance）込み**、**スイッチ専用プリセット**、将来 **Cisco** を見据えて **モデル一般化をパーサより先**に行う。

参照: Junos CLI Configuration Statements / Operational Commands（運用 show は入力対象外）。ELS 前提。

## Goals / Non-Goals

**Goals:**
- ロール横断の語彙（VirtualContext / VLAN / LAG / L2 ポート / VRF 付き Routing）を additive にモデルへ入れる。
- スイッチ専用プリセットを新設し、FW プリセットを壊さない。
- Junos EX を brace / set から正規化し、統一モデルへ載せる。
- FG/PA 既存 fixture・テストが通る。

**Non-Goals:**
- QFX / SRX の本格対応、Virtual Chassis / MC-LAG の詳細、EVPN フル。
- 運用コマンド（show）出力の取り込み、実機書き込み。
- Cisco パーサ本体（モデルが阻害しないことのみ保証）。
- 既存 JSON キーの破壊的リネーム。

## Decisions

### 1. Additive 一般化（破壊的リネームしない）

- `DeviceInfo` に `Role`（`firewall` | `switch`）と任意の `VendorFamily` を追加。
- `VirtualContexts []` を追加。type: `vdom` | `vsys` | `routing-instance` | `vrf` | `default`。
- 既存の `VDOMList` / 各要素の `vdom` フィールドは維持。Junos/Cisco 新規は `context`（または両書き）で載せる。
- 新規: `VLANs`, `LAGs`, Interface の L2 属性（`switchportMode`, `accessVlan`, `trunkVlans`, `lagParent` 等）。
- `Routing` / `Route` / OSPF / BGP に `Context` を必須相当（空＝`default`）で付与。
- IRB/SVI は `Interface` の type（`irb` / `svi` / `l3`）とアドレスで表現。

**代替不採用:** FW モデルと Switch モデルの完全分離（出力・差分・IPC が二重化するため）。

### 2. スイッチ専用プリセット

| Key | Sections（概略） |
| --- | --- |
| `switch-overview` | device_info, network, vlans |
| `switch-l2` | + lag, stp |
| `switch-l3` | + routing（VRF 付き）, dhcp |
| `switch-ops` | + system_settings, logging, ha |
| `switch-all` | スイッチ向けセクション和（policies/nat/vpn/security は含めない） |

SectionKeys に `vlans`, `lag`, `stp` を追加。`network` は IF +（既存）ルートのまま、VLAN/LAG は明示セクションでも `network` に含めてもよいが、プリセット制御のため **独立キー**とする。

### 3. Junos 正規化層 → EX マッパー

```
input text → detect brace|set
           → normalize to []Statement{Path []string, Value, Op set|deactivate|delete}
           → map EX ELS → model.Config
```

- set: 行単位。`deactivate` は Enabled=false 相当、`delete` は coverage または無視（MVP: coverage 注記）。
- brace: 階層を path に flatten。
- 識別: `set system host-name` / `system {`+`host-name`、`family ethernet-switching` / `vlans`、機種ヒント `EX\d+`。Role=switch, DeviceType=JunosEX（または Junos + Role）。
- `security {` のみ目立つ SRX 系は識別失敗または「EX 対象外」エラー（設計: **識別は Junos だが Role 推定に ethernet-switching/vlans を要求**し、無ければ coverage 多めでも switch として通すか、明示エラーにする → **ELS L2 痕跡が無い純 L3 は coverage 付きで受理**し、security ツリーは未展開 coverage）。

### 4. VRF / routing-instance

- `routing-instances <name>` → VirtualContext。
- インスタンス内の interface / routing-options / protocols は当該 Context を付与。
- グローバルは Context=`default`。

### 5. 実装順序（タスクに反映）

1. モデル + セクション/プリセット + 既存パーサの回帰（空フィールド）。
2. identify / ingest。
3. junos normalize + 単体テスト。
4. EX mapper（system, interfaces, vlans, ae, stp, irb, RI/L3）。
5. export / CLI / 文書 / fixture。

### 6. IPC

- `engine.version.vendors` に Junos EX を追加。
- `sections.list` は全 presets を返す（role フィルタは任意。GUI はキー接頭辞 `switch-` で見分け可能）。MVP で params 拡張は必須としない。

## Risks / Trade-offs

- [ELS と非 ELS の混在] → ELS のみ正式対応。旧 vlan 構文は coverage。
- [brace パーサの引用符・継続行] → 合成 fixture と段階的テスト。失敗は parseErrors。
- [モデル肥大] → スイッチセクションは omitempty。FW 出力は既存プリセットで非表示。
- [DeviceType 命名] → `JunosEX` を追加し、将来 `CiscoIOS` 等と並列。

## Migration Plan

- 既存クライアントは未知 JSON フィールドを無視できる前提。schemaVersion は 1 のまま（セクション追加のみ）または 2 に上げる場合は generator に明記。
- ロールバックは Git revert。ユーザー保存ファイルはアプリ管理外。

## Open Questions

なし（探索で決定済み）。実装中に SRX 誤検出が出たら識別ヒューリスティックを tasks で調整。
