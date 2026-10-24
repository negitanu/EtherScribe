## ADDED Requirements

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
