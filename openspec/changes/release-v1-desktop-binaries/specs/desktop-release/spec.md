## ADDED Requirements

### Requirement: Clean distribution payload
配布物 SHALL アプリ実行に必要なファイルだけを含み、実設定・個人の環境ファイル・資格情報・テスト入力を含んではならない。
#### Scenario: Package inspection
- **WHEN** CI がパッケージを生成する
- **THEN** app.asar と外側の内容を検査し、許可リスト外の内容があれば公開しない

### Requirement: Verified platform builds
CI SHALL Windows x64 / macOS arm64 に対応する Go エンジンと Electron アプリを生成し、エンジンの版と RPC 起動を各対象 OS で確認する。
#### Scenario: Wrong platform engine
- **WHEN** エンジンの実行または版の確認に失敗する
- **THEN** ビルドを失敗させ Release を公開しない

### Requirement: Versioned downloadable release
リポジトリ SHALL v1.0.0 タグと一致する版の配布物・SHA-256 を GitHub Release へ公開する。
#### Scenario: Successful release
- **WHEN** 両 platform の検証が成功する
- **THEN** 配布物を draft へ全件アップロードした後に公開する
#### Scenario: Version mismatch
- **WHEN** タグとアプリまたは Go エンジンの版が不一致である
- **THEN** 公開前に CI を失敗させる
