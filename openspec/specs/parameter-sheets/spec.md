# parameter-sheets Specification

## Purpose
統一モデルから HTML / Excel / JSON のパラメータシートを生成し、プリセットと秘匿・coverage を制御する。
## Requirements
### Requirement: Multi-format export
システム SHALL 統一モデルから HTML、Excel（`.xlsx`）、JSON のパラメータシートを生成できなければならない。

#### Scenario: Export same model to three formats
- **WHEN** 同一の解析済みモデルに対し html / excel / json を指定する
- **THEN** 各形式の成果物が生成され、機器概要など共通に定義された項目を含む

### Requirement: Presets and section selection
システム SHALL 用途別プリセットとセクション個別指定で出力範囲を制御できなければならない。

#### Scenario: Policy preset
- **WHEN** プリセット `policy`（または同等）で JSON 出力する
- **THEN** ポリシー監査に必要なセクションが含まれ、明示的に除外されたセクションは出力されない

### Requirement: Redaction of secrets
システム SHALL パスワード、`ENC` 文字列、SNMP コミュニティなど秘匿値を HTML / Excel / JSON に出力してはならない。

#### Scenario: Encrypted secret not emitted
- **WHEN** 設定に `ENC` 形式の秘密が含まれ解析される
- **THEN** いずれの出力形式にもその生値を含めない

### Requirement: Coverage appendix
システム SHALL 台帳に反映できなかった設定を付録「解析範囲」（JSON では `coverage`）として成果物に含めなければならない。

#### Scenario: Coverage in JSON
- **WHEN** 未対応ブロックを含む設定を JSON 出力する
- **THEN** `coverage`（または同等）に未反映箇所の情報が含まれる

### Requirement: HA and logging sections in exports
システム SHALL プリセットまたはセクション指定で `ha` および `logging` を選んだとき、解析済みの HA／ログ設定を JSON・HTML・Excel に含めなければならない。

#### Scenario: Ops preset includes HA
- **WHEN** プリセット `ops` で JSON 出力し HA がモデルにある
- **THEN** 成果物に HA 情報が含まれる

### Requirement: Network and VPN export include extended fields
システム SHALL `network` セクションに DHCP・動的ルーティング・ポリシールートを、`vpn` セクションに SSL-VPN を含められるようにしなければならない。

#### Scenario: Network section with DHCP
- **WHEN** `network` セクション指定で JSON 出力し DHCP がモデルにある
- **THEN** 成果物に DHCP サーバー情報が含まれる

### Requirement: Switch preset export
システム SHALL スイッチ専用プリセットまたはスイッチ向けセクション指定で出力するとき、VLAN／LAG／L2 ポート／VRF 付きルーティングなど解析済みのスイッチ関連項目を JSON・HTML・Excel に含めなければならない。ファイアウォール専用プリセットの既存挙動を壊してはならない。

#### Scenario: switch-l3 includes VRF routes
- **WHEN** プリセット `switch-l3` で JSON 出力し VRF 付きルートがモデルにある
- **THEN** 成果物にコンテキスト付きルーティング情報が含まれる

#### Scenario: Legacy policy preset unchanged
- **WHEN** プリセット `policy` で FortiGate モデルを JSON 出力する
- **THEN** 従来どおりポリシー監査向けセクションが含まれ、スイッチ専用セクションの強制追加はない

