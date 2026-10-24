## Why

WallScribe は FortiGate / Palo Alto の設定からパラメータシートを生成できるが、Python/Flask の Web・CLI 配布であり、オフラインのデスクトップ配布やネイティブ性能を前提にした再構築が必要になった。EtherScribe として Golang コアと Electron GUI で、macOS / Windows / Linux に同じ成果物を届ける。

## What Changes

- WallScribe 相当の「設定取り込み → 統一モデル → パラメータシート出力」を Go で実装する。
- Electron デスクトップシェルからファイル選択・変換・プレビュー・保存を完結させる。
- Go バイナリを CLI としても単独実行可能にし、Electron とは IPC で接続する。
- 初期出力は HTML / Excel / JSON。同一機器の 2 世代差分を MVP に含める。
- 設定ファイルと生成物をアプリが永続保存しない方針を維持する。
- WallScribe の構築テスト・ネットワーク調査・PDF は後続フェーズとする。

## Capabilities

### New Capabilities

- `desktop-shell`: Electron によるマルチプラットフォーム GUI、Go サイドカー起動、進捗・結果表示。
- `engine-ipc`: Go エンジンの CLI エントリと Electron 向け JSON-RPC（stdio）契約。
- `config-ingestion`: 設定ファイル入力、文字コード検出、機器種別識別、サイズ検証。
- `vendor-parsing`: FortiGate (.conf) / Palo Alto (.xml) から統一モデルへの解析。
- `parameter-sheets`: HTML / Excel / JSON のパラメータシート生成と出力範囲（プリセット／セクション）。
- `config-diff`: 同一機器の変更前・変更後比較レポート。

### Modified Capabilities

なし（新規プロジェクト）。

## Impact

新規リポジトリ構成（`core/` Go、`desktop/` Electron）、パッケージング（electron-builder 等）、合成 fixture と Go テスト、IPC 契約ドキュメントが必要。WallScribe コードの機械移植は必須とせず、仕様・fixture・出力契約を参照する。実機設定の書き込み、サーバー常駐 Web、クラウド連携は初期対象外。
