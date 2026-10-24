## Context

2026-10-25 時点で EtherScribe リポジトリは OpenSpec 初期化直後。実装コードは未着手。機能参照は隣接の WallScribe（Python）の仕様・fixture・出力項目とする。ユーザー要求の技術制約は **Golang + Electron** によるマルチプラットフォーム GUI。

WallScribe の強みはベンダー横断の統一モデルと非保存の変換パイプラインである。EtherScribe では同じドメインをデスクトップ配布向けに切り出し、コアを Go に寄せて単体 CLI としても使えるようにする。

## Goals / Non-Goals

**Goals:**
- FortiGate / Palo Alto 設定から HTML・Excel・JSON のパラメータシートを生成する。
- 差分比較を GUI / CLI の両方から実行できる。
- Electron アプリとして macOS / Windows / Linux 向けにパッケージできる骨格を持つ。
- Go コアは Electron なしで CLI 完結する。
- アプリが設定原文・生成物をディスクへ永続保存しない（ユーザーが明示保存した出力先を除く）。

**Non-Goals:**
- WallScribe 全機能（構築テスト、ネットワーク調査、PDF、Web サーバー、API）の初回一括移植。
- 実機への設定書き込み・SSH 取得。
- Wails / Tauri への切り替え（明示的に Electron を採用）。
- マルチユーザー認証付き共有サーバー。

## Decisions

### 1. 二層構成: Go コア + Electron シェル

```
[Renderer UI] ←IPC(preload)→ [Electron Main] ←stdio JSON-RPC→ [etherscribe engine]
```

- **Go (`core/`)**  
  解析、モデル、exporter、差分、CLI。パッケージ例: `internal/ingest`, `internal/parse/{fortigate,paloalto}`, `internal/model`, `internal/export/{html,excel,json}`, `internal/diff`, `internal/rpc`。エントリは `cmd/etherscribe`。
- **Electron (`desktop/`)**  
  Main がエンジン子プロセスを起動し、Renderer は Preload 経由でのみ RPC を呼ぶ。設定本文をログやクラッシュレポートへ出さない。

**代替と不採用理由:**
- Wails / Tauri: 単一ランタイムで軽いが、ユーザー指定が Electron。
- 純 Go GUI (Fyne 等): Web 技術のプレビュー（HTML シート）との相性が弱い。
- Go を HTTP ローカルサーバー化のみ: ポート競合と Origin 管理が増える。stdio RPC を主、必要なら将来 loopback を追加。

### 2. IPC 契約（stdio JSON-RPC 2.0）

1 行 1 メッセージ（NDJSON）。Main がリクエスト、engine が応答。長時間処理は `progress` 通知を返す。

初期メソッド:
| Method | 用途 |
| --- | --- |
| `engine.version` | 版と対応能力 |
| `convert.run` | 単体／HA 変換（入力パスまたは一時読込バイト、形式、preset/sections） |
| `compare.run` | 2 世代差分 |
| `sections.list` | 出力セクション一覧 |
| `cancel` | 実行中ジョブの取消 |

入力は原則 **ユーザーが選択したローカルパス** を渡し、巨大ファイルの二重コピーを避ける。Renderer が全文を保持する必要はない。結果は一時ディレクトリまたはメモリハンドル経由で Main が保存ダイアログへ渡す。一時ファイルはジョブ終了時に削除する。

### 3. ドメインモデルと WallScribe との関係

WallScribe の `ConfigModel` / セクション定義 / 秘匿方針を **契約の参照** とする。Go で同等の中間モデルを定義し、フィールド名・セクション ID は可能な限り WallScribe の `exporters/sections.py` 相当に揃えて回帰比較しやすくする。

パーサは Go で新規実装する。Python コードの literal port は必須としない。合成 fixture（匿名化済み）を `testdata/` に置き、主要セクションの JSON スナップショットで差分検知する。

秘匿: SNMP コミュニティ、`ENC` 文字列、パスワード類は HTML/Excel/JSON に出さない。ログにホスト名・設定値・ファイル名全文を出さない。

### 4. 出力と MVP 範囲

| 形式 | MVP | 備考 |
| --- | --- | --- |
| JSON | 含む | 機械可読・テスト容易 |
| HTML | 含む | アプリ内 WebContents プレビュー可 |
| Excel | 含む | excelize 等 |
| PDF | 後続 | ネイティブ依存が大きい |
| 差分 | 含む | 上記形式に準拠 |
| 構築テスト / ネットワーク図 | 後続 | WallScribe lab / network |

プリセット（概要・接続・ポリシー監査・運用・すべて）とセクション個別指定を CLI・GUI 共通で提供する。

### 5. リポジトリ構成（初期）

```
EtherScribe/
  Agent.md
  README.md                 # 後続タスクで追加
  openspec/
  core/                     # Go module (github.com/.../etherscribe は後で確定)
    cmd/etherscribe/
    internal/...
    testdata/
  desktop/                  # Electron + TypeScript + Vite
    src/main/
    src/preload/
    src/renderer/
  scripts/                  # ビルド・エンジン同梱
  docs/                     # IPC 契約、配布手順（実装時）
```

Go の module path と Git remote は初回 scaffold 時に確定する。

### 6. パッケージング

- 開発: `go run` / `go build` でエンジン、`electron` が相対パスのバイナリを起動。
- 配布: electron-builder（または同等）で各 OS 向けにパッケージし、`extraResources` に OS 別 `etherscribe` バイナリを同梱。
- 署名・公証は後続（macOS notarization 等）。MVP は未署名ローカルビルドで可。

### 7. 言語・テスト・品質

- Go: 表層 API のユニットテスト + fixture スナップショット。`go test ./...`
- Desktop: UI は薄い。IPC クライアントの契約テストを優先。
- OpenSpec: 変更ごとに strict validate。実装は tasks のチェック完了後に archive。

## Risks / Trade-offs

- **パーサ互換の乖離** → WallScribe fixture と JSON 出力の比較ゲートを早期に入れる。完全一致が難しい箇所は coverage（解析範囲）で明示。
- **Electron 同梱バイナリのパス／権限** → Main の解決ロジックを OS 別にテストし、asan や Gatekeeper は後続で扱う。
- **stdio の巨大レスポンス** → 成果物はファイルパス参照を返し、RPC ペイロードにバイナリを載せない。
- **PDF / lab の期待ギャップ** → Agent.md と README で MVP 境界を明示。
- **二重メンテ（WallScribe と EtherScribe）** → 当面は仕様参照のみ。共通化ライブラリ化はしない。

## Migration Plan

1. **P0** 本 change の設計合意、scaffold、IPC Hello World。
2. **P1** ingest + FortiGate 最小パース + JSON 出力 + CLI。
3. **P2** Palo Alto、HTML/Excel、差分、coverage。
4. **P3** Electron UI（変換・差分・プレビュー・保存）。
5. **P4** 各 OS パッケージ試作、回帰 fixture 拡充。
6. 後続 change で PDF・構築テスト・ネットワーク調査を個別提案。

ロールバックはデスクトップ未配布のため Git revert で足りる。ユーザー明示保存ファイルはアプリ管理外。

## Open Questions

1. Go module path / GitHub（または Origin）リポジトリ名。
2. アプリ表示名・バンドル ID（例: `EtherScribe` / `com.*.etherscribe`）。
3. HA クラスタ解析を MVP に含めるか（WallScribe 相当）。暫定: P2 以降。
4. Excel ライブラリ選定（excelize を第一候補）。
5. WallScribe のどの fixture を最初のゲートにするか（優先: VDOM / vsys の代表例）。
6. 自動更新（electron-updater）の要否。暫定: 不要。

未回答でも P0–P1 の骨格は着手可能。配布 ID と module path は scaffold 前に確定する。
