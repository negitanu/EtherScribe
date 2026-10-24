# engine-ipc Specification

## Purpose
TBD - created by archiving change plan-etherscribe-desktop. Update Purpose after archive.
## Requirements
### Requirement: Standalone CLI engine
システム SHALL Go 製エンジンを Electron なしで CLI 実行でき、少なくとも単体変換と差分をファイル入出力で完了できるようにする。

#### Scenario: CLI convert to JSON
- **WHEN** 利用者が CLI で対応設定ファイルと JSON 出力先を指定して実行する
- **THEN** 期待する JSON パラメータシートが出力され、終了コード 0 を返す

### Requirement: JSON-RPC over stdio
システム SHALL Electron Main とエンジンの間で stdio 上の JSON-RPC 2.0（NDJSON）を話し、少なくとも version・convert・compare・sections.list・cancel を提供する。

#### Scenario: Version handshake
- **WHEN** Main が `engine.version` を呼び出す
- **THEN** エンジン版と対応能力（形式・ベンダー）を返し、未対応メソッドは JSON-RPC エラーで示す

### Requirement: Progress and cancellation
システム SHALL 長時間ジョブで進捗通知を送れなければならず、cancel 要求で実行中ジョブを中断できなければならない。

#### Scenario: Cancel running convert
- **WHEN** 変換実行中に GUI または RPC から cancel する
- **THEN** エンジンは追加の成果物書き込みを止め、中断として応答する

### Requirement: Payload hygiene
システム SHALL RPC 応答に成果物バイナリ全体を載せてはならず、パス参照または短い要約に留め、ログに設定値や秘密情報を出してはならない。

#### Scenario: Convert result reference
- **WHEN** `convert.run` が成功する
- **THEN** 応答は出力パスまたは一時成果物参照と要約メタデータを含み、設定全文を含まない

