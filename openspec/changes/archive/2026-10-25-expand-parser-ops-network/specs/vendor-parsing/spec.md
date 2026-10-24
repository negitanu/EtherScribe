## ADDED Requirements

### Requirement: FortiGate HA parsing
システム SHALL FortiGate の `system ha` から HA モード、グループ、優先度、監視／ハートビート IF、管理 IF 有無を統一モデルに載せなければならない。パスワード生値は保持してはならない。

#### Scenario: Active-passive HA extracted
- **WHEN** `mode a-p` と `hbdev` / `monitor` を含む設定を解析する
- **THEN** モデルの HA に Active-Passive 相当のモードとハートビート IF が入り、パスワード生値が含まれない

### Requirement: FortiGate logging and monitoring parsing
システム SHALL Syslog（`log syslogd setting`）、SNMP（コミュニティまたは v3 ユーザー）、FortiAnalyzer 設定をログ／監視セクションとしてモデルに載せなければならない。SNMP コミュニティ生値は出力用モデルに載せてはならない。

#### Scenario: Syslog and SNMP presence
- **WHEN** Syslog 有効と SNMP コミュニティ定義を含む設定を解析する
- **THEN** ログ設定にサーバーと SNMP エントリ（コミュニティは伏せ字または未格納）が含まれる

### Requirement: FortiGate SSL-VPN parsing
システム SHALL `vpn ssl settings` およびポータル設定から SSL-VPN の待受／プール／モードをモデルの VPN に含めなければならない。

#### Scenario: SSL-VPN settings extracted
- **WHEN** SSL-VPN 設定ブロックを含む FortiGate 設定を解析する
- **THEN** VPN セクションに SSL-VPN エントリが含まれる

### Requirement: FortiGate DHCP and dynamic routing parsing
システム SHALL DHCP サーバー、OSPF／OSPF6、BGP、ポリシールートをネットワーク関連としてモデルに載せなければならない。未対応サブブロックは coverage に残してよい。

#### Scenario: DHCP and OSPF present
- **WHEN** `system dhcp server` と `router ospf` を含む設定を解析する
- **THEN** DHCP サーバーと OSPF（router-id 等）がモデルに入る

### Requirement: Palo Alto NAT and IPsec MVP
システム SHALL Palo Alto 設定から NAT ルールおよび IPsec トンネルの主要項目を統一モデルへ載せなければならない。未対応要素は coverage に記録する。

#### Scenario: NAT rule detected
- **WHEN** vsys の NAT ルールを含む Palo Alto XML を解析する
- **THEN** `natPolicies` にルール名と変換先相当のフィールドが含まれる
