# 設計

Batfish は参照コーパスとして `.test/batfish/grammar/` に配置する。原典には不正構文や秘密のダミー値も含むため通常の Go テストから読み込まない。取得スクリプトは固定 revision の Git tree から選択ディレクトリを列挙し、ダウンロード後の Git blob hash を検証する。manifest には出典パス・サイズ・SHA-256 を保存する。既存ファイルと一致しない再取得は失敗する。

独自 fixture は予約済みドキュメント用 IP と AS 番号を使う。FortiOS は多 VDOM・重複名・順序・グループ・NAT、Junos EX は LAG/VLAN/IRB/VRF/BGP、Palo Alto XML は vsys/共有オブジェクト/IF/NAT を組み合わせる。Cisco IOS/NX-OS と PAN-OS set は将来対応向け負例として明示する。出力を三形式で検証し、通常テストは合成 fixture に限定する。
