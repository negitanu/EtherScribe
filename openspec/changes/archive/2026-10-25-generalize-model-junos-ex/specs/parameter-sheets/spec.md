## ADDED Requirements

### Requirement: Switch preset export
システム SHALL スイッチ専用プリセットまたはスイッチ向けセクション指定で出力するとき、VLAN／LAG／L2 ポート／VRF 付きルーティングなど解析済みのスイッチ関連項目を JSON・HTML・Excel に含めなければならない。ファイアウォール専用プリセットの既存挙動を壊してはならない。

#### Scenario: switch-l3 includes VRF routes
- **WHEN** プリセット `switch-l3` で JSON 出力し VRF 付きルートがモデルにある
- **THEN** 成果物にコンテキスト付きルーティング情報が含まれる

#### Scenario: Legacy policy preset unchanged
- **WHEN** プリセット `policy` で FortiGate モデルを JSON 出力する
- **THEN** 従来どおりポリシー監査向けセクションが含まれ、スイッチ専用セクションの強制追加はない
