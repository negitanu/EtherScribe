## ADDED Requirements
### Requirement: Semantic collection tables
システム SHALL コレクションを意味別の固定列へ出力し、各オブジェクトを1行にまとめる。HTML と Excel の項目名と値の対応は一致し、内部 JSON パスや配列添字を利用者向けの項目名として表示してはならない。
#### Scenario: Address objects
- **WHEN** アドレス・サービス・グループを含む設定を出力する
- **THEN** それぞれ独立した表に、オブジェクト名／タイプ／値／説明、グループ名／メンバー等の日本語列で表示される
### Requirement: Nested record association
システム SHALL VPN・ルーティング・HA・監視等の子レコードを別表にし、親名・コンテキストと関連付ける。既存の選択項目・秘匿処理・解析範囲を維持する。
#### Scenario: BGP neighbors
- **WHEN** 別コンテキストの BGP にネイバーがある
- **THEN** ネイバーの行に親の識別情報とコンテキストが表示され、異なる親のレコードが混同されない
### Requirement: Stable export schema
システム SHALL データの並び・任意フィールドの有無に依存しない列順と日本語項目名を提供し、Excel シート名は有効かつ一意にする。
#### Scenario: Long context or table names
- **WHEN** 長い名前やシート名の使用禁止文字を含むコンテキストがある
- **THEN** Excel のシート名は31文字以内で一意となり、セルの元の定義名は保持される
