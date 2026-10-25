# 検証記録（2026-10-25）

## 公式根拠

[公式リファレンス対照](../../../docs/reference/official-parser-grounding.md) に出典・対象バージョン・実装・合成 fixture を記録した。FortiOS HTML の 403 を公式配布 PDF で補完し、7.4.0 の secondary-IP、LACP 既定値、static/static6 を確認した。Junos は公式 statement reference と deactivate/構文説明、PAN-OS は 11.1 Configure CLI hierarchy を確認した。

## 成功した確認

- `cd core && go test ./...`
- `./scripts/check-fixture-keys.sh`
- `./scripts/build-engine.sh`
- `openspec validate ground-parser-network-semantics --strict`
- `git diff --check`

## 回帰内容

- 新規合成 fixture 5 件。Junos の compact brace/set 同値、引用符・コメント、不正入力の値を含まない診断。
- BGP の global/group/neighbor 継承、instance の AS、対向 AS の override、local-as の分離、deactivate。
- 静的ルートの属性統合、qualified-next-hop、next-hop リスト、inet6.0、並べ替えても変わらない意味。
- FortiOS secondary IP / IPv6 / VRF / MTU / static aggregate / 宛先オブジェクトとデフォルトルートの区別。
- PAN-OS IPv6 / MTU / vsys と VR の両方の所属。
- JSON / HTML / Excel の追加項目保持と Excel XML の妥当性。

## 限界

実機・実設定の動作照合、Batfish との出力差分は行っていない。新規 Cisco パーサは追加していない。出力は設定台帳であり、運用状態や経路到達性の検証ではない。既存の未コミット変更は維持し、コミット・ステージは行っていない。
