# v1.0.0 デスクトップ配布

## Why
Windows x64 / macOS Apple Silicon の検証済み配布物を GitHub Release から入手できるようにする。

## What Changes
- 公開ソースと配布物に実設定・資格情報・ローカル環境が含まれないことを検査する。
- Go エンジンを CI 上でクロスコンパイルし、各 OS の Electron 配布物へ同梱する。
- Windows x64 ZIP / macOS arm64 ZIP・DMG と SHA-256 を生成する。
- アプリとエンジンを 1.0.0 に統一し、タグと版の不一致を拒否する。
- 両 OS の検証に成功してから v1.0.0 を公開する。
