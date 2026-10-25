# パーサ強化用テストコーパス

## Why
複雑な構文・機種差・未対応形式の検証材料を再現可能な形で整備する。

## What Changes
- `.test/batfish/` に Cisco / ASA / XR、FortiOS、Juniper、Palo Alto、FRR の公開テストリソースを原形で取得する。
- コミット固定・SHA-256 manifest・上流ライセンスを保存し、外部原典と独自合成 fixture を分離する。
- `core/testdata/` に複合シナリオとテストパターンを追加し、対応形式の意味と未対応形式の拒否を検証する。

## Impact
パーサの対応機種は拡張しない。通常のテストで外部アクセスや実設定を使わない。
