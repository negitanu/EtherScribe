# EtherScribe エージェント指針

このファイルは **EtherScribe リポジトリをルートとして開いたとき** の作業前提です。  
詳細な決定ログは [AGENT.md](AGENT.md) を正本とします。

## 必須

- 返答は日本語。
- 作業ルートは本リポジトリ（`EtherScribe/`）。親の `src/` や他プロジェクトを勝手に変更しない。
- WallScribe の機能参照は **本リポジトリ内** の `docs/reference/wallscribe/` を使う。隣接フォルダ `../WallScribe` は任意の追加参照であり、必須ではない。
- 変更は OpenSpec（`openspec/`）で要件・設計・タスクを揃えてから実装する。
- 実設定・秘密・生成物をコミットしない。検証は `core/testdata/` の合成 fixture。
- コミットはユーザーが依頼したときだけ。author を AI 名義にしない。

## スタック

- Go コア: `core/`（CLI + JSON-RPC）
- Electron: `desktop/`
- IPC: `docs/ipc.md`

## よく使う確認

```bash
cd core && go test ./...
./scripts/build-engine.sh
./scripts/build.sh --no-pack
./scripts/run.sh
./scripts/check-fixture-keys.sh
openspec list
```

## Git 運用

コミット準備・分割・名義・データ除外の詳細は [AGENT.md](AGENT.md) の「Git・データ整理」「コミットの名義・Contributor」に従う。
