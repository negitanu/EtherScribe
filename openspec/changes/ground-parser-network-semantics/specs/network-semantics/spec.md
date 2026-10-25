## ADDED Requirements

### Requirement: Reference grounded interpretation
実装 SHALL 公式 CLI の出典と対応する合成回帰テストを記録し、対応する構文の意味を保持する。
#### Scenario: Junos BGP inheritance
- **WHEN** 自 AS、グループ peer-as、neighbor 固有設定が異なる順序で存在する
- **THEN** 自 AS と対向 AS を区別し、継承済み peer を重複なく出力する
#### Scenario: Static route properties
- **WHEN** 同一ルートの next-hop、preference、metric、description が別行で存在する
- **THEN** ルートを重複生成せず、qualified-next-hop の優先属性も保持する
#### Scenario: Aggregation without LACP
- **WHEN** LAG メンバーが定義され LACP は定義されない
- **THEN** LACP 有効として出力しない

### Requirement: Rich network output
システム SHALL 複数 IPv4/IPv6 アドレス、明示された経路属性、仮想ルーティング所属を JSON/HTML/Excel に保持する。
#### Scenario: Secondary and IPv6 addresses
- **WHEN** 対応する IF に secondary IPv4 と IPv6 が設定される
- **THEN** 全アドレスを保持し、従来の先頭アドレス項目も維持する
#### Scenario: Virtual router membership
- **WHEN** PAN-OS の IF が vsys と virtual-router に所属する
- **THEN** 両者を混同せず別の項目として出力する
#### Scenario: Disabled Junos hierarchy
- **WHEN** BGP group またはルートの祖先が deactivate される
- **THEN** 有効な設定として出力しない

### Requirement: Equivalent Junos input forms
システム SHALL brace と display set で同じ意味の対応構文を同じモデルへ正規化する。
#### Scenario: Compact inactive configuration
- **WHEN** コメントや引用符内の区切り文字、inactive: を含む一行の brace 設定が入力される
- **THEN** 階層と値を壊さず、display set の deactivate と同じ状態へ変換する
