## ADDED Requirements
### Requirement: Real configuration structural regression
システム SHALL 実設定の構造比較で確認した解析漏れを、独立に作成した合成 fixture で再現・検証して改善する。
#### Scenario: Source and parsed inventories
- **WHEN** 対応する設定コレクションの件数を検証する
- **THEN** その件数とモデルの対応を確認でき、実値を含むスナップショットを必要としない

### Requirement: FortiGate export boundaries and policy order
システム SHALL VDOM export の暗黙の edit 終端を処理し、ポリシーをソースの edit 順に保持する。非 VDOM の不正終端を隠してはならない。
#### Scenario: Implicit VDOM next
- **WHEN** VDOM の edit が next を省略して end で閉じられる
- **THEN** 内部ブロックと外側ブロックを取り違えず、解析警告なしで後続 VDOM も解析する
#### Scenario: Ordered policies
- **WHEN** ポリシーの ID が数値順でない順序で定義される
- **THEN** モデルと出力はその評価順を保持する
### Requirement: Palo Alto interface and operational fields
システム SHALL 物理・集約・VLAN・サブ IF と動作モード、管理 IP/DNS/NTP、ポリシーのアプリケーション・ユーザー・カテゴリ・プロファイルを保持する。
#### Scenario: Layer2 aggregate and VLAN units
- **WHEN** L2 物理ポートと aggregate-ethernet、vlan/units が存在する
- **THEN** 全インターフェースがモデルへ入り、LAG メンバーと VLAN の L3 関連が保持される
#### Scenario: Application based security rule
- **WHEN** application と profile-setting を持つセキュリティルールがある
- **THEN** サービスと混同せず、それぞれ独立した項目に保持する

### Requirement: Honest structural coverage
システム SHALL 未モデル化の構造を解析範囲に示し、FortiGate の VDOM 内と Palo Alto の部分対応項目も対象にする。属性・本文・秘密を付録へ転記してはならない。
#### Scenario: Unmodeled VDOM or XML option
- **WHEN** VDOM 内に未対応ブロックがある、または Palo Alto に未モデル化オプションがある
- **THEN** 値と識別子を含まない構造パス・出現数を解析範囲に記録する
