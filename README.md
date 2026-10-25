# EtherScribe

FortiGate / Palo Alto / Junos EX（ELS）の設定ファイルからパラメータシートを生成するデスクトップアプリです。  
**Go エンジン** + **Electron GUI** で macOS / Windows / Linux を対象にします。

## このフォルダで作業する

開発時は **このディレクトリをルートに開く**（親の `src` ではなく）。

- エージェント指針: [AGENTS.md](AGENTS.md) / [AGENT.md](AGENT.md)
- WallScribe 由来の参照: [docs/reference/wallscribe-essence.md](docs/reference/wallscribe-essence.md)

## 識別情報

| 項目 | 値 |
| --- | --- |
| Go module | `github.com/negitanu/etherscribe` |
| 表示名 | EtherScribe |
| バンドル ID | `com.negitanu.etherscribe` |
| エンジン版 | 0.1.0 |

## クイックスタート（エンジン）

```bash
cd core
go test ./...
go run ./cmd/etherscribe version
go run ./cmd/etherscribe convert -i testdata/fortigate-vdom.conf -o /tmp/out.json -f json
go run ./cmd/etherscribe compare --before testdata/fortigate-vdom.conf --after testdata/fortigate-vdom.conf -o /tmp/diff.json -f json
```

## デスクトップ

```bash
# 開発起動（エンジンビルド + Vite + Electron）
./scripts/run.sh

# 本番相当で起動（ビルド済み renderer）
./scripts/run.sh --prod

# 配布用ディレクトリ成果物（desktop/release/）
./scripts/build.sh

# パックなしビルドのみ
./scripts/build.sh --no-pack
```

## OpenSpec

```bash
openspec list
openspec show plan-etherscribe-desktop
```

## GUI と出力

- 「パラメータシート」「設定差分」をサイドバーで切り替えます。
- HTML（初期値）/ Excel / JSON を選び、用途別プリセットまたは個別項目を指定して保存します。
- スイッチ用は switch-overview / switch-l2 / switch-l3 / switch-ops / switch-all。Junos EX の `.conf` / `.set` に対応します。
- 保存後に機器概要・件数・保存先が表示され、HTML は画面内でプレビューできます。
- HTML は目次と印刷レイアウト、Excel は固定見出し・フィルター・折返し付きです。両形式に選択したモデル全項目と解析範囲を含めます。内部の詳細項目名は一部英語です。
- 差分対象は現在インターフェース・ルート・ポリシー、出力は HTML / JSON です。PDF、HA 複数ファイル統合、lab / ネットワーク調査は後続です。

GUI のスモークテスト（ダイアログのみ置換し、実 Electron・IPC・Go エンジンで合成 fixture を処理）:

```bash
./scripts/build.sh --no-pack
env -u ELECTRON_RUN_AS_NODE desktop/node_modules/.bin/electron scripts/desktop-smoke.cjs
```

任意で `ETHERSCRIBE_SCREENSHOT=/tmp/etherscribe.png` を指定すると画面を保存できます。テストの設定・出力・プロファイルは一時ディレクトリに作成し、終了時に削除します。

## v1.0.0 ダウンロード

[GitHub Release v1.0.0](https://github.com/negitanu/EtherScribe/releases/tag/v1.0.0) から Windows x64 ZIP、Mac Apple Silicon ZIP/DMG、SHA256SUMS.txt をダウンロードできます。署名・実行時の制約は [Release notes](docs/releases/v1.0.0.md)、ビルドと検査は [Release 手順](docs/release-process.md) を参照してください。
