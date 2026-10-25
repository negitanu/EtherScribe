## ADDED Requirements
### Requirement: WallScribe compatible object semantics
システム SHALL 引用符内の空白・日本語・エスケープを含む名前を保持し、IPv4/IPv6 の subnet・FQDN・範囲、グループのメンバー、説明とタグを正しく解析する。
#### Scenario: Quoted FortiGate object names
- **WHEN** 名前に空白を含む edit と、その名前を含む member がある
- **THEN** 名前を途中で切らず、参照先と同一の文字列でモデル化する
### Requirement: Complete service and shared object parsing
システム SHALL FortiGate の複数プロトコルとポートを保持し、Palo Alto の shared/vsys のアドレス・サービス・グループを区別して取り込む。
#### Scenario: TCP and UDP service
- **WHEN** 同一サービスに TCP と UDP のポートがある
- **THEN** どちらも失わず、プロトコルと各ポートの対応が分かるモデルへ変換する
#### Scenario: Shared and dynamic address groups
- **WHEN** Palo Alto の shared に定義があり vsys に動的アドレスグループがある
- **THEN** shared のコンテキスト、動的フィルター、タグを保持する
