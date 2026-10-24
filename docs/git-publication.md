# Git 初回公開の検証記録

日付: 2026-10-25（日本時間）

- リポジトリ: https://github.com/negitanu/EtherScribe
- 既存 Git メタデータを再初期化。開始時のローカル履歴・リモートブランチ・タグはいずれもなし。
- リポジトリローカルの user.name / user.email: `Teppei Kobayashi <tkobayashi@kirisame.tech>`。
- `AGENT.md` に指定の Git・データ整理／コミットの名義・Contributor のルールを記録。大文字小文字の表記と参照を統一。
- `.gitignore` に秘密、実設定、DB、バックアップ、ローカル環境、キャッシュ、ログ、出力・テスト生成物を追加。ユーザーデータとバックアップの削除なし。
- 初回追跡候補684ファイルを走査。秘密鍵・主要トークン・認証付き URL・資格情報代入を確認し、公開すべき秘密は検出されなかった。トークン形式の1件は外部依存の翻訳テーブル内の単語による誤検出と確認。
- ビルド成果物・node_modules・実設定は追跡対象外。合成 fixture、仕様、検証記録、共有 OpenSpec ツールを追跡。
- `core/third_party/text` は go.mod のローカル replace に必要な依存。再生成可能なアプリ出力とは区別して同梱し、元の LICENSE・PATENTS・著作権表示・上流ソースを保持。

## コミットと検証

1. `320a75e`: Go エンジン、合成 fixture、参照仕様、OpenSpec、Git ルール。ステージ済みツリーのみを一時ディレクトリへ展開し、`go test ./...` と CLI `version` の起動を確認。
2. `0e0e10a`: Electron、ビルド・実行・GUI 検証スクリプト、共有開発ツール、操作資料。`build.sh --no-pack`、fixture key check、OpenSpec 全10項目の strict 検証を確認。アプリ機能の実 Electron / Go IPC 検証は `docs/mvp-status.md` に記録。

両コミットの author と committer が本人名義であること、全ブランチ・タグに共同著者／署名の自動 trailer がないことを検査。履歴書き換えはなく、通常 push を使用した。リモート `refs/heads/main` とローカルの `0e0e10ae601553bf56989de0950bc8c6b751beaf` の一致を確認。この記録は独立した docs コミットとして追加し、同じ main に通常 push する。

Go の検証は実行環境のキャッシュ権限に合わせ `GOCACHE=/private/tmp/etherscribe-go-cache` を使用した。
