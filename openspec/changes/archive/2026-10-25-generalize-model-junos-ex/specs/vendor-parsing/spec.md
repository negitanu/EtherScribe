## ADDED Requirements

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
