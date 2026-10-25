# WallScribe から EtherScribe へ移植したエッセンス

隣接の WallScribe リポジトリがなくても作業できるよう、必要な参照とロジックを本リポジトリへ取り込んだ記録。

## 同梱参照（読み取り専用メモ）

| パス | 由来 |
| --- | --- |
| `docs/reference/wallscribe/SPECIFICATION.md` | 出力・抽出項目の正本コピー |
| `docs/reference/wallscribe/textfsm_templates/` | 機器識別・ヘッダの TextFSM |
| `docs/reference/wallscribe/device_identification.py` | 識別ロジック原典 |
| `docs/reference/wallscribe/redaction.py` | 秘匿方針原典 |
| `docs/reference/wallscribe/sections.py` | プリセット／セクション原典 |
| `core/testdata/*.conf` / `*.xml` | 合成 fixture（WallScribe examples 由来） |

## Go 実装へ移植済み

- エンコーディング順: utf-8-sig → utf-8 → cp932 → latin-1
- 機器識別正規表現（TextFSM 相当）と FortiGate 表示モデル名変換
- `#config-version` フォールバック、`:opmode=` / `:vdom=`
- FortiGate 設定ツリー（config / edit / set / append / next / end）と引用符結合
- `ip` + マスク → CIDR
- 秘匿フィールド名と `ENC …` 伏せ字
- セクション／プリセット名（overview / connectivity / policy / ops / all）
- FortiGate: 機器概要、IF、ルート、ポリシー、オブジェクト、NAT（VIP/IP Pool）、IPsec Phase1/2、SSL-VPN、セキュリティプロファイル（AV / Web / App / IPS / SSL。App ID は同梱カタログで解決し、未知 ID は ID のまま）、HA、Syslog/SNMP/FortiAnalyzer、DHCP、OSPF/OSPF6、BGP、ポリシールート
- Palo Alto XML: ホスト、IF（ゾーン・vsys）、ルート、vsys ポリシー、vsys アドレス／サービス、NAT ルール、IKE ゲートウェイ
- モデル一般化: Device Role、VirtualContext、VLAN / LAG / STP、Interface L2 属性、Routing Context（Cisco 追加を見据えた additive 拡張）
- Junos EX（ELS）: brace + display set 正規化、ホスト／IF／VLAN／LAG／STP／IRB／routing-instance／静的・BGP（詳細は `docs/reference/junos-ex.md`）
- スイッチ専用プリセット: `switch-overview` / `switch-l2` / `switch-l3` / `switch-ops` / `switch-all`
- 差分: ポリシー / IF / ルートの追加・削除・変更
- 秘匿: HA パスワード・SNMP コミュニティ・IPsec PSK は生値をモデルに載せない（設定有無フラグ）

## 項目名とオブジェクト出力の追加移植

- `report-object-mapping.md`（`docs/reference/wallscribe/`）に原典との対照を記録。
- オブジェクトは名前・タイプ・値・説明を固定列で表示し、グループ・サービスを別表に分離。
- FortiGate の引用符付き名前、IPv6 オブジェクト／グループ、タグ、複数プロトコルのサービスと説明。
- Palo Alto の共有オブジェクト、タグ、動的アドレスグループ、送信元ポートと説明、アドレス種別とプロトコルの共通表記。
- VPN／ルーティング／HA／監視／セキュリティの子レコードは親設定・コンテキスト付きの別表として出力。

## 実設定の構造検証による追加改善

- FortiGate の VDOM 暗黙終端と評価順の保持、後続 VDOM の読み落とし修正。
- Palo Alto の集約・VLAN・論理／サブ IF、L2/L3/tap/virtual-wire、管理 IP/DNS/NTP、ポリシーのアプリ・ユーザー・カテゴリ・プロファイル。
- ソース構造との独立した件数・順序照合、未対応構造の付録。実設定は非追跡・匿名検証とし、回帰には合成 fixture のみを使う。

## 意図的に後続へ残すもの

- 構築テスト（lab）、ネットワーク調査、PDF、HA クラスタ複数ファイル結合の高度解析
- Palo Alto のセキュリティプロファイル、GlobalProtect 詳細、IKE crypto profile 詳細展開
- Junos QFX / SRX、非 ELS、Virtual Chassis / EVPN
- Cisco IOS / NX-OS スイッチパーサ（モデル語彙は先行整備済み）
- FortiGate 公式アプリカタログ全件（`core/internal/appcatalog/appid.csv` はフィクスチャ用の合成行のみ）
- SD-WAN / スケジュール / ユーザー認証サーバー（WallScribe extended）
- TextFSM ランタイム依存（Go では正規表現へ落とし込み済み）
- WallScribe Web / Flask API
