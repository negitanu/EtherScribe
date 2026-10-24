## Decisions
- 既存 Git リポジトリは再初期化し、履歴とリモートを確認してから接続する。リモートが空なら main を通常 push する。既存履歴を強制上書きしない。
- 既存の本人名義の Git 設定をリポジトリローカルにも明示する。
- case-insensitive な環境の Agent.md を AGENT.md に統一し、参照も更新する。
- 初回の Go コミットにはローカル replace に必要な golang.org/x/text をライセンス・原著表示ごと含める。
- 設定・生成物・秘密の候補を走査し、fixture のテスト用パスワードと外部依存のテストデータは由来を確認して扱う。ユーザーデータ・バックアップの削除は行わない。
## Validation
Go 全テスト、fixture key check、TypeScript／ビルド、OpenSpec strict、ステージ差分と追跡リスト、全ブランチ・タグの author／committer／共同著者表記、push 後のリモート SHA を確認する。

- 公開前の全仕様 strict 検証で検出した config-ingestion の Purpose 文字数警告は、要件を変更せず目的の説明を補足して解消する。
