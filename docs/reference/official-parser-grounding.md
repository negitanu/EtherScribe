# 公式リファレンスに基づくネットワーク解析

確認日: 2026-10-25。公式構文の意味を対応するコード・独立した合成 fixture に結び付けた記録。公式資料の全文は転載しない。Batfish 原典は文法の追加参照であり、正解モデルを自動生成する oracle にはしない。

## Junos EX / ELS

1. [peer-as](https://www.juniper.net/documentation/us/en/software/junos/cli-reference/topics/ref/statement/peer-as-edit-protocols-bgp.html) と [autonomous-system](https://www.juniper.net/documentation/us/en/software/junos/cli-reference/topics/ref/statement/autonomous-system-edit-routing-options.html)
   - 自 AS と対向 AS を分離する。`routing-options autonomous-system` を自 AS とし、BGP の global/group/neighbor の設定を階層で解決する。routing-instance の明示 AS を優先する。
   - `local-as` はセッションのローカル AS として別に保持する。group の peer-as を自 AS に転用しない。
   - 実装: `core/internal/parse/junos/routing.go`。回帰: `grounded-junos-network.set` の WAN/IBGP と順序入れ替えテスト。
2. [local-as](https://www.juniper.net/documentation/us/en/software/junos/cli-reference/topics/ref/statement/local-as-edit-protocols-bgp.html)、[local-address](https://www.juniper.net/documentation/us/en/software/junos/cli-reference/topics/ref/statement/local-address-edit-protocols-bgp.html)、[import](https://www.juniper.net/documentation/us/en/software/junos/cli-reference/topics/ref/statement/import-edit-protocols-bgp.html)
   - ローカル AS・送信元アドレス・受信/送信ポリシーを peer ごとに継承して出力する。ポリシー名リストは評価順を維持する。ポリシー自体の評価や経路シミュレーションはしない。
3. [Static Route Preferences and Qualified Next Hops](https://www.juniper.net/documentation/us/en/software/junos/static-routing/topics/topic-map/static-route-prefer-qualified-next-hop.html) と [qualified-next-hop](https://www.juniper.net/documentation/us/en/software/junos/cli-reference/topics/ref/statement/qualified-next-hop-edit-routing-options.html)
   - ルートの複数の設定行を統合し、next-hop ごとに 1 レコードを出す。qualified-next-hop の preference/metric は親属性より優先する。
   - `routing-options rib inet6.0 static route` も対象。実装: `routing.go`。回帰: `grounded-junos-network.set`、`grounded-junos-compact.conf/.set`。
4. [Aggregated Ethernet LACP for Switches](https://www.juniper.net/documentation/us/en/software/junos/interfaces-ethernet-switches/topics/topic-map/aggregated-ethernet-lacp-switches.html)
   - LAG 所属と LACP 設定を区別する。`802.3ad` メンバーだけでは LACP を有効としない。`aggregated-ether-options lacp` を確認する。
   - 回帰: 同じ fixture の ae0（LACP 無指定）と ae1（passive）。
5. [deactivate](https://www.juniper.net/documentation/us/en/software/junos/cli-reference/topics/ref/command/deactivate.html)、[CLI Configuration Mode Overview](https://www.juniper.net/documentation/us/en/software/junos/cli/topics/topic-map/cli-configuration.html)
   - `inactive:` と display set の `deactivate` を同じ状態として扱う。brace は改行でなく引用符外の区切りで正規化し、コメントや引用符内の区切り文字を保護する。
   - 実装: `normalize.go`、`routing.go`。回帰: compact brace/set 同値と不正入力診断。

## FortiOS 7.4.0

正本: [FortiOS 7.4.0 CLI Reference PDF](https://fortinetweb.s3.amazonaws.com/docs.fortinet.com/v2/attachments/9193cdf6-ef51-11ed-8e6d-fa163e15d75b/FortiOS-7.4.0-CLI_Reference.pdf)。HTML リファレンスの取得が 403 となるため、公式配布 PDF の該当箇所を確認した。

- `config system interface`、secondaryip、ipv6/ip6-extra-addr（印刷ページ 1230–1286 付近）: primary/secondary IPv4 と IPv6 の設定アドレスを保持する。IPv4 の secondary は enable 時のみ対象。MTU override、VRF、aggregate member を抽出する。
- 同節の lacp-mode（1251–1252 付近）: static と active/passive を区別し、aggregate で省略された場合は公式既定の active として扱う。
- `config router static` / `static6`（878–882 付近）: distance、priority、comment、blackhole、vrf を別項目として保持する。dstaddr と省略された既定宛先を混同しない。
- 実装: `core/internal/parse/fortigate/parse.go`、`interfaces_extra.go`。回帰: `grounded-fortios-network.conf`。

## PAN-OS 11.1 XML

正本: [PAN-OS 11.1 Configure CLI Command Hierarchy](https://docs.paloaltonetworks.com/ngfw/pan-os-cli-quick-start/cli-command-hierarchy/pan-os-11-1-configure-cli-command-hierarchy)。CLI の階層を XML の対応ノードへマップする。CLI set 入力自体の対応を意味しない。

- L3 IF と論理 unit の IPv6 address、enabled、MTU を抽出する。アドレスは設定値であり、実際のリンク状態や経路到達性を示さない。
- virtual-router の interface リストを vsys import とは別に保持する。
- static route の distance、metric、IPv6 next-hop、discard/next-vr を別の意味として扱う。
- 実装: `core/internal/parse/paloalto/parse.go`、`network.go`、`coverage.go`。回帰: `grounded-paloalto-network.xml` と既存 `paloalto-network-ops.xml`。

## 出力契約

既存の JSON キーを維持する加算的な変更。HTML/Excel は同じモデルから列を作る。

- IF: `ipAddress` は先頭アドレス、`addresses` は設定アドレス全件。`ipv6Enabled` は明示された値だけ出力し、省略と false を区別する。`mtu`、`routingContext` を追加。
- IF の `context` / `vdom` は従来どおり VDOM/vsys または Junos routing-instance。`routingContext` は別に FortiOS VRF / PAN-OS virtual-router / Junos routing-instance を保持する。
- Route: `priority`、`action`、`description`、`destinationObject`、`routingContext` を追加。`distance` は Junos の preference も格納し、レポートの列名は「管理距離 / Preference」とする。metric と混同しない。
- BGP peer: `group`、`localAs`、`importPolicies`、`exportPolicies`。継承元そのものの追跡は行わず、解決済みの値を出力する。
- レポート回帰: `core/internal/pipeline/grounded_network_test.go` で JSON 値、HTML 値、Excel XML の妥当性と値の保持を検証。

## 対応限界

Cisco IOS/NX-OS パーサ、PAN-OS set/Advanced Routing、Junos apply-groups 展開・論理システム・任意 RIB・ポリシー評価は未対応。Junos delete は適用せず coverage に示す。認証情報や未知の属性を汎用の値付きダンプとして出力しない。実機・実設定・外部エンジンとの動作同等性は検証していない。

## PAN-OS DHCP / Syslog とメーカー別サービス出力（2026-10-25）

出典: [PAN-OS 11.1 Configure CLI hierarchy](https://docs.paloaltonetworks.com/ngfw/pan-os-cli-quick-start/cli-command-hierarchy/pan-os-11-1-configure-cli-command-hierarchy)、[DHCP Server 設定](https://docs.paloaltonetworks.com/ngfw/networking/dhcp/configure-an-interface-as-a-dhcp-server)、[Syslog Server Profiles](https://docs.paloaltonetworks.com/ngfw/help/11-1/device/device-server-profiles-syslog)、[Configure Syslog Monitoring](https://docs.paloaltonetworks.com/pan-os/11-1/pan-os-admin/monitoring/use-syslog-for-monitoring/configure-syslog-monitoring)。

- DHCP: `network/dhcp/interface/entry/server` を client と分離する。auto / enabled / disabled、複数 IP プール（範囲または CIDR）、予約 IP/MAC、マスク、ゲートウェイ、DNS、probe-ip、lease の timeout / unlimited を保持する。DNS の inherited は解決せずそのまま表示する。リース timeout は設定値を保持し単位変換しない。MAC なしの予約は動的割当の除外 IP としても保持する。IPv4 relay は DHCP リレー表に分ける。
- Syslog: shared と vsys に属するプロファイル・サーバー名・送信先・ポート・transport・format・facility を出力する。プロファイルが存在するだけでは転送が有効とは限らないため、状態は「設定あり」とし、ログ転送プロファイルとの関連付けは今回解析しない。
- `logging.syslogConfigured` は送信先定義があれば true、解析済みで未定義なら false、解析未対応なら未出力。HTML/Excel は設定あり・設定なし・未解析を区別する。
- FortiGate は syslogd〜syslogd4 の setting/override-setting を解析する。無効設定の送信先も明示する。
- FortiAnalyzer は FortiGate の出力に限定する。現在の HA モデルは FortiOS 専用なので、Palo Alto/Junos では空の FortiOS 型 HA 表を出さない。メーカー判定は device_info 節の選択有無によらない。
- 実装: `core/internal/parse/paloalto/services.go`、`core/internal/export/vendor_services.go`。合成回帰: `core/testdata/paloalto-dhcp-syslog.xml`、`core/testdata/fortigate-service-report.conf`、`core/internal/pipeline/vendor_services_test.go`。
- DHCP の追加オプション、IPv6 relay、Syslog カスタム format は構造のみ coverage に記録する。
