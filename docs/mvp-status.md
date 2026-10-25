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

## 項目名・オブジェクト解析の修正（2026-10-25）

OpenSpec: `align-wallscribe-semantic-reports`。

- [x] 汎用 flatten を廃止し、日本語の固定列・1オブジェクト1行に修正。アドレス／グループ／サービスを別表に分離。
- [x] 全モデルフィールドの日本語ラベルと、全ネスト枝の値が表示に残ることを自動検証。
- [x] FortiGate: 引用符付き edit 名、タブ区切り、空文字、再編集、IPv6、タグ、複数プロトコルとポート、説明。
- [x] Palo Alto: shared/vsys の区別、共通のアドレス種別、タグ、動的グループ、サービスの送信元ポートと説明。
- [x] BGP ネイバー／アプリ制御等の子レコードに親設定と親コンテキストを表示。
- [x] Excel の日本語を考慮した列幅・行高、使用禁止文字・長さ・重複するシート名の処理。
- [x] HTML と XLSX の実データのヘッダー・全セルを照合。5種の fixture で XLSX 全 XML の整合性を検証。
- [x] 生成 HTML を実 Electron で表示し、FortiGate／Palo Alto の列名、値、行幅、エスケープを確認。

互換性の修正: Palo Alto のオブジェクト type は `subnet` / `iprange` / `fqdn`、サービス protocol は `TCP` / `UDP` に統一。既存 JSON キーは維持し、詳細を追加キーで拡張する。出力物とテスト用画面はリポジトリ外の一時領域にのみ生成する。

本修正の最終確認: `go test ./...`、`build.sh --no-pack`、fixture key check、OpenSpec 全11項目の strict 検証、実 Electron の GUI smoke は成功。コミット／push は未実施。


## ローカル実設定での解析強化（2026-10-25）

OpenSpec: `harden-real-config-ingestion`。詳細は `docs/private-config-validation.md`。

- [x] 実設定2件を `.data/` から読み取り、元ファイル名や値を出さず検証。`.data/` / `data/` はディレクトリ単位で Git 除外。
- [x] FortiGate の VDOM 暗黙終端による後続定義の読み落としを修正し、評価順と再編集時の順序を保持。
- [x] Palo Alto の集約・論理・サブ IF、モード、複数 IP / DHCP、LAG/VLAN の関連付け、管理設定、アプリベースのルール詳細を追加。
- [x] FortiGate の VDOM 内と Palo Alto の未モデル化項目を、値を含まない構造パスと件数で付録へ出力。
- [x] 実設定の独立したソース件数／順序とモデルを照合し、JSON/HTML/Excel のメモリ内生成・XLSX XML 検証を通過。
- [x] 新しい構文は独立した合成 fixture で回帰検証。通常テストは実設定なしで成功し、実設定テストは opt-in のみ。
- [x] Go 全テスト、7 fixture の出力検証、fixture key check、TypeScript、ビルド、OpenSpec 全12項目の strict 検証が成功。
- [x] 実 Electron / Go IPC の生成・差分・プレビュー・取消・エラー復帰を合成 fixture で検証。
- [x] 実設定2件の Git 除外・追跡数0を確認。実データのレポート保存・コミット・push は未実施。

## 公式 CLI に基づくネットワーク解析の強化

`ground-parser-network-semantics` で Junos BGP 継承・自 AS/対向 AS 分離、静的ルート属性統合、LACP 明示設定、compact brace/inactive を修正。FortiOS の複数 IPv4/IPv6・VRF・MTU・aggregate・経路属性、PAN-OS XML の IPv6・MTU・virtual-router 所属を追加し、JSON/HTML/Excel に反映。公式根拠・対応限界は [対照記録](reference/official-parser-grounding.md)、検証は `core/internal/pipeline/grounded_network_test.go`。Go 全テスト・fixture gate・エンジンビルド成功。
