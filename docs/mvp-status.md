# MVP 受け入れ記録（plan-etherscribe-desktop）

日付: 2026-10-25

## CLI

- [x] `go test ./...`（core）成功
- [x] `etherscribe version` → EtherScribe 0.1.0
- [x] `convert` FortiGate → JSON（testdata/fortigate-vdom.conf）
- [x] `convert` Palo Alto → JSON（testdata/paloalto-vsys.xml）
- [x] `convert` Excel 生成
- [x] `compare` 異種ベンダーでエラー
- [x] `rpc` の `engine.version` ハンドシェイク

## GUI

- [x] Electron 骨格（contextIsolation / preload のみ）
- [x] 変換・差分 UI と保存ダイアログ
- [x] HTML プレビュー（srcdoc）
- [x] `electron-builder --dir --mac` で `EtherScribe.app` を生成し、同梱 `engine/etherscribe version` を確認

## 非保存

- [x] アプリ専用 DB なし。出力はユーザー指定パスまたは一時ファイルのみ

## WallScribe からの移植

- 機器識別正規表現（TextFSM 相当）
- `#config-version` フォールバック、opmode/vdom
- FortiGate 設定ツリーパーサ（config/edit/set/end）
- エンコーディング順 utf-8-sig → utf-8 → cp932 → latin-1
- 秘匿フィールド / ENC 伏せ字
- セクション／プリセット名

## WallScribe UI／台帳品質の移植（2026-10-25）

OpenSpec: `port-wallscribe-desktop-quality`（実装完了・アーカイブ前）。

- [x] 生成／差分ワークスペース、対応機器表示、HTML 初期選択
- [x] エンジンから 10 プリセット／13 セクション取得、スイッチ用途と個別選択の IPC 接続
- [x] 機器概要・結果件数・保存先、二重実行防止、保存取消／エラー復帰
- [x] 生成／差分の sandbox HTML プレビュー
- [x] 秘匿済み共通モデルから HTML/Excel 全選択項目と解析範囲を出力
- [x] HTML 目次／印刷／空状態、Excel 見出し／列幅／折返し／固定行／フィルター
- [x] 差分 HTML の表形式化と入力値エスケープ
- [x] FortiGate／Palo Alto／Junos EX fixture の回帰テスト、XLSX XML 検証、選択項目／秘匿／HTML エスケープ検証
- [x] `go test ./...`、fixture key check、renderer／Electron TypeScript、`build.sh --no-pack`
- [x] 実 Electron + Go RPC の GUI smoke（生成、スイッチ選択、プレビュー、取消、差分、エラー復帰）
- [x] OpenSpec strict validation

検証環境の Go ビルドキャッシュは権限制約を避けるため `GOCACHE=/private/tmp/etherscribe-go-cache` を指定。GUI smoke はダイアログのみ fixture／一時保存先に置換。macOS で確認済み。Windows／Linux の実機確認と配布パッケージ再生成は未実施。

未対応パーサ領域、PDF、HA 統合、lab／ネットワーク調査は従来の後続範囲を維持。今回の変更は既存 Go モデルにある項目の GUI と成果物への公開を対象とする。
