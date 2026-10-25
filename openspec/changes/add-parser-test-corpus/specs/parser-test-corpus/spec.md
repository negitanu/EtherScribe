## ADDED Requirements

### Requirement: Reproducible reference corpus
開発環境 SHALL 外部テスト原典を独自 fixture と分離し、固定 revision・出典・ライセンス・内容ハッシュを保持する。
#### Scenario: Fetch public resources
- **WHEN** 取得スクリプトを実行する
- **THEN** 指定したベンダーの原典を内容変更なしで取得しハッシュ検証する

### Requirement: Offline synthetic regression
通常のテスト SHALL `core/testdata/` の合成設定だけを使い複合構造の意味と対応限界を検証する。
#### Scenario: Supported complex input
- **WHEN** 多コンテキストや L2/L3 を含む対応形式を解析する
- **THEN** 件数・コンテキスト・評価順を保持し JSON/HTML/Excel を生成できる
#### Scenario: Unsupported input
- **WHEN** Cisco または PAN-OS set の合成設定を解析する
- **THEN** 対応済みとして扱わずエラーを返す
