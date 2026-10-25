# メーカー別サービス出力と Palo Alto DHCP / Syslog

## Why
Palo Alto の DHCP が未解析で、ログ・監視は FortiAnalyzer など他社専用項目を空欄表示してしまう。Syslog の設定有無も判別しづらい。

## What Changes
- PAN-OS XML の DHCP サーバー（モード、全プール、予約、リース、オプション）と IPv4 relay を抽出する。
- shared / vsys の Syslog プロファイルと送信先を抽出し設定有無を明示する。
- FortiAnalyzer と未対応の FortiOS 型 HA は FortiGate のみ出力する。既知の他社出力から項目を除く。
- 三形式のメーカー適合・有無・詳細値の合成回帰を追加する。
