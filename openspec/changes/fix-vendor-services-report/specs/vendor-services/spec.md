## ADDED Requirements

### Requirement: Vendor appropriate service output
出力 SHALL メーカーに属するサービス項目を表示し、FortiGate 専用項目を Palo Alto / Junos に表示してはならない。
#### Scenario: Palo Alto logging section
- **WHEN** Palo Alto を logging 節だけ、または全節で JSON/HTML/Excel 出力する
- **THEN** FortiAnalyzer の項目・見出しは現れず Syslog の設定有無が明示される
#### Scenario: FortiGate logging section
- **WHEN** FortiGate を出力する
- **THEN** FortiAnalyzer の項目は引き続き利用できる

### Requirement: Palo Alto DHCP service
パーサー SHALL DHCP client、server、relay を区別し、サーバーモード・全 IP プール・予約・リース・ゲートウェイ・マスク・DNS を保持する。
#### Scenario: Multiple pools and reserved clients
- **WHEN** 複数 IP プールと予約・disabled/auto モードを含む DHCP 設定がある
- **THEN** 全設定を失わずサーバーとして出力し、client-only IF はサーバーとして扱わない

### Requirement: Explicit syslog configuration presence
システム SHALL Syslog 定義の有無を未解析と区別して保持する。
#### Scenario: Shared and vsys server profiles
- **WHEN** shared と vsys に Syslog プロファイルがある
- **THEN** スコープ別の送信先と設定ありを出力し、転送有効と断定しない
#### Scenario: No syslog configuration
- **WHEN** Palo Alto 設定に Syslog 定義がない
- **THEN** 設定なしを明示する
