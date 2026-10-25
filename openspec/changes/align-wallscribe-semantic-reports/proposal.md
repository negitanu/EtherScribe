## Why
汎用 JSON flatten がオブジェクト名・タイプ・値を内部パスごとの行に分解し、WallScribe と異なる読めない台帳になっている。FortiGate の引用符付き edit 名切断、サービスの情報欠落、Palo Alto の共有定義未取り込みも修正する。
## What Changes
- アドレス／グループ／サービス等のコレクションごとに日本語の固定列・1定義1行の表を生成する。
- ネストした VPN・動的ルーティング・HA・監視・セキュリティ詳細を別表へ分離し、親名・コンテキストとの対応を保持する。
- FortiGate の引用符・タブ・空文字・再編集、IPv6 アドレス／グループ、タグ、サービスの複数プロトコル・説明を移植する。
- Palo Alto の shared オブジェクト、共通のアドレス種別、タグ、動的グループ、サービス説明・送信元ポートを移植する。
- 合成 fixture による値と表の対応、秘匿、JSON 互換性、Excel シート名の検証を追加する。
## Capabilities
### Modified Capabilities
- `parameter-sheets`: 意味別テーブル、項目名、親子関係。
- `vendor-parsing`: オブジェクト解析の WallScribe 互換。
## Impact
Go モデル・FortiGate/Palo Alto パーサ・HTML/Excel 共通表。JSON の既存キーを維持し新情報を additive に追加する。パース誤りの名前・種別・サービス値は正規化する。コミット／push は今回行わない。
