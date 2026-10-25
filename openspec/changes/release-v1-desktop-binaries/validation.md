# v1.0.0 配布検証（2026-10-26）

## ローカル確認

- 全 Go テスト、fixture key gate、UI/Electron ビルド: 成功。
- Mac arm64 ZIP/DMG: 生成成功。ad-hoc codesign の deep/strict 検証成功。
- app.asar と外側 resources の許可リスト、ライセンス原文、Mac アプリ/エンジンの arm64 形式、1.0.0 の版、RPC handshake: 成功。
- Windows amd64 エンジンのクロスコンパイルと PE x86-64 形式: 成功。
- 配布拒否ゲート（config / .env / source map / version / CPU）6 テスト: 成功。
- actionlint と OpenSpec strict 検証: 成功。
- 公開ソース・履歴の監査: タグ時点で 1922 追跡パス、814 非上流 blob、1155 一意の Batfish 原典 blob を確認。実設定・資格情報の検出なし。Release タグを含めて再検査済み。
- 本人名義の Git 情報、外部依存の著作権表示は維持。公開テスト用設定はソースに保持し、配布物へは含めない。

## GitHub CI と公開結果

- 対象コミット: `5b53da29a3a57424f456ef9995c93a8e0d3e1f65`、タグ `v1.0.0`。コミット・タグ名義は所有者本人。
- [main CI](https://github.com/negitanu/EtherScribe/actions/runs/40237092565): 全成功。
- [タグ CI / 公開](https://github.com/negitanu/EtherScribe/actions/runs/40237453292): クロスコンパイル、Windows x64、macOS arm64、publish とも成功。
- [Release v1.0.0](https://github.com/negitanu/EtherScribe/releases/tag/v1.0.0): 公開済み。Windows ZIP、Mac ZIP/DMG、SHA256SUMS.txt の 4 件。
- 全 4 件を認証なしで実ダウンロードし、サイズ一致と 3 配布物の SHA-256 一致を確認した。
- Windows ZIP: 112,957,833 bytes。Mac ZIP: 96,681,688 bytes。Mac DMG: 100,076,397 bytes。

Windows の版/RPC 実行確認は対象 OS の CI で成功。Apple notarization / Windows Authenticode は実施せず、Mac の ad-hoc 署名を検証した。制約は Release notes に記載済み。署名やセキュリティ設定については [Apple 公式案内](https://support.apple.com/ja-jp/102445) を参照。

CI 検証中に Windows の ASAR 読み取りパス区切りを修正した。本人の HTTPS 認証に workflow scope がなかったため、既存の本人 SSH 接続で push し、OAuth 権限は拡張していない。配布物に関係しない既存未コミット変更は維持した。
