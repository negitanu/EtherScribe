# 設計

公式 command reference の構文と意味を docs/reference/official-parser-grounding.md に記録し、取得日と回帰 fixture を紐付ける。全文コピーはしない。

共通モデルは加算的拡張とし、従来の ipAddress は先頭アドレス、addresses は全アドレスとする。routingContext は IF の vsys/VDOM と別に VRF/virtual-router を保持する。経路は action / priority / description、BGP peer は group / localAs / importPolicies / exportPolicies を追加する。

Junos は設定文から BGP の global/group/neighbor を集め、個別 > group > global の順で解決する。peer-as は RemoteAS にのみ使用する。routing-options autonomous-system と router-id はコンテキストから引き継ぐ。静的ルートは prefix/context と next-hop の組み合わせで統合し、qualified-next-hop の属性を優先する。LACP は ae の明示設定でのみ有効とする。deactivate は祖先パスも含めて無効化し、delete は未適用として値を含まない coverage に示す。

FortiOS は明示値を抽出する。primary/secondary IPv4 と IPv6、aggregate member を保持し、route priority と metric を混同しない。PAN-OS は IPv6 アドレスと virtual-router interface を XML から抽出する。既存のレポート生成は型から列を構築するため項目名と三形式の値保持を検証する。

Junos brace 入力は引用符とコメントを尊重した区切り走査で正規化し、一行内の複数階層と inactive: を display set 相当の文へ変換する。
