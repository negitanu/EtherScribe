# config-diff Specification

## Purpose
TBD - created by archiving change plan-etherscribe-desktop. Update Purpose after archive.
## Requirements
### Requirement: Two-generation compare
システム SHALL 同一機器の変更前・変更後設定を受け取り、追加・削除・項目単位の変更を示す差分レポートを生成できなければならない。

#### Scenario: Detect added policy
- **WHEN** 変更後にのみ存在するポリシーがある 2 ファイルを比較する
- **THEN** 差分レポートにそのポリシーが追加として示される

### Requirement: Diff output formats
システム SHALL 差分レポートを少なくとも JSON および HTML で出力でき、Excel も提供する方針とする（Excel は MVP 完了条件に含めてよい）。

#### Scenario: Diff as JSON
- **WHEN** 比較結果を JSON で出力する
- **THEN** 機械可読な変更一覧が得られ、秘匿生値を含まない

### Requirement: Incompatible pair handling
システム SHALL 機器種別が一致しない、または比較不能な入力ペアを拒否しなければならない。

#### Scenario: FortiGate compared with Palo Alto
- **WHEN** FortiGate 設定と Palo Alto 設定を差分対象として渡す
- **THEN** 比較を行わずエラーを返す

