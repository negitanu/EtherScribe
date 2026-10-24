## ADDED Requirements

### Requirement: HA and logging sections in exports
システム SHALL プリセットまたはセクション指定で `ha` および `logging` を選んだとき、解析済みの HA／ログ設定を JSON・HTML・Excel に含めなければならない。

#### Scenario: Ops preset includes HA
- **WHEN** プリセット `ops` で JSON 出力し HA がモデルにある
- **THEN** 成果物に HA 情報が含まれる

### Requirement: Network and VPN export include extended fields
システム SHALL `network` セクションに DHCP・動的ルーティング・ポリシールートを、`vpn` セクションに SSL-VPN を含められるようにしなければならない。

#### Scenario: Network section with DHCP
- **WHEN** `network` セクション指定で JSON 出力し DHCP がモデルにある
- **THEN** 成果物に DHCP サーバー情報が含まれる
