## Why
WallScribe の解析機能は Go に移植されているが、Electron は最小フォームのままでスイッチプリセット・個別出力選択がなく、HTML/Excel にはモデルの欠落がある。起動から成果物の確認まで一貫した操作と台帳品質を提供する。

## What Changes
- 日本語の生成／差分ワークスペース、対応機器案内、用途別プリセット、個別セクション選択、状態と結果サマリー。
- エンジンの sections.list を選択肢の正本とし、スイッチ全プリセットを公開。
- HTML/Excel は秘匿処理済みの共通テーブルを利用し、選択項目と解析範囲を欠落なく出力。
- HTML の目次・印刷・空状態、Excel の見出し・列幅・折返し・固定行・フィルター。
- 保存取消、失敗、二重実行、過去プレビューの扱いを明確化。

## Capabilities
### Modified Capabilities
- `desktop-shell`: 用途別 GUI と状態管理。
- `parameter-sheets`: 読みやすい形式間共通の台帳。

## Impact
Electron renderer/main/preload、Go export、fixture 検証。Web/Flask の実行環境は移植しない。PDF・lab・ネットワーク調査・未対応パーサ拡張は既存の後続方針を維持する。
