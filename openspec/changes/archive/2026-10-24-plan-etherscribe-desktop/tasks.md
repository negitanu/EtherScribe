本 change は設計と骨格までを対象とする。パーサ完全移植や各 OS 署名付き配布は後続 change でよい。検証記録は実装開始後に `docs/` へ残す。

## 1. P0 — リポジトリ骨格と合意事項の固定

- [x] 1.1 Go module path、アプリ表示名、バンドル ID を確定し Agent.md / README に反映する。検証: 文書に値が明記されている。
- [x] 1.2 `core/` に Go module と `cmd/etherscribe` の Hello（version 出力）を作る。検証: `go test ./...` と `go run ./cmd/etherscribe version` が通る。
- [x] 1.3 `desktop/` に Electron + TypeScript + Vite の最小ウィンドウを作る。検証: 開発起動で空ウィンドウが表示される。
- [x] 1.4 stdio JSON-RPC の `engine.version` だけつなぎ、Main から呼べることを確認する。検証: 起動時に版情報が UI またはログ（秘密なし）に出る。
- [x] 1.5 `.gitignore` を整備し、node_modules、ビルド成果、実設定、一時生成物を除外する。検証: 秘密・生成物が追跡候補に出ない。

## 2. P1 — 取り込みと FortiGate → JSON MVP

- [x] 2.1 エンコーディング検出と入力検証（空・サイズ・拡張子）を実装する。検証: ユニットテストで UTF-8 / CP932 / 空ファイルを確認する。
- [x] 2.2 機器識別（FortiGate / Palo Alto）を実装する。検証: fixture で識別成功・失敗を確認する。
- [x] 2.3 FortiGate 最小パーサ（機器概要・主要インターフェース／ポリシーの一部）と統一モデルを実装する。検証: 合成 `.conf` fixture でモデルが埋まる。
- [x] 2.4 JSON エクスポートと CLI `convert` を実装する。検証: CLI で JSON が出力され秘匿値が含まれない。
- [x] 2.5 coverage（解析範囲）の最小表現を JSON に含める。検証: 未対応ブロック fixture で `coverage` が付く。

## 3. P2 — Palo Alto・HTML/Excel・差分

- [x] 3.1 Palo Alto XML パーサの MVP 範囲を実装する。検証: vsys 付き合成 fixture で主要フィールドを確認する。
- [x] 3.2 HTML エクスポートとプリセット／セクション指定を実装する。検証: preset により出力セクションが変わる。
- [x] 3.3 Excel エクスポートを実装する。検証: 開可能な `.xlsx` が生成される。
- [x] 3.4 差分エンジンと CLI `compare`（JSON/HTML）を実装する。検証: 追加ポリシー fixture で追加が検出される。異種ベンダーはエラー。
- [x] 3.5 WallScribe 由来の代表 fixture との回帰ゲート（JSON 主要キー）を用意する。検証: CI またはローカルスクリプトで差分が検知できる。

## 4. P3 — Electron GUI ワークフロー

- [x] 4.1 変換 UI（ファイル選択、形式、プリセット、実行、保存）を実装する。検証: 合成設定で保存まで完了する。
- [x] 4.2 差分 UI（2 ファイル選択、実行、保存）を実装する。検証: 合成ペアで差分保存まで完了する。
- [x] 4.3 HTML プレビューと進捗／取消を実装する。検証: 実行中 cancel で中断し、一時ファイルが残らない。
- [x] 4.4 Preload 経由以外の Node 直接呼び出しを Renderer から不可にする。検証: contextIsolation / 権限境界を確認する。

## 5. P4 — パッケージ試作と設計クローズ

- [x] 5.1 OS 別エンジン同梱スクリプトと electron-builder（または同等）設定を追加する。検証: 少なくとも開発 OS でパッケージ起動し version RPC が通る。
- [x] 5.2 IPC 契約を `docs/ipc.md` に固定する。検証: メソッド一覧が実装と一致する。
- [x] 5.3 MVP 受け入れ（CLI 変換・差分、GUI 変換・差分、非保存）を記録する。検証: チェックリストとコマンド結果を docs に残す。
- [x] 5.4 OpenSpec strict validate 後、本 change を archive し main specs へ同期する。検証: `openspec validate --strict` 成功、archive 済み。
