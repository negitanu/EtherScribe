# EtherScribe — エージェント／開発メモ

**このリポジトリ（EtherScribe）をルートとして開いて作業する。**  
詳細決定の正本はこのファイル。

## プロジェクト概要

- **目的**: FortiGate / Palo Alto / Junos EX（スイッチ）の設定ファイルからパラメータシートを生成するマルチプラットフォーム・デスクトップツール。
- **機能参照（リポジトリ内）**: `docs/reference/wallscribe/` と `docs/reference/wallscribe-essence.md`。隣接 `../WallScribe` は任意。
- **非目的（初期）**: 実機書き込み、常駐 Web、構築テスト／ネットワーク調査の完全移植、PDF。

## 決定済み事項

| 項目 | 決定 |
| --- | --- |
| 名称 | EtherScribe |
| Go module | `github.com/negitanu/etherscribe` |
| バンドル ID | `com.negitanu.etherscribe` |
| コア | Golang CLI + JSON-RPC |
| GUI | Electron + TypeScript |
| 作業ルート | 本リポジトリ単独（親 `src/` で他プロジェクトを混ぜない） |
| 参照移植 | TextFSM・仕様書・識別/秘匿/セクション原典を `docs/reference/wallscribe/` に同梱 |
| パーサ移植 | 識別正規表現、FortiGate ツリー、オブジェクト/NAT/IPsec、ENC 伏せ字、CIDR 変換 |
| 出力 | HTML / Excel（自前 OOXML）/ JSON + 差分 |
| 永続化 | アプリは設定・生成物を保存しない |

## OpenSpec

- `openspec/config.yaml`
- main specs: `openspec/specs/*`
- archive: `openspec/changes/archive/2026-10-24-plan-etherscribe-desktop/`、`openspec/changes/archive/2026-10-25-expand-parser-ops-network/`
- active change: `openspec/changes/port-wallscribe-desktop-quality/`（UI／台帳品質の移植、実装・検証完了）
- IPC: `docs/ipc.md` / 受け入れ: `docs/mvp-status.md`

## ディレクトリ

```
AGENTS.md / AGENT.md
core/          Go module + testdata + third_party/text
desktop/       Electron
docs/          ipc, mvp-status, reference/wallscribe
openspec/
scripts/
```

## 運用ルール

- 親ディレクトリや WallScribe 本体を etherscript 作業で変更しない。
- OpenSpec で要件を揃えてから実装する。
- 実設定・秘密・生成物をコミットしない。`core/testdata/` のみ。
- コミットは依頼時のみ。author を AI 名義にしない。
- 返答は日本語。

## 再開手順

1. `EtherScribe` フォルダを開く。
2. `AGENT.md` と `docs/reference/wallscribe-essence.md` を読む。
3. `cd core && go test ./...`
4. `./scripts/run.sh` でデスクトップ開発起動（ビルドは `./scripts/build.sh`）。


## Git・データ整理

- 不要なものをコミットしない。.gitignore を保守し、秘密情報、実データ DB、バックアップ、ローカル環境、キャッシュ、ログ、エクスポート、テスト生成物を除外する。
- ソース・必要なテスト fixture・設定例・仕様・検証記録は追跡対象とする。コミット前に追跡済み／未追跡の候補と秘密情報を確認する。
- 古い仮想環境や再生成可能な不要データは整理する。既存のユーザーデータや復旧用バックアップは、古いという理由だけで削除しない。
- コミット準備とコミット実行を区別する。既存の未コミット変更を消したり、無関係なファイルを一括ステージしたりしない。
- コミットは一括にせず、目的・機能・依存関係が分かる単位へ分割する。共有ファイルは差分単位で分け、各中間コミットでも起動・検証できる状態を保つ。
- コミットメッセージは `feat:`・`fix:`・`docs:`・`chore:` などの英語プレフィクスを使い、件名の説明と本文は日本語で変更の目的と内容が分かる表現にする。過去履歴の書き換え時は復旧用 bundle を保存し、push は明示的な lease でリモートの予期しない更新を保護する。

## コミットの名義・Contributor

- このリポジトリの author・committer は、所有者本人がローカル Git 設定で指定した名義に統一する。コミット前にリポジトリの `user.name`・`user.email` と、ステージ済み差分を確認する。
- AI・エージェント・Bot・第三者を author、committer、共同著者、contributor として追加しない。`Co-authored-by`、`Signed-off-by`、謝辞・contributors 一覧などに本人以外の名義を自動挿入しない。GitHub の Web 編集・マージなど、別の committer 名義が付く操作も避け、本人名義のローカルコミットを push する。
- 過去履歴や push 対象の全ブランチ・タグについて、author・committer・共同著者表記を検査する。名義修正は復旧用 bundle と対応表を保存し、ファイル内容・日時・親子関係を保持して行う。履歴書き換え後に旧履歴を merge し直さない。
- 明示的な指示なしに第三者の実際の著作・ライセンス表示を削除したり、第三者の成果を本人作と偽って帰属変更したりしない。外部依存の著作権表示はコミットの名義と区別して維持する。
