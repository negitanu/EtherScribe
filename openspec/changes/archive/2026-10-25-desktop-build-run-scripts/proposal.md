## Why

ルートからアプリ全体をビルド／起動する入口が分散しており、`build-engine.sh` と `dev-desktop.sh` だけでは「配布用ビルド」と「日常の起動」が分かりにくい。開発者が一コマンドで再現できる導線が必要。

## What Changes

- デスクトップアプリ一式をビルドする `scripts/build.sh` を追加する（エンジン同梱 + Vite + Electron 本ビルド、任意で `electron-builder --dir`）
- アプリを起動する `scripts/run.sh` を追加する（開発モード既定、本番ビルド起動オプション）
- 既存の `dev-desktop.sh` は `run.sh` へ委譲するか同等の薄いラッパに整理する
- README / Agent.md のクイックスタートを新スクリプトに合わせて更新する

## Capabilities

### New Capabilities

- `build-run-scripts`: リポジトリルートから EtherScribe デスクトップをビルド／起動するシェル入口の要件

### Modified Capabilities

- （なし）— アプリ本体の機能要件は変更しない

## Impact

- `scripts/`（新規・既存ラッパ）
- `README.md` / `Agent.md` の起動手順
- `desktop/package.json` の npm scripts は既存を再利用（必要なら薄いエイリアス追加）
