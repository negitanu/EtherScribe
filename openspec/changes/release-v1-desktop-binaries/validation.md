# v1.0.0 配布検証（2026-10-26）

## ローカル確認

- 全 Go テスト、fixture key gate、UI/Electron ビルド: 成功。
- Mac arm64 ZIP/DMG: 生成成功。ad-hoc codesign の deep/strict 検証成功。
- app.asar と外側 resources の許可リスト、ライセンス原文、Mac アプリ/エンジンの arm64 形式、1.0.0 の版、RPC handshake: 成功。
- Windows amd64 エンジンのクロスコンパイルと PE x86-64 形式: 成功。
- 配布拒否ゲート（config / .env / source map / version / CPU）5 テスト: 成功。
- actionlint と OpenSpec strict 検証: 成功。
- 公開ソース・履歴の監査: 1906 追跡パス、789 非上流 blob、1155 一意の Batfish 原典 blob を確認。実設定・資格情報の検出なし。最終コミット後に再検査する。
- 本人名義の Git 情報、外部依存の著作権表示は維持。公開テスト用設定はソースに保持し、配布物へは含めない。

GitHub CI と Release の結果は実行完了後に追記する。Windows 実行検証は GitHub CI の対象 OS で行う。Apple notarization / Authenticode は実施しない。
