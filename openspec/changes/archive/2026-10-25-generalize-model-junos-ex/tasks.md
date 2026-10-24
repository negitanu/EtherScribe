## 1. モデル一般化（パーサより先）

- [x] 1.1 `DeviceInfo` に Role / VendorFamily を追加し、FG/PA パーサで既存どおり Role=firewall をセットする
- [x] 1.2 `VirtualContext` と Config.VirtualContexts を追加する（default コンテキスト規約を文書化）
- [x] 1.3 VLAN / LAG / STP モデルと Interface の L2 属性・IRB 向け type を追加する
- [x] 1.4 Route / Routing（OSPF・BGP）に Context フィールドを追加し、既存 FG/PA は default で埋める
- [x] 1.5 `go test ./...` で既存 FG/PA 回帰が通ることを確認する

## 2. セクションとプリセット

- [x] 2.1 SectionKeys に vlans / lag / stp を追加する
- [x] 2.2 スイッチ専用プリセット switch-overview / switch-l2 / switch-l3 / switch-ops / switch-all を追加する（FW セクションを含めない）
- [x] 2.3 JSON / HTML / Excel で新セクションを出力する（未使用時は空で可）
- [x] 2.4 `sections.list` / version vendors 表示を更新する

## 3. 取り込みと識別

- [x] 3.1 Junos EX 識別（set / brace、ethernet-switching または vlans 等）を identify に追加する
- [x] 3.2 ingest で `.set` / `.txt` の Junos テキストを許可し、識別失敗時は変換しない

## 4. Junos 正規化

- [x] 4.1 `parse/junos` に Statement 型と set 行パーサを実装する（deactivate/delete 方針は design どおり）
- [x] 4.2 brace 階層を Statement 列へ flatten する
- [x] 4.3 同一論理設定の brace / set fixture で正規化結果が一致することをテストする

## 5. Junos EX マッパー

- [x] 5.1 system（host-name、版・機種ヒント）、interfaces（物理/論理、L2、ae、irb）をモデルへ載せる
- [x] 5.2 vlans（ELS）と STP 基本を載せる
- [x] 5.3 routing-instances（VRF）と静的／OSPF／BGP（コンテキスト付き）を載せる
- [x] 5.4 未対応ブロックを coverage に残し、合成 fixture（brace + set）で結合テストする
- [x] 5.5 pipeline convert / compare（同一 Junos 同士）と秘匿がないことの確認

## 6. 文書とクローズ

- [x] 6.1 `docs/reference/` に Junos EX 抽出メモ（対象ステートメント、ELS 前提）を追加する
- [x] 6.2 Agent.md / wallscribe-essence 相当の記述を更新する
- [x] 6.3 `go test ./...` と fixture キー検査、openspec validate
