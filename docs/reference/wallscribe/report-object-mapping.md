# WallScribe 互換のオブジェクト解析・出力対照

確認日: 2026-10-25

正本は本ディレクトリの `SPECIFICATION.md`（2.2、3.1、3.3、3.5）。列名と変換の補足として以下を読み取り参照した。隣接リポジトリの変更や実設定の取り込みは行っていない。

- WallScribe `exporters/html_parts/vdom.py`: オブジェクトごとに1行、アドレス・サービス・グループを別表で表示。
- WallScribe `exporters/excel_parts/vdom_sheets.py`: アドレスの列は「オブジェクト名」「タイプ」「値」「説明」。グループは「グループ名」「メンバー」「説明」「動的フィルター」。サービスは「オブジェクト名」「プロトコル」「ポート」「説明」。
- WallScribe `parsers/fortigate/converters/objects.py`: IPv4/IPv6、subnet の CIDR、FQDN、IP 範囲、agent-id フォールバック、グループ・タグ、TCP/UDP/SCTP/ICMP の正規化。
- WallScribe `parsers/paloalto.py`: shared と vsys の区別、subnet/iprange/fqdn の共通語彙、タグ、動的グループの filter、サービスの送信元ポートと説明。

## EtherScribe への適用

- HTML/Excel 共通でアドレス・アドレスグループ・サービス・サービスグループを独立した表にする。
- アドレスは「オブジェクト名／タイプ／値／説明／タグ／VDOM・vsys」の固定列。内部の `addresses[1].name` 等は列名・項目名にしない。
- グループではメンバー配列を同じセルへ結合し、名前と参照の文字列を変えない。動的グループはメンバーが空でもフィルターを表示する。
- サービスは WallScribe の共通プロトコル／ポート表現に加え、TCP・UDP・SCTP の個別ポートを保持し、プロトコルとポートの対応を確認できるようにする。
- 全表に元の VDOM/vsys またはコンテキストを残し、同名の別スコープの定義を混ぜない。
- VPN、システム管理者、動的ルーティング、監視、HA、セキュリティの子リストも別表にし、繰り返しレコードの子に親設定と親コンテキストを表示する。
- 全項目のラベルを日本語辞書で管理し、型定義に従って列順を固定する。新しいモデルフィールドにラベルがなければテストで検出する。
- JSON の既存キーは維持。タグ、説明、動的フィルター、個別ポート等の追加キーは省略可能とする。Palo Alto のアドレス種別は `ip-netmask` → `subnet`、`ip-range` → `iprange`、サービスプロトコルは `tcp` / `udp` → `TCP` / `UDP` に正規化する。これは WallScribe と揃えるための既存値の修正であり、JSON の利用者もこれらの値を使用する。
- 秘匿済み JSON 値を表示に使い、文字列値を HTML エスケープする。解析警告の本文は出さず件数のみ表示する。

## 検証データ

- `core/testdata/fortigate-object-parity.conf`: 引用符・空白・日本語名、同名の別 VDOM、再編集、IPv4/IPv6、FQDN/range、タグ、メンバーの append/unselect、TCP+UDP+SCTP、ICMP。
- `core/testdata/paloalto-object-parity.xml`: shared/vsys の同名定義、IPv6、タグ、静的／動的グループ、送信元ポートと説明。

WallScribe 全機能との完全一致を示すものではない。PDF、HA 複数ファイル統合、未実装の認証・スケジュール等の後続領域と、VDOM ごとのシート配色・高度な参照ツールチップは別途実装する。
