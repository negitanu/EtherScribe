## Context

EtherScribe は FortiGate / Palo Alto から統一モデル経由でパラメータシートを出す。既存実装は機器概要・IF・静的ルート・ポリシー・オブジェクト・NAT・IPsec・セキュリティプロファイル（FG）・vsys オブジェクト（PA）まで。プリセット `ops` は `ha` / `logging` を含むがモデルが空。WallScribe の `converters/{misc,vpn,network}.py` を契約参照とする。

## Goals / Non-Goals

**Goals:**
- FortiGate: HA、Syslog/SNMP/FortiAnalyzer、SSL-VPN、DHCP、OSPF/OSPF6、BGP、ポリシールートをモデルと JSON/HTML/Excel に反映する。
- Palo Alto: NAT ルールと IPsec トンネルの MVP 抽出。
- 秘匿フィールドはフラグまたは伏せ字のみ。合成 fixture で回帰する。

**Non-Goals:**
- 構築テスト、ネットワーク図、PDF、公式 appid 全件。
- Palo Alto のセキュリティプロファイル詳細、GlobalProtect フル機能。
- HA クラスタ複数ファイル結合解析の高度化。

## Decisions

1. **モデル拡張は WallScribe フィールド名に寄せる**  
   `HASettings`、`LoggingSettings`、`DHCPServer`、`RoutingSettings`（OSPF/BGP）、`SSLVPNSettings`、`PolicyRoute` を `model.Config` に追加。JSON キーは camelCase。

2. **コンバータ分割**  
   FortiGate は `ops.go`（HA/ログ）、`network_extra.go`（DHCP/OSPF/BGP/ポリシールート）、既存 `convert_extra.go` に SSL-VPN 追加。Palo Alto は `parse.go` に NAT/VPN を追記。

3. **出力**  
   `SectionSelected` の `ha` / `logging` / `network` / `vpn` で出し分け。network に DHCP・動的ルート・ポリシールートを含める。

4. **秘匿**  
   HA `password`、SNMP `community` は既存 `redaction.IsSecretField` に寄せるか専用フラグ（`passwordConfigured`）のみ。生値をモデルに載せない方針を優先。

## Risks / Trade-offs

- [動的ルーティングの木構造差] → MVP は主要フィールドのみ、未対応は coverage。
- [hbdev の list/string 両形式] → WallScribe と同様に両方受理。
- [モデル肥大] → 詳細ネストは必要最小限。証明書・SD-WAN 等は本 change 外。

## Migration Plan

後方互換。既存 fixture の主要キーは維持。新フィールドは omitempty。ロールバックは revert。

## Open Questions

なし（本 change 範囲で着手可能）。
