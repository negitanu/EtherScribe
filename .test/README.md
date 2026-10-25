# パーサ強化用リソース

`batfish/grammar/` は Batfish の公開テストリソースの原形コピー。
取得元: https://github.com/batfish/batfish/tree/2f6128b8f113c6406254c1557daad0c2b05fb739/projects/batfish/src/test/resources/org/batfish/grammar

- Cisco IOS (`cisco`)、ASA、IOS XR、FortiOS、Juniper、Palo Alto を収録。
- FRR は標準的な BGP / OSPF / ルートポリシーの追加参照。
- `batfish/manifest.json` に全ファイルの出典パス・Git blob・SHA-256・サイズを記録。
- 上流の Apache-2.0 ライセンスは `batfish/LICENSE` を参照。第三者の著作・帰属を維持する。
- 更新・再取得: `python3 scripts/fetch-batfish-test-resources.py`。固定 revision を取得し内容を検証する。既存ファイルが相違する場合は上書きせず失敗する。
- `.test` は `*.test` 除外に対する限定例外。実設定をここへ置かないこと。

## 使い分け

原典は文法の検討・将来の手動比較専用。不正入力・警告検証・部分設定を含み、EtherScribe で全件成功することは期待しない。Batfish 自体の実行や解析結果との自動差分比較は行わない。

通常の Go テストは独立に作成した `core/testdata/corpus-*` のみを使用する。原典の設定を実機へ投入しない。公開原典に含まれるパスワード等の値もテストデータとして扱う。

## テストパターン

1. `corpus-fortios-complex.conf`: 2 VDOM、VDOM 別 IF・同名アドレス/サービス、CIDR 変換、FQDN、引用符付きグループ、VIP ポート変換、静的ルート、90→4 のポリシー順序。期待: IF 4 / ルート 2 / ポリシー 4 / アドレス 4 / NAT 2。
2. `corpus-junos-complex.set`: EX/ELS、2 ポートの LAG、3 VLAN/IRB、access/trunk、RSTP、BLUE/GREEN routing-instance、BGP。期待: IF 9 / VLAN 3 / LAG 1 / ルート 3。BGP 詳細の期待値拡充は今後の課題。
3. `corpus-paloalto-complex.xml`: 共有アドレス、2 vsys の同名オブジェクト、L3 IF、静的ルート、セキュリティルール順序、宛先 NAT。期待: IF 2 / ルート 1 / ポリシー 4 / アドレス 3 / NAT 2。
4. `corpus-cisco-ios.conf`: VRF / LAG / VLAN / ACL / BGP の将来用入力。現在は未対応エラーを期待。
5. `corpus-cisco-nxos.conf`: vPC / trunk / VRF / SVI の将来用入力。現在は未対応エラーを期待。
6. `corpus-paloalto-cli.set`: XML と異なる PAN-OS set 入力。現在は未対応エラーを期待。

実行: `cd core && go test ./...`。対応する 3 件について意味と件数を確認し、JSON / HTML / Excel の生成を検証する。

今後の拡張候補: Junos brace/set 同値、apply-groups / deactivate / delete、FortiOS rename / unset / append / move、不正終端、PAN-OS Panorama 継承・CLI 対応、Cisco IOS/NX-OS のモデル化。上流ファイルが存在することと、本ツールが対応することは別として管理する。

## 公式リファレンスに基づく追加回帰

`core/testdata/grounded-*` と `core/internal/pipeline/grounded_network_test.go` に複数 IP、BGP 継承、qualified-next-hop、LACP、IPv6、brace/set 同値を追加。出典と対応範囲は [公式対照記録](../docs/reference/official-parser-grounding.md) を参照。
