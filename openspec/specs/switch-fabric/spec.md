# switch-fabric Specification

## Purpose
スイッチ向け L2/L3/VRF の統一モデル語彙と、スイッチ専用プリセット／セクションを定義する。
## Requirements
### Requirement: Device role and virtual contexts
システム SHALL 統一モデル上でデバイス役割（少なくとも `firewall` と `switch`）および仮想コンテキスト（VDOM / vsys / routing-instance / VRF / default）を表現できなければならない。コンテキスト未指定のネットワーク要素は default コンテキストに属するものとして扱ってよい。

#### Scenario: Switch role with routing-instance
- **WHEN** スイッチ設定に routing-instance（または VRF）が含まれる
- **THEN** モデルに role=switch 相当と、当該インスタンスを表す仮想コンテキストが含まれる

### Requirement: L2 fabric fields
システム SHALL VLAN、LAG（集約リンク）、およびポートの L2 属性（アクセス／トランク、所属 VLAN、LAG 所属）を統一モデルに保持できなければならない。

#### Scenario: Access port and VLAN
- **WHEN** アクセスポートと VLAN ID 定義を解析してモデルに載せる
- **THEN** VLAN 一覧と、ポートのアクセス VLAN／モードを参照できる

### Requirement: L3 and VRF-scoped routing
システム SHALL IRB／SVI 等の L3 インターフェースおよび静的／動的ルーティングを仮想コンテキスト単位でモデルに保持できなければならない。

#### Scenario: Route in named VRF
- **WHEN** 名前付き VRF（routing-instance）内の静的ルートをモデル化する
- **THEN** 当該ルートにコンテキスト識別子が付き、default コンテキストのルートと区別できる

### Requirement: Switch-specific presets
システム SHALL スイッチ向けプリセット（少なくとも overview / l2 / l3 / ops / all 相当）を提供し、ファイアウォール専用セクション（ポリシー・NAT・VPN・セキュリティプロファイル）をこれらのプリセットに含めてはならない。

#### Scenario: switch-overview excludes firewall policies
- **WHEN** プリセット `switch-overview` で JSON 出力する
- **THEN** 機器・ネットワーク／VLAN 系は含まれ、firewallPolicies セクションはプリセット範囲外である
