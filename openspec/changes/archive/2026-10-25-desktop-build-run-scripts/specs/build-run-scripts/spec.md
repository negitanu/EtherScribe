## ADDED Requirements

### Requirement: ルートからデスクトップをビルドできる

システム SHALL リポジトリルートの `scripts/build.sh` により、Go エンジンを `desktop/resources/engine` へ同梱したうえで、Electron デスクトップ成果物を生成する。

#### Scenario: 既定ビルド（pack）

- **WHEN** 利用者がリポジトリルートで `./scripts/build.sh` を実行する
- **THEN** エンジンバイナリが `desktop/resources/engine` に配置され、`desktop/release/` に `electron-builder --dir` 相当の成果物が生成される

#### Scenario: パックなしビルド

- **WHEN** 利用者が `./scripts/build.sh --no-pack` を実行する
- **THEN** エンジン・renderer・Electron 本ビルドまで完了し、`electron-builder` は実行されない

### Requirement: ルートからデスクトップを起動できる

システム SHALL リポジトリルートの `scripts/run.sh` により、EtherScribe デスクトップを起動できるようにする。既定は開発モードとする。

#### Scenario: 開発起動

- **WHEN** 利用者が `./scripts/run.sh` を実行する
- **THEN** エンジンがビルドされ、Vite 開発サーバ経由で Electron ウィンドウが開く

#### Scenario: 本番相当起動

- **WHEN** 利用者が `./scripts/run.sh --prod` を実行する
- **THEN** 必要なビルドのあと、ビルド済み renderer を読み込む Electron プロセスが起動する

### Requirement: 既存エンジンビルド入口を維持する

システム SHALL 既存の `scripts/build-engine.sh` を維持し、`build.sh` / `run.sh` から再利用する。

#### Scenario: エンジン単体ビルド

- **WHEN** 利用者が `./scripts/build-engine.sh` を実行する
- **THEN** 現行どおり `desktop/resources/engine` にプラットフォーム向けバイナリが出力される
