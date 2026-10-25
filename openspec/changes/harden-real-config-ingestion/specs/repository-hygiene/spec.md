## ADDED Requirements
### Requirement: Private configuration validation
システム SHALL `.data/` と `data/` を Git の追跡対象から除外し、実設定のローカル検証は opt-in にする。通常のテストは実設定に依存してはならない。
#### Scenario: Private files with arbitrary extensions
- **WHEN** 利用者が実設定や関連ファイルを `.data/` または `data/` に配置する
- **THEN** 拡張子によらず Git 除外され、既定の検証・ログ・ドキュメントに実設定の値が出ない
### Requirement: Aggregate only diagnostics
検証ツール SHALL ファイル本文・識別子・秘密・生の例外文字列を表示せず、構造別件数と成功・失敗のみを報告する。
#### Scenario: Malformed private configuration
- **WHEN** ローカル実設定の解析でエラーになる
- **THEN** 入力の本文や元ファイル名を出さず、匿名サンプル番号と失敗区分を示す
