## ADDED Requirements
### Requirement: Consistent readable report tables
システム SHALL 選択したセクションの秘匿済み JSON に含まれるモデル項目を HTML と Excel でも出力し、解析範囲を常に付録に含める。
#### Scenario: Switch report parity
- **WHEN** Junos EX fixture を switch-all で出力する
- **THEN** HTML/Excel に VLAN、LAG、STP、コンテキスト付きルート、L2 ポート属性が含まれる
#### Scenario: Explicit section selection
- **WHEN** vlans のみを選んで出力する
- **THEN** 機器概要やポリシーを追加せず VLAN と解析範囲を出力する
### Requirement: Report presentation
システム SHALL HTML に章目次・印刷レイアウト・空項目の案内を、Excel に見出しスタイル・列幅・折返し・固定行・フィルターを提供する。
#### Scenario: Offline report
- **WHEN** HTML または Excel をネットワーク接続なしで開く
- **THEN** 外部資源なしで全項目を閲覧できる
