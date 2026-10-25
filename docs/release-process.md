# デスクトップ Release

GitHub Actions の `Desktop binaries` が Go を Windows amd64 / Darwin arm64 向けにクロスコンパイルし、Windows x64 / Mac Apple Silicon ランナーで Electron アプリへ梱包する。

## 配布対象と検証

- Windows: ZIP（展開して起動）。Mac: Apple Silicon の ZIP / DMG。
- 固定した npm lockfile を `npm ci` で使い、GitHub Actions はコミット SHA に固定する。
- 追跡ソースと履歴の資格情報パターン・機微ファイルパスを検査する。Batfish 原典は固定 blob と SHA-256 の一致を検証する。
- Go ビルドでは `trimpath` / `buildvcs=false` を使用し、個人のパスと Git ビルドメタデータを除く。
- `app.asar` は UI、コンパイル済み Electron main/preload、package.json のみ。エンジンと第三者ライセンスを外側へ同梱する。
- 各 OS でアプリと Go の実行ファイル形式、エンジンの版、JSON-RPC 起動を検証する。OS の実機 UI 操作や署名証明書の検証を代替するものではない。
- ソースの本人名義の Git 作者情報と第三者著作権表示は意図的に維持する。検査で資格情報パターンが検出されないことは、任意の秘密を完全に判定できることを意味しない。

## 公開

1. アプリ・lockfile・Go の版と Release notes を揃え、OpenSpec と合成回帰を検証して本人名義でコミットする。
2. main へ push し両 OS の artifact と検査の成功を確認する。
3. 対象コミットに `v1.0.0` タグを付けて push する。
4. タグ CI が再検証し、SHA256SUMS.txt を生成する。draft を作り全ファイルをアップロードしてから公開する。
5. Release の 3 配布物とチェックサムをダウンロードして照合する。

途中失敗した draft は公開されない。既存タグ・Release の上書きはせず原因を直す。新しい版では notes のパスも更新する。

Windows の Authenticode 署名と macOS notarization は未設定。Mac は ad-hoc 署名を行い、Release notes に利用時の制約を記載する。
