## ADDED Requirements

### Requirement: Supported input formats
システム SHALL FortiGate のテキスト設定（`.conf`）と Palo Alto の XML 設定（`.xml`）を入力として受け付ける。

#### Scenario: Reject unsupported extension
- **WHEN** 未対応拡張子のファイルを変換対象として渡す
- **THEN** 変換を行わず、利用者向けに形式エラーを返す

### Requirement: Encoding detection
システム SHALL 入力バイト列のエンコーディングを検出し、少なくとも UTF-8、UTF-8 BOM、CP932、Latin-1 を扱えるようにする。

#### Scenario: CP932 FortiGate config
- **WHEN** CP932 で保存された FortiGate 設定を取り込む
- **THEN** 文字化けせず解析可能な文字列へデコードする

### Requirement: Device identification
システム SHALL 内容に基づき FortiGate と Palo Alto を識別し、識別不能な場合は変換を失敗させる。

#### Scenario: Ambiguous content
- **WHEN** どちらの機器設定とも判定できない内容を渡す
- **THEN** 機器種別不明のエラーを返し、部分的な偽のシートを出力しない

### Requirement: Size and content validation
システム SHALL 過大な入力や空入力を拒否し、上限を超える場合は明確なエラーを返す。

#### Scenario: Empty file
- **WHEN** 空ファイルを変換する
- **THEN** 検証エラーとなり成果物を生成しない
