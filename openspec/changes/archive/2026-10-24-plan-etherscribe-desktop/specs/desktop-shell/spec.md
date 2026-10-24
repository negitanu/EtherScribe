## ADDED Requirements

### Requirement: Cross-platform desktop application
システム SHALL Electron ベースのデスクトップアプリとして macOS、Windows、Linux で起動し、パラメータシート生成の主要操作を GUI から完了できるようにする。

#### Scenario: Launch on supported OS
- **WHEN** 利用者が対応 OS 上で配布物または開発起動でアプリを開く
- **THEN** メインウィンドウが表示され、設定ファイルの選択操作が可能である

### Requirement: Convert workflow in GUI
システム SHALL 設定ファイル選択、出力形式・プリセット選択、実行、成功時の保存（および HTML のプレビュー）を GUI で提供する。

#### Scenario: Successful convert and save
- **WHEN** 利用者が対応する FortiGate または Palo Alto 設定を選び HTML または Excel または JSON で変換を実行し保存先を指定する
- **THEN** 指定形式の成果物が保存され、失敗時は設定本文を含まないエラーを表示する

### Requirement: Compare workflow in GUI
システム SHALL 変更前・変更後の 2 ファイルを指定した差分生成を GUI から実行できるようにする。

#### Scenario: Successful compare
- **WHEN** 利用者が同一機器とみなせる 2 つの設定ファイルで差分を実行する
- **THEN** 差分レポートを生成し保存でき、進捗または完了状態を画面に示す

### Requirement: No silent persistence of configs
システム SHALL 入力設定と生成物をアプリ専用の永続ストアへ保存してはならない。利用者が選んだ出力パスへの書き込みと、ジョブ用の一時ファイル（終了時削除）のみ許可する。

#### Scenario: Temporary artifacts cleaned up
- **WHEN** 変換ジョブが成功または失敗で終了する
- **THEN** ジョブ用一時ファイルは削除され、アプリ設定領域に設定原文が残らない
