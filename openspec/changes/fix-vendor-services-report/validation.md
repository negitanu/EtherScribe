# 検証（2026-10-25）

- `cd core && go test ./...`: 全成功。
- `./scripts/check-fixture-keys.sh`: 3 機種成功。
- `./scripts/build-engine.sh`: 成功。
- `openspec validate fix-vendor-services-report --strict`: 成功。
- `git diff --check`: 成功。

## 合成回帰

- Palo Alto DHCP サーバー 3 件（auto/disabled/enabled）、client-only IF の除外、複数プール・CIDR、予約/除外 IP、リース timeout/unlimited、probe-ip true/false、DNS inherited、IPv4 relay の 2 送信先。
- shared/vsys 各 1 件の Syslog、送信方式・ポート・形式・facility、設定なしとの区別。
- JSON/HTML/Excel の全節・ログのみ・ネットワークのみで確認。他社 FortiAnalyzer 値を意図的にモデルへ混ぜても Palo Alto 出力へ現れない。
- FortiGate では FortiAnalyzer と有効/無効 Syslog を保持。Junos は今回 Syslog 未解析のため、設定なしと断定せず未解析と表示する。
- 未対応 DHCP オプション・IPv6 relay・Syslog カスタム format の coverage に実値を含めない。

実設定・実機での動作検証、Syslog 転送先への通信や転送有効化は行っていない。コミットは行っていない。
