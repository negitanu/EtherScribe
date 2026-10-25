## Evidence
正本は docs/reference/wallscribe/SPECIFICATION.md。補足として隣接 WallScribe の exporters/html_parts/vdom.py、exporters/excel_parts/vdom_sheets.py、parsers/fortigate/converters/objects.py、parsers/paloalto.py を読み取り参照し、項目の対照を docs/reference/wallscribe/report-object-mapping.md に記録する。外部リポジトリは変更しない。

## Decisions
- JSON と同じ秘匿処理済み値を使用するが、内部パスの flatten を廃止する。型定義に基づく固定列と明示的な日本語ラベルで、配列を別表へ正規化する。データの有無や Go map の順序で列が変わらない。
- オブジェクトは WallScribe の「オブジェクト名／タイプ／値／説明」、グループは「グループ名／メンバー／説明／動的フィルター」を軸とし、タグ・VDOM/vsys も残す。
- 子リストは親名・VDOM/コンテキストを付けた行にし、配列添字を列名や項目名にしない。表示値の有効／無効・設定あり／なしは項目に合わせる。機器の識別子・IP/FQDN・ユーザーの定義名は翻訳しない。
- FortiGate サービスの Protocol/Port は WallScribe と同じ結合表現とし、TCP/UDP/SCTP 個別ポートも additive に保持して対応関係を失わない。IPv6 も既存オブジェクト配列に載せる。
- HTML/Excel は同じ意味別表を使用。Excel は使用禁止文字・31文字制限・重複するシート名を処理する。

## Validation
空白・日本語・引用符・同名の別VDOM、subnet/FQDN/range/IPv6、TCP+UDP/ICMP、shared/vsys、タグ・動的フィルターの合成 fixture。期待する列と行を具体値で照合する。全モデルフィールドの日本語ラベルと出力保持、空状態、親子対応、HTML エスケープ、Excel XML とシート名、既存全テスト・ビルド・画面検証を実施する。
