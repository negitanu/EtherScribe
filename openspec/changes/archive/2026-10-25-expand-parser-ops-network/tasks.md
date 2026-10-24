## 1. モデルと秘匿

- [x] 1.1 `model.Config` に HA / Logging / DHCP / Routing(OSPF・BGP) / PolicyRoute / SSLVPN を追加する
- [x] 1.2 秘匿: HA パスワード・SNMP コミュニティは生値をモデルに載せない（フラグまたは空）

## 2. FortiGate 解析（運用 → VPN → ネットワーク）

- [x] 2.1 HA（`system ha`）コンバータと fixture 検証
- [x] 2.2 Syslog / SNMP / FortiAnalyzer コンバータと fixture 検証
- [x] 2.3 SSL-VPN（settings + portal）を VPN 変換に追加
- [x] 2.4 DHCP サーバー変換
- [x] 2.5 OSPF / OSPF6 / BGP / ポリシールート変換
- [x] 2.6 既知ブロックを coverage 除外リストへ追加

## 3. Palo Alto

- [x] 3.1 NAT ルール MVP
- [x] 3.2 IPsec トンネル MVP
- [x] 3.3 fixture 更新とテスト

## 4. 出力と文書

- [x] 4.1 JSON / HTML / Excel で ha・logging・拡張 network・vpn を出力
- [x] 4.2 `wallscribe-essence.md` と essence fixture を更新
- [x] 4.3 `go test ./...` と fixture キー検査
