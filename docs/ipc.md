# EtherScribe IPC 契約

Electron Main と Go エンジン（`etherscribe rpc`）は **stdio 上の JSON-RPC 2.0（NDJSON）** で通信する。

## 輸送

- 1 行 1 JSON メッセージ
- リクエスト: `{ "jsonrpc":"2.0", "id": <number>, "method": "<name>", "params": { ... } }`
- 応答: `{ "jsonrpc":"2.0", "id": <number>, "result": ... }` または `error`
- 通知（id なし）: `{ "jsonrpc":"2.0", "method":"progress", "params": { ... } }`

## メソッド

| Method | Params | Result |
| --- | --- | --- |
| `engine.version` | （なし） | `{ name, version, formats, vendors, rpcMethods }` |
| `sections.list` | （なし） | `{ sections, presets }` |
| `convert.run` | `{ path, format, preset?, sections?, output? }` | `{ outputPath, format, summary }` |
| `compare.run` | `{ before, after, format, output? }` | `{ outputPath, format, summary }` |
| `cancel` | （なし） | `{ cancelled: true }` |

## 制約

- 成果物バイナリは RPC に載せない（パス参照）
- 応答・ログに設定全文や秘匿値を含めない
- `output` 未指定時はエンジンが一時ファイルを作りパスを返す（GUI は保存ダイアログで明示パスを渡す）

## Electron の出力選択とプレビュー

renderer は `sections.list` の `presets` / `sections` を表示し、`convert:run` に `path`, `format`, `preset`, `sections` を渡す。明示的なセクション選択はプリセットより優先する。空選択は GUI で実行不可とする。

生成・差分の成功応答は `outputPath`, `format`, `summary`。保存ダイアログの取消は `{cancelled: true}`。HTML プレビューの読み取りは同一アプリセッション内で生成した HTML パスに限定し、renderer の iframe は sandbox を使用する。
