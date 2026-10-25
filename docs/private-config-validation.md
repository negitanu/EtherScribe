# 実設定を使ったローカル検証と解析改善

検証日: 2026-10-25（日本時間）

OpenSpec: `harden-real-config-ingestion`。

## データの扱い

- 実設定2件（FortiGate / Palo Alto XML）は利用者の `.data/` に置いたまま読み取り専用で検証。
- `.data/` と `data/` をディレクトリ全体で Git 除外し、関連メモや拡張子のない入力も追跡しない。
- 設定本文、ホスト名、IP、元ファイル名、オブジェクト名、秘密、生の例外メッセージは検証出力・この記録へ転記しない。
- 実設定由来のレポートや置換コピーは作らず、出力形式の検証もメモリ内で完結する。調査用の一時プローブコードは削除済み。
- 通常テストに実設定の依存を加えず、観測した構文だけを独立した予約 IP・架空の名前の fixture で再現する。

## 確認した不具合と対応

FortiGate では VDOM の edit に next がない export 終端を誤解釈し、警告とともに後続の定義を読み落としていた。VDOM に限定して暗黙の edit 終端を処理し、通常ブロックの不正終端は引き続き警告する。ポリシーの順序を内部メタデータで保持し、再編集しても順序を動かさない。

修正前後の集計（秘密値・識別子を含まない件数のみ）:

- FortiGate: 解析警告 4 → 0、アドレス 14 → 25、サービス 90 → 176、ポリシー 5 → 6、静的ルート 0 → 1。IF は 27。修正後は独立したソースカウンターの件数とポリシー順序に一致。
- Palo Alto: IF 8 → 17、LAG 0 → 1、VLAN 0 → 4。アドレス 13、サービス 1、ポリシー 10、静的ルート 0。修正後は XML 要素を直接数えた対応コレクション件数に一致。

最初の FortiGate 調査で同じ設定ツリー由来の件数同士を比べた結果は、完全性の根拠にはならなかった。このため検証ツールでは別の構文スタックで定義・評価順を数え、再編集を重複カウントしない。XML 側もパーサの typed struct ではなく XML トークンのパスから独立して数える。

Palo Alto では L2／L3／tap／virtual-wire のモード、集約メンバー、サブ IF、VLAN／loopback／tunnel の論理 IF、複数アドレスと DHCP、管理 IP／マスク／GW・DNS・NTP・タイムゾーン、アプリベースのルール項目を追加。リンク状態は実機の稼働状態を推測せず設定値（未指定は auto）を表示する。ルートメトリックと管理距離も別項目にする。

未対応項目は FortiGate の VDOM 内も含めて解析範囲に集計し、Palo Alto も未モデル化の構造パスを記録する。coverage の増加は解析範囲の可視化を増やした結果であり、全機能を実装したという意味ではない。SD-WAN、認証・スケジュール定義の全面移植、Palo Alto 動的ルーティング等は後続。

## 再実行

```bash
GOCACHE=/private/tmp/etherscribe-go-cache ./scripts/check-private-configs.sh
```

このスクリプトは `ETHERSCRIBE_PRIVATE_DATA_DIR` を指定した opt-in テストだけを実行する。入力ディレクトリ不在・解析警告・件数不一致・順序不一致・出力失敗時は失敗として終了する。失敗時も入力の名前や値は出さない。通常の `go test ./...` ではこの実設定テストはスキップする。

## 合成 fixture

- `core/testdata/fortigate-vdom-boundaries.conf`: 暗黙終端を含む複数 VDOM、数値順でないポリシー、再編集、後続定義、未対応 VDOM ブロックの付録。
- `core/testdata/paloalto-network-ops.xml`: L2 集約と VLAN、tap／virtual-wire、L3 サブ IF、IPv6 ルート、DHCP、管理設定、アプリ・ユーザー・URL カテゴリ・プロファイル、未モデル化項目。

具体値の回帰、メタデータとオブジェクト名の非衝突、通常ブロックの不正終端、エラーメッセージの秘匿、全選択モデル値の表出力、7種類の fixture の出力整合性を通常テストで確認する。


## 最終確認

Go 全テスト、合成 fixture の出力整合性、fixture key check、renderer / Electron TypeScript、`build.sh --no-pack`、実 Electron / Go IPC smoke、OpenSpec 全12項目の strict 検証が成功。実設定2件はともに Git 除外され、追跡は0件。作業変更のコミット／push は行っていない。
