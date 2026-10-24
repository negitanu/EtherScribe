## Context

現状は `scripts/build-engine.sh`（Go バイナリのみ）と `scripts/dev-desktop.sh`（ビルド後に Electron を直起動）がある。`desktop/package.json` には `dev` / `build` / `pack` / `start` があるが、リポジトリルートからの統一入口がない。

## Goals / Non-Goals

**Goals:**

- ルートから `./scripts/build.sh` でデスクトップ一式をビルドできる
- ルートから `./scripts/run.sh` で開発起動（既定）またはビルド済み起動ができる
- 既存の `build-engine.sh` は内部部品として維持する
- README / Agent.md の手順を新入口に合わせる

**Non-Goals:**

- CI マトリクスやクロスコンパイル配布パイプラインの整備
- インストーラ（dmg / nsis 等）の追加（現状どおり `electron-builder --dir`）
- Go CLI 単体の起動ラッパ（`go run` / 既存 CLI で十分）

## Decisions

1. **`build.sh` は `npm run pack` 相当**  
   エンジン → Vite renderer → Electron main/preload → `electron-builder --dir`。成果物は `desktop/release/`。`--no-pack` でディレクトリ成果物のみ（`npm run build`）も選べる。

2. **`run.sh` 既定は開発モード**  
   エンジンビルド → `npm run build:electron` → Vite を起動し、`VITE_DEV_SERVER_URL` を渡して Electron を起動。`--prod` で `npm run build` 後に `electron .`（または release 内アプリがあればそれを優先）。

3. **`dev-desktop.sh` は `run.sh --prod` 相当へ委譲**  
   既存呼び出しを壊さない薄い互換ラッパにする。

4. **依存インストール**  
   `desktop/node_modules` が無い場合は `npm install` を実行する。Go は既存と同様にホストの `go` を前提とする。

## Risks / Trade-offs

- [Risk] 開発モードで Vite と Electron の二重プロセス管理が複雑 → Mitigation: フォアグラウンドで Vite を起動し、別プロセスで Electron を起動。終了時に両方を片付ける trap を置く
- [Risk] macOS の `.app` パスが OS ごとに違う → Mitigation: `--prod` ではまず `desktop/` の `electron .` を既定とし、`--app` で release 内 `.app` / 実行ファイルを明示起動可能にする

## Migration Plan

- 新スクリプト追加後、README / Agent.md のコマンドを差し替え
- 旧 `dev-desktop.sh` は委譲のみで残す

## Open Questions

- なし（インストーラ形式は後続）
