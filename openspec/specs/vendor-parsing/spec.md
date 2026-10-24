# vendor-parsing Specification

## Purpose
FortiGate / Palo Alto / Junos EX 設定を統一中間モデルへ解析し、未対応箇所を coverage として保持する。
## Requirements
### Requirement: Unified intermediate model
システム SHALL ベンダー差分を吸収した統一中間モデルへ解析結果を載せ、出力層がベンダー固有フォーマットに直接依存しないようにする。

#### Scenario: Common device summary fields
- **WHEN** FortiGate または Palo Alto の有効な設定を解析する
- **THEN** ホスト名および OS バージョン相当のフィールドをモデル上で参照できる

### Requirement: FortiGate parsing scope
システム SHALL FortiOS 6.x / 7.x / 8.0 系の `.conf` を対象に、VDOM を含む基本カテゴリ（機器概要、システム、ネットワーク、オブジェクト、ポリシー、NAT、VPN、セキュリティプロファイル）を解析できなければならない。未対応ブロックは解析範囲として保持する。

#### Scenario: VDOM config retains per-VDOM policies
- **WHEN** 複数 VDOM を含む FortiGate 設定を解析する
- **THEN** VDOM 単位でポリシー等をモデルに保持し、未対応箇所を coverage に記録する

### Requirement: Palo Alto parsing scope
システム SHALL PAN-OS 10.x / 11.x 系の `.xml` を対象に、vsys を含む基本カテゴリを解析できなければならない。未対応要素は解析範囲として保持する。

#### Scenario: vsys aware objects
- **WHEN** 複数 vsys を含む Palo Alto 設定を解析する
- **THEN** vsys 単位のオブジェクト／ポリシーをモデルに保持する

### Requirement: Preserve unknowns without inventing data
システム SHALL 解析不能な設定を捏造して埋めてはならず、未対応・警告を coverage として保持しなければならない。

#### Scenario: Unknown block reported
- **WHEN** パーサが未対応の設定ブロックに遭遇する
- **THEN** そのブロックを coverage に含め、対応済みフィールドの値を偽って補完しない

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

### Requirement: Junos EX parsing scope
システム SHALL Juniper EX シリーズ（Enhanced Layer 2 Software 相当の設定）を対象に、brace 形式および `display set` 形式の設定テキストから、機器概要、インターフェース、VLAN、LAG、STP（基本）、IRB／L3、routing-instance（VRF）、静的／主要動的ルートを統一モデルへ解析できなければならない。未対応ブロックは coverage に保持し、捏造してはならない。

#### Scenario: ELS VLAN and interface from set format
- **WHEN** `display set` 形式の EX 設定に VLAN と ethernet-switching ポートが含まれる
- **THEN** モデルに VLAN とポート L2 属性が入り、role が switch 相当になる

#### Scenario: Brace format equivalent
- **WHEN** 同一内容の brace 形式設定を解析する
- **THEN** set 形式と同等の主要フィールド（ホスト名、VLAN、主要 IF）を得られる

### Requirement: Junos statement normalization
システム SHALL brace 形式と set 形式を共通の論理ステートメント列へ正規化してから EX マッピングを行わなければならない。

#### Scenario: Normalize both inputs
- **WHEN** brace と set の各入力を正規化する
- **THEN** パスと値を持つステートメント列が得られ、マッパーが形式差を直接分岐しなくてよい

