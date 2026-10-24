## ADDED Requirements
### Requirement: Explicit publication boundary
リポジトリ SHALL 秘密、実設定、DB、バックアップ、ローカル環境、キャッシュ、ログ、生成物を除外し、ソース・合成 fixture・仕様・検証記録・必要な外部依存を追跡する。
#### Scenario: Publication candidates
- **WHEN** 初回コミットを準備する
- **THEN** 全追跡候補と秘密情報の走査結果を確認し、生成物はステージされない
### Requirement: Owner attributed local commits
リポジトリ SHALL 本人のローカル設定の名義だけで目的別コミットを作成し、外部依存の著作権表示を維持する。
#### Scenario: Publish main
- **WHEN** main を GitHub に push する
- **THEN** 全 push 対象の author と committer が本人名義であり、自動の共同著者表記がなく、リモート SHA がローカルと一致する
