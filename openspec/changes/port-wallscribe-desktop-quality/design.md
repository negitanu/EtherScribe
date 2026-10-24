## Context
参照元は docs/reference/wallscribe/SPECIFICATION.md の 3 章（出力、プリセット、解析範囲）、6 章（変換と差分の操作）。Electron/stdio RPC と現在の Go パーサは維持する。

## Decisions
- renderer は既存 TypeScript と CSS で構築。アプリ内の設定保存は追加しない。
- sections.list のキー・プリセットを GUI に供給し、明示的な sections を IPC から convert.run に渡す。空選択は GUI で拒否。
- JSON の既存 scrub 処理を経た値を共通テーブルへ変換。配列はレコード行、ネストはフィールドパスと値として保持し、未実装の表示項目による欠落を防ぐ。主要見出しは日本語にする。
- HTML はスクリプト不要の目次とレスポンシブ表・印刷 CSS。プレビュー iframe は sandbox を設定。
- Excel は既存の依存なし OOXML を維持し、全セルを文字列として扱う。
- プレビューの読み取りはそのセッションで生成した HTML のみに制限する。

## Validation
FortiGate/Palo Alto/Junos EX の合成 fixture、出力項目の選択と秘匿、HTML エスケープ、XLSX XML 整合性と表示属性をテスト。Go 全テスト、TypeScript、製品ビルド、OpenSpec strict 検証を実行する。
