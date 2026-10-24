## Why
初回公開に先立って、不要データと秘密情報を除外し、本人名義・目的別コミットのルールを永続化する。
## What Changes
- AGENT.md に Git・データ整理および名義・Contributor の指定ルールを記録する。
- .gitignore を拡張し、ソース・合成 fixture・仕様・検証記録と必要な外部依存を追跡する。
- Go コアと Electron アプリを独立した検証可能なコミットに分け、GitHub origin の main へ通常 push する。
## Capabilities
### New Capabilities
- `repository-hygiene`: 公開対象・名義・履歴検査。
## Impact
Git メタデータ、ルール・除外設定・公開手順。アプリ機能は変更しない。外部依存の著作権とライセンスを維持する。
