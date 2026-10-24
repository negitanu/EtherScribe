# Junos EX（ELS）抽出メモ

EtherScribe が対象とする Juniper EX シリーズ設定の範囲。公式 CLI リファレンス（Configuration Statements）を辞書とし、運用コマンド（show）は入力対象外。

## 前提

- 機種: **EX シリーズ**（QFX / SRX は後続）
- ソフトウェア: **Enhanced Layer 2 Software（ELS）** 相当の VLAN / ethernet-switching 構文
- 入力形式: **brace** および **`show configuration | display set`**
- 正規化: 両形式 → 共通ステートメント列 → EX マッパー

## MVP で読むステートメント

| 領域 | 例 |
| --- | --- |
| 機器 | `system host-name`, `version`, `system time-zone`, `name-server`, `ntp server` |
| ポート | `interfaces <if> … family ethernet-switching`, `description`, `ether-options 802.3ad` |
| IRB | `interfaces irb.<unit> family inet address` / brace の `irb { unit N … }` |
| VLAN | `vlans <name> vlan-id`, `l3-interface` |
| STP | `protocols rstp` / `mstp` / `vstp` |
| 経路 | `routing-options static route` |
| VRF | `routing-instances <name>`（interface / static / protocols） |
| BGP | `protocols bgp`（local-as / neighbor） |

## 非対応（coverage）

- 非 ELS の `ethernet-switching-options` / 旧 vlan 構文
- Virtual Chassis、MC-LAG、EVPN の詳細
- `security {}`（SRX）のポリシー展開
- `delete` 文の適用（注記のみ）

## 合成 fixture

- `core/testdata/junos-ex-els.set`
- `core/testdata/junos-ex-els.conf`
