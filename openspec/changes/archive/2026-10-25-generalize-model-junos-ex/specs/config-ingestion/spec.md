## MODIFIED Requirements

### Requirement: Supported input formats
システム SHALL FortiGate のテキスト設定（`.conf`）、Palo Alto の XML 設定（`.xml`）、および Junos のテキスト設定（brace または `display set`。拡張子は少なくとも `.conf` / `.txt` / `.set`）を入力として受け付ける。

#### Scenario: Reject unsupported extension
- **WHEN** 未対応拡張子のファイルを変換対象として渡す
- **THEN** 変換を行わず、利用者向けに形式エラーを返す

#### Scenario: Accept Junos set file
- **WHEN** `.set` または `.txt` の Junos `display set` 設定を渡す
- **THEN** 拡張子理由では拒否せず、内容識別へ進む

### Requirement: Device identification
システム SHALL 内容に基づき FortiGate、Palo Alto、および Junos EX（スイッチ）を識別し、識別不能な場合は変換を失敗させる。

#### Scenario: Ambiguous content
- **WHEN** 既知の機器設定とも判定できない内容を渡す
- **THEN** 機器種別不明のエラーを返し、部分的な偽のシートを出力しない

#### Scenario: Identify Junos EX set config
- **WHEN** `set system host-name` と ethernet-switching または vlans を含む設定を渡す
- **THEN** Junos EX（switch）として識別される
