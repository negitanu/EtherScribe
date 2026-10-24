## ADDED Requirements
### Requirement: Task oriented workspace
システム SHALL 日本語の生成／差分ワークスペースと対応機器案内を表示し、HTML を初期出力形式とする。
#### Scenario: Initial workspace
- **WHEN** アプリを起動する
- **THEN** 生成画面、対応形式、ファイル選択、出力設定、未生成の案内が表示される
### Requirement: Engine supplied output choices
システム SHALL sections.list から全プリセットとセクションを表示し、個別選択を変換要求へ渡す。
#### Scenario: Switch selection
- **WHEN** switch-l2 を選択する
- **THEN** network/vlans/lag/stp/device_info が選択され、個別変更して保存できる
### Requirement: Clear job outcomes
システム SHALL 実行中の重複操作を防止し、保存取消・失敗・成功を日本語で通知する。成功時は出力先と件数を示し、生成と差分の HTML を sandbox 内でプレビューする。
#### Scenario: Save cancelled
- **WHEN** 保存ダイアログを取消する
- **THEN** 取消が表示され実行可能な状態へ戻り古いプレビューは表示されない
