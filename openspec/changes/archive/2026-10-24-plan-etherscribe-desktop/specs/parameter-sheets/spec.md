## ADDED Requirements

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
