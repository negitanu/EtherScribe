## 1. スクリプト

- [x] 1.1 `scripts/build.sh` を追加（`--no-pack` 対応、node_modules 自動 install）
- [x] 1.2 `scripts/run.sh` を追加（既定開発 / `--prod`、終了時クリーンアップ）
- [x] 1.3 `scripts/dev-desktop.sh` を `run.sh --prod` へ委譲

## 2. ドキュメント

- [x] 2.1 README.md のデスクトップ手順を `build.sh` / `run.sh` に更新
- [x] 2.2 Agent.md / AGENTS.md の再開手順に新スクリプトを追記

## 3. 検証

- [x] 3.1 `./scripts/build.sh --no-pack` が成功することを確認
- [x] 3.2 `openspec validate desktop-build-run-scripts --strict`（または status）でアーティファクト整合を確認
