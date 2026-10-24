# WallScribe 仕様書

**バージョン 1.4**  
**最終更新日: 2026-04-01**

`docs/` 配下の仕様・設計の正本は本ファイルです。クイックスタートや日々の操作はリポジトリ直下の [README.md](../README.md) を参照してください。Docker イメージには `docs/` は含めません（[.dockerignore](../.dockerignore)）。

## 目次

| 章 | 内容 |
| --- | --- |
| [1. 概要](#1-概要) | 目的、対象機器、出力形式 |
| [2. 機能要件](#2-機能要件) | 入力処理、抽出項目 |
| [3. 出力仕様](#3-出力仕様) | HTML / PDF / Excel / JSON、出力項目とプリセット、付録「解析範囲」、差分レポート |
| [4. アーキテクチャ設計](#4-アーキテクチャ設計) | モジュール構成、機器識別（TextFSM）、クラス設計、処理フロー |
| [5. CLI 仕様](#5-cli-仕様) | コマンドライン |
| [6. Web インターフェース仕様](#6-web-インターフェース仕様) | UI・API・運用 |
| [7. 技術仕様](#7-技術仕様) | 動作環境、依存パッケージ、エンコーディング |
| [8. エラーハンドリング](#8-エラーハンドリング) | 想定エラー、ログ |
| [9. 制限事項](#9-制限事項) | 取得不可情報、バージョン、性能 |
| [10. 今後の拡張予定](#10-今後の拡張予定) | ロードマップ |
| [11. 付録](#11-付録) | 参照先、更新履歴 |

---

## 1. 概要

### 1.1 目的

FortiGate と PA-Series ファイアウォールの設定ファイルから、統一フォーマットのパラメータシートを自動生成するツール。

### 1.2 対象機器

| ベンダー           | 機種                    | 設定ファイル形式       |
| ------------------ | ----------------------- | ---------------------- |
| Fortinet           | FortiGate シリーズ      | `.conf` (テキスト形式) |
| Palo Alto Networks | PA シリーズ (PA-440 等) | `.xml` (XML 形式)      |

### 1.3 出力形式

- **HTML**: スタイル付きの見やすいレポート（印刷対応、インタラクティブツールチップ付き）
- **PDF**: 印刷に最適化されたPDF形式（章がページをまたがない）
- **Excel (.xlsx)**: 編集可能なスプレッドシート形式
- **JSON**: 機械可読形式（CI でのレビュー・他ツール連携向け。秘匿値は出力しない）
- **差分レポート**: 同一機器の 2 世代を比較した変更一覧（HTML / PDF / Excel / JSON）

いずれの形式も、台帳に反映できなかった設定を付録「解析範囲」（JSON では `coverage`）に載せる。

---

## 2. 機能要件

### 2.1 入力処理

#### 2.1.1 FortiGate 設定ファイル (.conf)

- FortiOS 6.x / 7.x / 8.0 形式に対応
- VDOM（Virtual Domain）構成に対応
- エンコーディングは自動検出（UTF-8 / UTF-8 BOM / CP932(Shift-JIS) / Latin-1）

#### 2.1.2 Palo Alto 設定ファイル (.xml)

- PAN-OS 10.x / 11.x 形式に対応
- vsys（Virtual System）構成に対応
- エンコーディングは自動検出（UTF-8 / UTF-8 BOM / CP932(Shift-JIS) / Latin-1）

### 2.2 抽出項目

両機種で共通の情報を抽出し、統一フォーマットで出力する。

#### カテゴリ 1: 機器概要

| 項目          | FortiGate                  | Palo Alto                      | 説明                       |
| ------------- | -------------------------- | ------------------------------ | -------------------------- |
| ホスト名      | `system global > hostname` | `deviceconfig/system/hostname` | 機器名                     |
| モデル名      | `#config-version` 等（TextFSM 含む） | 表示上は `PA Series`（XML からの機種要素は今後拡張可） | 機種名 |
| OS バージョン | ヘッダから抽出             | ルート属性 `detail-version` 優先、なければ `version`（TextFSM 補助） | ファームウェアバージョン   |
| シリアル番号  | -                          | -                              | 設定ファイルからは取得不可 |
| 動作モード    | `opmode`                   | -                              | NAT/Route, Transparent     |

#### カテゴリ 2: システム設定

| 項目             | FortiGate          | Palo Alto                        | 説明                |
| ---------------- | ------------------ | -------------------------------- | ------------------- |
| 管理 IP          | `system interface` | `deviceconfig/system/ip-address` | 管理用 IP アドレス  |
| 許可プロトコル   | `allowaccess`      | `service`                        | HTTPS, SSH, Ping 等 |
| NTP サーバー     | `system ntp`       | `ntp-servers`                    | 時刻同期先          |
| DNS サーバー     | `system dns`       | `dns-setting/servers`            | 名前解決先          |
| タイムゾーン     | `timezone`         | `timezone`                       | タイムゾーン設定    |
| 管理者アカウント | `system admin`     | `mgt-config/users`               | 管理ユーザー一覧    |

#### カテゴリ 3: ネットワーク設定

| 項目                  | FortiGate                 | Palo Alto                                      | 説明                   |
| --------------------- | ------------------------- | ---------------------------------------------- | ---------------------- |
| 物理インターフェース  | `system interface`        | `network/interface/ethernet`                   | 物理ポート設定         |
| VLAN インターフェース | `system interface (vlan)` | `network/interface/vlan`                       | VLAN 設定              |
| LAG インターフェース  | `system interface (lag)`  | `network/interface/aggregate-ethernet`         | リンクアグリゲーション |
| ゾーン                | -                         | `zone`                                         | セキュリティゾーン     |
| スタティックルート    | `router static`           | `virtual-router/routing-table/ip/static-route` | ルーティング設定       |
| OSPF / OSPFv6         | `router ospf` / `router ospf6` | `network/virtual-router/protocol/ospf`      | 動的ルーティング（OSPF） |
| BGP                   | `router bgp`              | `network/virtual-router/protocol/bgp`         | 動的ルーティング（BGP） |
| ポリシールート        | `router policy`           | `network/virtual-router/routing-table/ip/route` | ポリシーベースルーティング |
| DHCP サーバー         | `system dhcp server`      | `network/dhcp`                                 | DHCP 設定              |

#### カテゴリ 4: オブジェクト定義

| 項目                 | FortiGate                 | Palo Alto       | 説明                  |
| -------------------- | ------------------------- | --------------- | --------------------- |
| アドレスオブジェクト | `firewall address`        | `address`       | IP アドレス/FQDN 定義 |
| アドレスグループ     | `firewall addrgrp`        | `address-group` | アドレスのグループ化  |
| サービスオブジェクト | `firewall service custom` | `service`       | ポート/プロトコル定義 |
| サービスグループ     | `firewall service group`  | `service-group` | サービスのグループ化  |

#### カテゴリ 5: ファイアウォールポリシー

| 項目                           | FortiGate                         | Palo Alto                    | 説明           |
| ------------------------------ | --------------------------------- | ---------------------------- | -------------- |
| セキュリティポリシー           | `firewall policy`                 | `security/rules`             | FW ルール      |
| Internet Service               | `internet-service-name`           | -                            | FortiGate 固有 |
| NAT ポリシー                   | `firewall vip`, `firewall ippool` | `nat/rules`                  | SNAT/DNAT 設定 |
| Local-in ポリシー              | `firewall local-in-policy`        | -                            | FortiGate 固有 |
| アプリケーションオーバーライド | -                                 | `application-override/rules` | PA 固有        |

#### カテゴリ 6: VPN 設定

| 項目      | FortiGate                 | Palo Alto                       | 説明                 |
| --------- | ------------------------- | ------------------------------- | -------------------- |
| IPsec VPN | `vpn ipsec phase1/phase2` | `network/tunnel/ipsec`          | Site-to-Site VPN     |
| SSL-VPN   | `vpn ssl settings`        | `network/tunnel/global-protect` | リモートアクセス VPN |

#### カテゴリ 7: セキュリティプロファイル

| 項目                       | FortiGate                  | Palo Alto                | 説明               |
| -------------------------- | -------------------------- | ------------------------ | ------------------ |
| アンチウイルス             | `antivirus profile`        | `profiles/virus`         | AV 設定            |
| Web フィルタ               | `webfilter profile`        | `profiles/url-filtering` | URL フィルタリング |
| アプリケーションコントロール | `application list`         | `profiles/application`   | アプリ制御         |
| IPS                        | `ips sensor`               | `profiles/vulnerability` | 侵入防御           |
| SSL インスペクション       | `firewall ssl-ssh-profile` | `profiles/decryption`    | SSL 復号設定       |

##### アプリケーションID解決機能

FortiGateのアプリケーションコントロールプロファイルでは、アプリケーションIDが使用されます。
本ツールでは`static/data/appid.csv`を参照し、IDからアプリケーション名・カテゴリ・リスクレベルを自動解決します。

| フィールド | 説明 |
| ---------- | ---- |
| app_id | アプリケーションID（数値） |
| app_name | アプリケーション名（例: Facebook, Skype） |
| category | カテゴリ（例: Social.Media, Collaboration） |
| risk | リスクレベル（1-5） |
| description | 説明 |

##### Internet Service対応

FortiGateのファイアウォールポリシーで使用されるInternet Service（ISDB）に対応しています。
`internet-service-name`で指定されたサービス名を、`static/data/appid.csv`から解決し、アプリケーション名として表示します。

- Internet Service名はバッジ形式で表示
- ツールチップでサービス名、ISDB ID、アプリケーション名を表示
- 通常のアドレスオブジェクトと区別して表示

#### カテゴリ 8: 高可用性（HA）設定

| 項目                 | FortiGate          | Palo Alto                      | 説明               |
| -------------------- | ------------------ | ------------------------------ | ------------------ |
| HA モード            | `system ha > mode` | `high-availability/group/mode` | A-P / A-A          |
| グループ ID          | `group-id`         | `group-id`                     | HA グループ識別子  |
| 優先度               | `priority`         | `election-option/priority`     | マスター選出優先度 |
| HA管理ステータス     | `ha-mgmt-status`   | -                              | HA管理インターフェースの状態 |
| 監視インターフェース | `monitor`          | `link-monitoring`              | 障害検知対象       |

#### カテゴリ 9: ログ・監視設定

| 項目                   | FortiGate             | Palo Alto                  | 説明         |
| ---------------------- | --------------------- | -------------------------- | ------------ |
| Syslog 転送            | `log syslogd setting` | `syslog`                   | ログ転送先   |
| SNMP                   | `system snmp`         | `snmp-trap-server`         | 監視設定     |
| FortiAnalyzer/Panorama | `log fortianalyzer`   | `panorama/panorama-server` | 集中管理連携 |

---

## 3. 出力仕様

### 3.1 HTML 出力

#### 3.1.1 レイアウト構成

```text
[ヘッダー]
  - ドキュメントタイトル
  - 機器名、OSバージョン、作成日

[目次]
  - 各セクションへのページ内リンク

[セクション1: 機器概要]
  - 基本情報（ホスト名、モデル名、シリアル番号、OSバージョン、動作モード、VDOM/vsys）
  - ライセンス情報（サポート、UTM Bundle、AV、Web、IPS有効期限）
  - サマリー（インターフェース数、ルート数、ポリシー数、オブジェクト数、HA状態）

[セクション2: システム設定]
  - 管理アクセス（管理用IPアドレス/サブネットマスク、管理インターフェース、許可プロトコル、HTTPS/SSHポート）
  - 基本設定（DNS、NTP、タイムゾーン）
  - 管理者アカウント（ユーザー名、権限プロファイル、VDOM、信頼ホスト）

[セクション3: ネットワーク設定]
  - インターフェース（物理・VLAN・LAG: 名前、タイプ、役割、IPアドレス、VLAN ID、ゾーン、VDOM、許可アクセス）
  - スタティックルート（名前、宛先ネットワーク、ゲートウェイ、デバイス、VDOM）
  - OSPF/OSPFv6設定（Router ID、エリア設定、インターフェース設定、再配布設定）
  - BGP設定（AS番号、Router ID、ネイバー設定、ネットワーク広告、再配布設定）
  - ポリシールート（Seq、送信元、宛先、プロトコル、入力IF、出力IF、ゲートウェイ、状態）
  - DHCPサーバー（インターフェース、IP払い出し範囲、サブネットマスク、除外IP、ゲートウェイ、DNS、リース時間）

[セクション4: オブジェクト定義]
  - アドレスオブジェクト（オブジェクト名、タイプ、値、VDOM）
  - アドレスグループ（オブジェクト名、メンバー、VDOM）
  - サービスオブジェクト（オブジェクト名、プロトコル、ポート、VDOM）
  - サービスグループ（オブジェクト名、メンバー、VDOM）

[セクション5: ファイアウォールポリシー]
  - ポリシー一覧（No、ID、ポリシー名、送信元IF、宛先IF、送信元アドレス、宛先アドレス、Internet Service、サービス、アクション、NAT、セキュリティプロファイル、ログ、備考）
  - インタラクティブツールチップ（インターフェース、アドレス、サービス、セキュリティプロファイル、Internet Serviceの詳細情報を表示）
  - Local-in ポリシー（No、ID、ポリシー名、送信元IF、送信元アドレス、宛先アドレス、サービス、アクション、備考）

[セクション6: NAT設定]
  - VIP/DNAT（ルール名、外部インターフェース、外部IP、内部IP、ポートフォワーディング、VDOM、備考）
  - SNAT/IP Pool（ルール名、変換後送信元、インターフェース、VDOM、備考）

[セクション7: VPN設定]
  - IPsec-VPN Phase1（VPN名、対向機器IP、インターフェース、暗号化/認証、DHグループ、ライフタイム、事前共有鍵）
  - IPsec-VPN Phase2（VPN名、Phase1名、ローカルセグメント、リモートセグメント、暗号化/認証、PFS、ライフタイム）
  - SSL-VPN/GlobalProtect（レルム/ポータル名、リスニングポート、インターフェース、認証、払い出しIPプール、許可ユーザーグループ、モード）

[セクション8: セキュリティプロファイル（UTM）]
  - プロファイル一覧（プロファイル名、タイプ、有効/無効、VDOM、説明）
  - アンチウイルスプロファイル（プロファイル名、有効/無効、スキャンモード、スキャン対象プロトコル、アクション）
  - Webフィルタプロファイル（プロファイル名、有効/無効、適用カテゴリ、アクション）
  - アプリケーションコントロールプロファイル（プロファイル名、有効/無効、適用カテゴリ/シグネチャ、アクション）
  - IPSプロファイル（プロファイル名、有効/無効、適用シグネチャ、アクション）
  - SSLインスペクションプロファイル（プロファイル名、有効/無効、モード）

[セクション9: HA設定]
  - HA基本設定（HAモード、グループID、優先度、プリエンプト、HA管理ステータス）
  - 監視設定（監視インターフェース、HAインターフェース）

[セクション10: ログ・監視設定]
  - ログ転送（サーバーIPアドレス、ポート、転送ログ種別、状態）
  - 集中管理（FortiAnalyzer/Panoramaサーバー）
  - SNMP（SNMP有効化、バージョン、コミュニティ名/ユーザー名、ホスト、トラップ送信先IPアドレス）
[フッター]
  - ファイル名、生成情報
```

#### 3.1.2 スタイル仕様

- レスポンシブデザイン対応
- 印刷用スタイル（@media print）対応
- アクション表示の色分け
  - allow/accept: 緑色
  - deny: 赤色
  - drop: 暗赤色
- ゾーンタグの視覚的表示
  - Trust 系: 緑背景
  - Untrust 系: 赤背景
  - DMZ 系: 黄背景

#### 3.1.3 インタラクティブツールチップ機能

HTML出力では、以下のオブジェクトにマウスオーバーで詳細情報を表示するツールチップ機能を提供します：

- **インターフェース**: IPアドレス、サブネットマスク、ゾーン、役割、許可アクセス
- **アドレスオブジェクト**: タイプ、値、CIDR表記、説明
- **サービスオブジェクト**: プロトコル、ポート範囲、説明
- **セキュリティプロファイル**: タイプ、名前、設定内容
- **Internet Service**: サービス名、ISDB ID、アプリケーション名

ツールチップは表の境界を越えて表示され、画面端で自動的に位置調整されます。

#### 3.1.4 CIDR表記への自動変換

IPアドレスとサブネットマスクの組み合わせを、自動的にCIDR表記（例: `192.168.1.0/24`）に変換します。

- アドレスオブジェクト（subnetタイプ）
- インターフェースのIPアドレス
- ルーティング情報の宛先ネットワーク

### 3.2 PDF 出力

#### 3.2.1 概要

HTML出力をベースに、WeasyPrintを使用してPDF形式で出力します。
章（セクション）がページをまたがないよう、CSSの `page-break-inside: avoid` を適用しています。

#### 3.2.2 ページ設定

- 用紙サイズ: A3 横向き（landscape）
- 余白: 1.5cm（上下左右）
- ヘッダー: 「パラメータシート」
- フッター: ページ番号（「ページ X / Y」形式）
- フォントサイズ: ベース9pt、テーブル7pt（折り返しを減らすため最適化）

#### 3.2.3 ページ分割ルール

- **章（section）**: ページをまたがない（`page-break-inside: avoid`）
- **セクション見出し（h2）**: ページの先頭に配置されるよう調整
- **テーブル行（tr）**: 可能な限りページをまたがない
- **目次**: 独立したページに配置（`page-break-after: always`）
- **ヘッダー・フッター**: ページ分割時に適切に配置

#### 3.2.4 日本語フォント対応

- **フォント**: Noto Sans CJK JP（日本語フォント）を使用
- **フォントサイズ最適化**: 折り返しを減らすため、ベース9pt、テーブル7ptに調整
- **文字化け対策**: Dockerイメージに`fonts-noto-cjk`パッケージをインストール

#### 3.2.5 レイアウト

HTML出力と同じレイアウト構成を維持し、印刷に最適化されたスタイルを適用します。

### 3.3 Excel 出力

#### 3.3.1 シート構成

Excel出力は、**グローバル設定**と**VDOM/vsys単位の設定**に分けて出力されます。

##### グローバル設定シート（VDOM横断の共通設定）

| シート名         | 内容                             |
| ---------------- | -------------------------------- |
| クラスタ概要     | HAクラスタ情報（メンバー詳細: 役割、ホスト名、モデル、OSバージョン、優先度、HA管理IP、シリアル番号、設定ファイル）、設定差分 |
| 機器概要         | 機器基本情報、設定統計           |
| システム設定     | 管理設定、NTP、DNS、管理者アカウント等 |
| HA設定           | 高可用性設定                     |
| ログ・監視設定   | Syslog、SNMP、集中管理設定       |

##### VDOM/vsys単位のシート

各VDOM/vsysごとに、以下のシートが作成されます（シート名は「タイトル (VDOM名)」形式）：

| シート名         | 内容                             |
| ---------------- | -------------------------------- |
| IF (VDOM名)      | ネットワークインターフェース一覧 |
| ルート (VDOM名)  | スタティックルート、OSPF、OSPFv6、BGP、ポリシールート |
| オブジェクト (VDOM名) | アドレス/サービスオブジェクト    |
| ポリシー (VDOM名) | ファイアウォールポリシー、Local-inポリシー |
| NAT (VDOM名)     | NAT設定                          |
| VPN (VDOM名)     | IPsec VPN、SSL-VPN設定           |
| セキュリティ (VDOM名) | セキュリティプロファイル         |

**注意**: データが存在しないVDOMのシートは作成されません。

#### 3.3.2 スタイリング機能

##### モダンカラーパレット

- **VDOMごとの色テーマ**: 各VDOM/vsysに異なる色テーマを自動割り当て（最大10個のカラーパレット）
  - Indigo（デフォルト/root）、Emerald、Amber、Rose、Cyan、Violet、Orange、Teal、Pink、Sky
- **タブ色**: 各シートのタブにVDOMごとの色を設定
- **ヘッダー**: VDOMごとの色テーマに合わせたヘッダー背景色
- **ゼブラストライプ**: 交互行の背景色で視認性を向上

##### 視覚的区別

- **アクション表示**: Allow（緑）、Deny（赤）、Drop（暗赤）を色分け
- **ステータス表示**: 有効/無効を色とテキストで明確に区別
- **セクションタイトル**: 各セクションに色付きタイトル行を配置

#### 3.3.3 機能

- **自動列幅調整**: 日本語文字を考慮した列幅の自動調整
- **フリーズペイン**: ヘッダー行を固定してスクロール時も見出しを表示
- **セル結合**: セクションタイトルや見出しのセル結合による見やすいレイアウト
- **中央揃え**: 数値やステータス項目を中央揃えで表示
- **セクション指定**: 出力するセクションを指定可能（部分出力対応）
- **解析範囲シート**: 末尾に「解析範囲」シートを常に追加し、未対応の設定ブロック・未モデル化オプション・解析警告の行番号を全件記載する

### 3.4 出力項目とプリセット

出力項目のキーとプリセットは `exporters/sections.py` に定義し、Web UI と CLI で共有する（`python main.py --list-sections` で一覧表示）。

| プリセット | 含むセクション | 用途 |
| ---------- | -------------- | ---- |
| `overview` 概要 | device_info, network, policies | 初見の把握 |
| `connectivity` 接続 | device_info, topology, network, nat, vpn | 疎通・経路の確認 |
| `policy` ポリシー監査 | device_info, objects, policies, nat, security_profiles | ルール監査 |
| `ops` 運用（既定） | device_info, system_settings, network, objects, policies, nat, vpn, security_profiles, ha, logging | 引き継ぎ・障害対応 |
| `all` すべて | 全セクション | 設計書・納品 |

### 3.5 付録「解析範囲」

台帳に反映できなかった設定を、値を出さずに示す。出力項目の選択に関わらず常に出力する。

- **未対応の設定ブロック**: パーサーがモデル化していない設定パス。VDOM/vsys は範囲としてまとめる
- **未モデル化のオプション**: 台帳に載っているブロックのうち、反映していないオプション名（キー）をブロックごとに列挙する
- **解析警告**: 解釈できなかった行の件数と行番号。警告の本文は設定値を含み得るため表示しない
- HTML / PDF は各表 40 行・オプション名 8 種までを表示し、残りは件数で示す。Excel の「解析範囲」シートと JSON の `coverage` は全件を記録する
- 表題欄の件数行に「未対応・解析警告」の件数を表示し、付録へリンクする

### 3.6 JSON 出力

- `schema`: `wallscribe.parameter-sheet`、`schema_version`: `1`
- `config`: 選択したセクションに対応する統一モデルのフィールド（`device_info`、`interfaces`、`firewall_policies` など）
- `analysis`: `security_analysis` / `topology` を選択した場合の診断結果
- `coverage`: 付録「解析範囲」と同じ集約結果（全件）
- `cluster`: HA クラスタ時のメンバーと設定差分
- `redaction`: 値を伏せたフィールドの場所。`password`・`community` などの文字列は `<redacted>` に置き換える（PSK の有無などの真偽値は残す）
- HTML / PDF / Excel も同じ方針で、SNMP コミュニティの値と `ENC` で始まる文字列は出さない。パスワードや証明書の秘密鍵は「設定あり」だけを残す
- スケジュール、ユーザー / グループ、LDAP / RADIUS、SD-WAN、IPv6 ポリシー、プロキシポリシー、トラフィックシェーパー、証明書名、復号ルール、アプリケーション上書き、Panorama device-group の pre/post セキュリティルールを各セクションに含める

### 3.7 差分レポート

同一機器の変更前・変更後の 2 ファイルを、それぞれ単一機器として解析して比較する。

- **突き合わせ**: VDOM/vsys と名前（ポリシーは ID、NAT は種別と名前）。名前を変えたものは削除と追加になる
- **区分**: 追加 / 削除 / 変更（項目単位の変更前後。リストはメンバーの追加・削除を強調）/ 順序（ポリシー類で、相対順序を保った最長の並びから外れたもの）
- **対象**: 機器情報、システム設定、管理者、HA、ログ・Syslog・SNMP、インターフェース、スタティック／ポリシールート、OSPF、BGP、DHCP、仮想ルーター、アドレス・サービスとグループ、スケジュール、ローカルユーザー、ユーザーグループ、認証サーバー、SD-WAN、トラフィックシェーパー、証明書、復号ルール、アプリケーション上書き、ファイアウォール／Local-in ポリシー、NAT、IPsec Phase1/2、SSL-VPN、セキュリティプロファイル
- **秘匿値**: パスワード・SNMP コミュニティは変更の有無だけを示す
- **確認事項**: ホスト名・機種の不一致、解析警告、比較対象外の未対応ブロックの件数をレポート冒頭に示す
- **形式**: HTML / PDF（同じ HTML を組版）/ Excel（差分サマリー・変更一覧の 2 シート）/ JSON（`schema`: `wallscribe.config-diff`）

---

## 4. アーキテクチャ設計

### 4.1 モジュール構成

```text
WallScribe/
├── main.py                    # エントリーポイント（CLI）
├── app.py                     # Web アプリケーション（Flask）
├── exceptions.py              # カスタム例外クラス
├── docs/
│   ├── SPECIFICATION.md       # 本仕様書（正本）
│   └── README.md              # 本書への誘導
├── README.md                  # 使用方法（リポジトリ直下）
├── requirements.txt           # 依存パッケージ
├── docker-compose.yml         # Docker Compose
├── Dockerfile                 # Docker ビルド定義
│
├── api/                       # REST API仕様（Swagger定義・ドキュメント）
│   ├── __init__.py
│   ├── specs.py               # Swagger 2.0 定義（flasgger用）
│   └── README.md              # REST API ドキュメント
│
├── parsers/                   # パーサーモジュール
│   ├── __init__.py
│   ├── base.py                # 基底パーサー・エンコーディング・パーサー解決
│   ├── cluster.py             # HAクラスタ構成パーサー（複数ファイル統合）
│   ├── device_identification.py  # TextFSM による機種・OS の先行識別
│   ├── textfsm_utils.py       # TextFSM テンプレート実行
│   ├── textfsm_templates/     # .textfsm（identify / fortigate / paloalto 等）
│   ├── utils.py               # パーサー補助（appid 解決など）
│   ├── fortigate/             # FortiGate用パーサー（パッケージ）
│   │   ├── __init__.py
│   │   └── converters/        # 変換ロジック（device/network/objects/policies/...）
│   ├── paloalto_text.py       # Palo Alto CLI set 形式（TextFSM 取り込み）
│   └── paloalto.py            # Palo Alto用パーサー（XML ＋ set 分岐）
│
├── models/                    # データモデル
│   ├── __init__.py
│   ├── config.py              # 設定データ構造の定義
│   └── cluster.py             # HAクラスタ統合データモデル
│
├── exporters/                 # 出力モジュール
│   ├── __init__.py
│   ├── html.py                # HTML出力
│   ├── pdf.py                 # PDF出力
│   └── excel.py               # Excel出力
│
├── utils/                     # 共通ユーティリティ
│   ├── logging_config.py      # 構造化ログ（JSON）
│   ├── metrics.py             # Prometheus メトリクス
│   └── validation.py          # 入力/ファイル検証
│
├── web/                       # Web インターフェース
│   ├── __init__.py
│   └── templates/             # Web UI テンプレート
│       ├── index.html
│       ├── result.html
│       └── error.html
│
├── static/
│   ├── css/style.css
│   ├── js/main.js
│   └── data/appid.csv         # アプリケーションIDマッピング
│
└── tests/
    ├── test_parsers.py
    ├── test_device_identification.py
    └── ...
```

### 4.2 機器識別（TextFSM）

パース前に、設定テキストの**先頭最大 65536 文字**を対象に [textfsm](https://github.com/google/textfsm) テンプレートで機種を推定する（`parsers/device_identification.py` の `identify_device`）。

| フィールド | 説明 |
| --- | --- |
| `IdentifiedDevice.device_type` | `DeviceType`（FortiGate / Palo Alto / UNKNOWN） |
| `IdentifiedDevice.model` | 推定モデル名（FortiGate は `format_fortigate_display_model` で整形、Palo Alto 識別時は `PA Series`） |
| `IdentifiedDevice.os_version` | 推定 OS バージョン（テンプレートが抽出した場合） |

**識別用テンプレート（`parsers/textfsm_templates/identify/`）**

| テンプレート | 用途 |
| --- | --- |
| `fortigate_header.textfsm` | `#config-version=` 行から機種コード・FortiOS 版を抽出 |
| `paloalto_config_root.textfsm` | `<config version="...">` から `version` を抽出 |
| `paloalto_set_cli.textfsm` | 先頭付近の `set deviceconfig|vsys|shared|…` 行で Palo Alto（set 形式）を識別 |

**パース時の TextFSM（例）**

| テンプレート | 用途 |
| --- | --- |
| `fortigate/header_config_version.textfsm` | FortiGate ヘッダ 1 行からモデル・ビルド・opmode・vdom 等を抽出 |
| `paloalto/set_cli.textfsm` | PAN-OS `set` 形式から hostname・address（ip-netmask / fqdn、vsys / shared）等を行単位で抽出 |

**TextFSM と階層パーサーの分界**

- **FortiGate**: `config` / `edit` / `set` / `next` / `end` の入れ子は [TextFSM](https://github.com/google/textfsm) の行レコードモデルに適さないため、`parsers/fortigate/__init__.py` の `_parse_config_tree` による**文法パーサー**を維持する。ヘッダ等、単一行で完結するパターンのみ TextFSM に寄せる（現状 `header_config_version`、機器識別）。`_separate_config` は単純な部分文字列・トークン分割であり、追加の TextFSM 化の優先度は低い。
- **Palo Alto**: **XML エクスポート**は `ElementTree` による階層走査のまま。**CLI `set` 形式**は `parsers/paloalto_text.py` で TextFSM に渡し、宣言的にフィールド抽出する（第 1 弾はアドレス・ホスト名等。ポリシー等はテンプレート拡張で対応可能）。

`get_parser_for_content` は、識別結果が FortiGate / Palo Alto のいずれかに一致すれば対応パーサーを選択する。一致しない場合は各パーサーの `detect_content_type` にフォールバックする。`get_parser_for_file` は拡張子で候補が決まったあと、読み込んだ内容に対して `identify_device` を実行し、パーサーインスタンスの `identification` に格納する。**拡張子が `.conf` / `.xml` / `.set` に該当しない場合**は、ファイル全文を読み `get_parser_for_content` で内容ベースの選択にフォールバックする。

### 4.3 クラス設計

#### 4.3.1 基底パーサークラス

実装の要点（詳細は `parsers/base.py`）。

```python
class BaseConfigParser(ABC):
    """設定パーサーの基底クラス"""

    def __init__(self):
        self.config_model = ConfigModel()
        self.errors: list = []
        self.identification = None  # Optional[IdentifiedDevice]

    @abstractmethod
    def parse(self, file_path: str) -> ConfigModel: ...

    @abstractmethod
    def parse_content(self, content: str, filename: str = "") -> ConfigModel: ...

    @staticmethod
    @abstractmethod
    def detect_file_type(file_path: str) -> bool: ...

    @staticmethod
    @abstractmethod
    def detect_content_type(content: str) -> bool: ...
```

#### 4.3.2 データモデル

```python
@dataclass
class ConfigModel:
    """統一設定データモデル"""
    device_info: DeviceInfo           # 機器概要
    system_settings: SystemSettings   # システム設定
    interfaces: List[Interface]       # インターフェース
    routes: List[Route]               # ルーティング
    objects: Objects                  # オブジェクト定義
    policies: List[Policy]            # ポリシー
    vpn: VPNSettings                  # VPN設定
    ha: HASettings                    # HA設定
    logging: LoggingSettings          # ログ設定
```

### 4.4 処理フロー

```text
1. 入力の取得（ファイルパスまたはアップロード）
   ↓
2. 文字エンコーディングの決定とデコード
   （detect_encoding または read_file: utf-8-sig → utf-8 → cp932 → latin-1）
   ↓
3. パーサー候補の決定
   （パス: detect_file_type → identify_device；拡張子不一致時は読込後 `get_parser_for_content`／内容のみ: identify_device → フォールバック detect_content_type）
   ↓
4. 機器識別（TextFSM、先頭最大 65536 文字）→ parser.identification
   ↓
5. 対応パーサーで parse / parse_content
   ↓
6. 統一データモデル（ConfigModel）で返却
   ↓
7. 指定形式で出力（HTML / PDF / Excel）
```

---

## 5. CLI 仕様

### 5.1 コマンド構文

```bash
python main.py <input_file> [options]
```

### 5.2 オプション

| オプション  | 短縮形 | 説明                  | デフォルト      |
| ----------- | ------ | --------------------- | --------------- |
| `--output`  | `-o`   | 出力ファイルパス。HTML 以外は拡張子を形式に合わせる | `output.<拡張子>`（`--compare` 時は `diff.<拡張子>`） |
| `--format`  | `-f`   | 出力形式 (html/excel/pdf/json)。pdf は WeasyPrint が必要 | `html` |
| `--preset`  |        | 出力項目のプリセット（3.4 参照） | 全セクション |
| `--sections`|        | 出力するセクションをカンマ区切りで指定（`--preset` より優先） | - |
| `--list-sections` |  | セクションとプリセットの一覧を表示して終了 | - |
| `--compare` |        | 2 ファイルを変更前・変更後として比較し、差分レポートを出力 | - |
| `--ha-mode` |        | HAモード (auto/single/cluster) | `auto` |
| `--verbose` | `-v`   | 詳細ログ出力（解析警告の本文も表示） | `false` |
| `--help`    | `-h`   | ヘルプ表示            | -               |
| `--version` |        | バージョン表示        | -               |

### 5.3 使用例

```bash
# FortiGate設定からHTML生成
python main.py fortigate.conf -o fortigate_param.html

# Palo Alto設定からExcel生成
python main.py pa440.xml -f excel -o pa440_param.xlsx

# HA構成（複数ファイル）をクラスタとして処理（自動判定）
python main.py primary.conf secondary.conf --ha-mode auto -o ha_cluster.html

# 詳細ログ付きで実行
python main.py config.conf -v

# ポリシー監査向けの JSON
python main.py fortigate.conf -f json --preset policy -o policy.json

# 変更前・変更後の差分を Excel で出力
python main.py before.conf after.conf --compare -f excel -o changes.xlsx
```

---

## 6. Web インターフェース仕様

### 6.1 概要

ブラウザから設定ファイルをアップロードし、パラメータシートを生成・ダウンロードできる Web アプリケーション。

### 6.2 画面構成

#### 6.2.1 メインページ（アップロード画面）

```text
┌─────────────────────────────────────────────────────────────┐
│  [Logo] WallScribe                                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                                                       │  │
│  │     設定ファイルをドラッグ＆ドロップ                  │  │
│  │            または                                     │  │
│  │        [ファイルを選択（複数可）]                     │  │
│  │                                                       │  │
│  │     対応形式: .conf (FortiGate), .xml (Palo Alto)     │  │
│  │                                                       │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                             │
│  選択済みファイル: primary.conf / secondary.conf ...         │
│  HA構成オプション: ○ auto (推奨) ○ cluster ○ single          │
│                                                             │
│  出力形式:  ○ HTML (推奨)   ○ PDF   ○ Excel               │
│                                                             │
│              [パラメータシートを生成]                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

#### 6.2.2 結果表示

生成物は応答本文としてブラウザに届き、ページのメモリ上（Blob）にだけ置く。結果はアップロード画面と同じページに表示する。

- 表題欄: ホスト名・機器・OS バージョン・（HA 時）モードと台数・出力ファイル名
- 件数: インターフェース・ルート・ポリシー・NAT・オブジェクト（差分は追加・削除・変更・順序変更）
- 未対応・解析警告がある場合は付録「解析範囲」への注意を表示
- 操作: ダウンロード / 新しいタブで開く（Excel 以外）/ 別のファイルを処理
- ページを閉じると生成物は消える（サーバーにも残らない）

### 6.3 API エンドポイント

設定ファイルと生成物はサーバーに保存しない。変換・比較は 1 リクエストで完結し、応答本文が生成物になる。

| メソッド | パス                   | 説明                             |
| -------- | ---------------------- | -------------------------------- |
| GET      | `/`                    | メインページ表示                 |
| POST     | `/api/v1/convert`      | パラメータシートを生成して直接返す（複数ファイル/HA対応） |
| POST     | `/api/v1/compare`      | 2 世代の差分レポートを生成して直接返す |
| GET      | `/api/v1/spec`         | Swagger仕様（JSON）              |
| GET      | `/swagger.json`        | Swagger仕様（JSON エイリアス）   |
| GET      | `/metrics`             | Prometheusメトリクス             |
| GET      | `/health`              | ヘルスチェック（基本）           |
| GET      | `/health/ready`        | レディネスチェック（`storage: none` と出力形式の可用性） |
| GET      | `/health/live`         | ライブネスチェック               |

サーバー側で生成物を保持していた旧エンドポイント（`/upload`、`/upload_async`、`/api/v1/upload`、`/api/(v1/)status|progress/<id>`、`/api/v1/jobs`、`/(api/v1/)download|preview/<id>`）は `410 Gone`（`code: STATELESS_MODE`）で移行先を返す。`/result/<id>` も 410 のページを返す。

### 6.4 変換・比較 API 仕様

#### 6.4.1 リクエスト

```http
POST /api/v1/convert HTTP/1.1
Content-Type: multipart/form-data

config_files[]  設定ファイル（1 つ以上。2 つ以上で HA 構成として処理）
output_format   html | pdf | excel | json（既定 html）
ha_mode         auto | single | cluster（既定 auto）
sections        出力するセクションの JSON 配列（空・不正な JSON は全セクション）
```

```http
POST /api/v1/compare HTTP/1.1
Content-Type: multipart/form-data

before          変更前の設定ファイル
after           変更後の設定ファイル
output_format   html | pdf | excel | json（既定 html）
```

#### 6.4.2 レスポンス（成功時）

- 本文: 生成物（`Content-Disposition: attachment`、日本語ファイル名は `filename*` で返す）
- `Cache-Control: no-store`
- `X-WallScribe-Summary`: 要約の JSON（ASCII）。変換は `kind`・`filename`・`device_type`・`hostname`・`version`・件数・`coverage`、比較は `before`・`after`・`same_device`・`added`・`removed`・`modified`・`moved`

#### 6.4.3 レスポンス（エラー時）

```json
{
  "success": false,
  "error": {
    "code": "UNSUPPORTED_FORMAT",
    "message": "サポートされていないファイル形式です",
    "details": "対応形式: .conf (FortiGate), .xml (Palo Alto)"
  }
}
```

### 6.5 セキュリティ要件

#### 6.5.1 ファイルアップロード制限

| 項目               | 制限値                    |
| ------------------ | ------------------------- |
| 最大ファイルサイズ | 50MB（`MAX_CONTENT_LENGTH`、リクエスト全体） |
| 許可拡張子         | `.conf`, `.xml`           |
| 同時アップロード数 | 複数ファイル（HA構成を想定、合計サイズは最大ファイルサイズに従う） |
| サーバーでの保存   | しない（メモリ上で処理し、応答で返す） |

#### 6.5.2 セキュリティ対策（現状実装）

- 設定ファイル・生成物・要約をサーバーのディスクに書かない（アップロードはメモリで受け、openpyxl も一時ファイルを使わずに保存する）
- ログに設定由来の値（ホスト名・ファイル名・パース警告の本文・例外メッセージ）を出さない。例外は種類と発生箇所のみ
- Docker はボリュームを持たず、読み取り専用のルート FS・全 capability 削除・no-new-privileges で起動する
- ファイル内容の検証（拡張子に加えて内容を簡易検証）
- ダウンロード名から制御文字・パス区切りを除去
- レート制限（変換・比較は 10 リクエスト/分）
- セキュリティヘッダー（CSP / HSTS(HTTPS時) / Referrer-Policy / Permissions-Policy）

### 6.6 UI/UX 要件

#### 6.6.1 ドラッグ＆ドロップ

- ファイルをドロップエリアにドラッグすると視覚的フィードバック
- ドロップ後にファイル選択として反映（その後「生成」ボタンで送信）
- 進捗表示（プログレスバー）

#### 6.6.2 レスポンシブデザイン

- デスクトップ / タブレット / モバイル対応
- 最小幅: 320px

#### 6.6.3 処理中の表示

```text
┌───────────────────────────────────────────────────┐
│                                                   │
│          ⏳ パラメータシートを生成中...           │
│                                                   │
│          [=========>          ] 45%               │
│                                                   │
│          設定ファイルを解析しています             │
│                                                   │
└───────────────────────────────────────────────────┘
```

### 6.7 エラーハンドリング

| エラーコード        | 説明                         | ユーザーへの表示                           |
| ------------------- | ---------------------------- | ------------------------------------------ |
| `FILE_TOO_LARGE`    | ファイルサイズ超過           | ファイルサイズが 50MB を超えています       |
| `UNSUPPORTED_FORMAT`| 非対応ファイル形式           | 対応形式: .conf, .xml                      |
| `PARSE_ERROR`       | パースエラー                 | 設定ファイルの解析に失敗しました           |
| `INTERNAL_ERROR`    | サーバー内部エラー           | 予期しないエラーが発生しました             |

### 6.8 Web サーバー起動

```bash
# 開発モード
python app.py

# 本番モード（Gunicorn使用）
gunicorn -w 4 -b 0.0.0.0:8080 app:app
```

### 6.9 環境変数

| 変数名              | 説明                  | デフォルト値      |
| ------------------- | --------------------- | ----------------- |
| `FLASK_ENV`         | 実行環境              | `production`      |
| `FLASK_PORT`        | リッスンポート        | `8080`            |
| `MAX_CONTENT_LENGTH`| 最大アップロードサイズ| `52428800` (50MB) |
| `SECRET_KEY`        | セッション暗号化キー  | ランダム生成      |
| `TZ`                | タイムゾーン          | `Asia/Tokyo`      |
| `LOG_FORMAT`        | ログ形式（text/json） | `text`            |
| `LOG_LEVEL`         | ログレベル            | `INFO`            |
| `RATE_LIMIT_DEFAULT`| デフォルトレート制限  | `100 per hour`    |
| `RATE_LIMIT_STORAGE`| レート制限ストレージ  | `memory://`       |

`UPLOAD_FOLDER`・`CLEANUP_INTERVAL`・`WALLSCRIBE_DISABLE_CLEANUP_THREAD` は廃止した（設定しても使わない）。

---

## 7. 技術仕様

### 7.1 動作環境

- Python 3.12 以上
- OS: Windows / Linux / macOS

### 7.2 依存パッケージ

直接の範囲は `requirements.in`、インストールする版は `pip-compile` でピン留めした `requirements.txt` です。テスト用パッケージは含めず、`pip install -e ".[dev]"`（`pyproject.toml` の `[project.optional-dependencies]`、`pytest>=8.3.0` 等）で入れます。主な依存の分類は次のとおりです。

```text
flask>=2.0.0              # Web フレームワーク
werkzeug>=2.0.0           # WSGI ユーティリティ
jinja2>=3.0.0             # HTMLテンプレート
gunicorn>=20.0.0          # 本番用WSGIサーバー
openpyxl>=3.0.0           # Excel出力
weasyprint>=61.2          # PDF出力
defusedxml>=0.7.1         # XXE対策（Palo Alto XML）
textfsm>=1.1.3            # 機器識別・TextFSM テンプレート
flask-limiter>=3.5.0      # レート制限
prometheus-client>=0.19.0 # メトリクス
flasgger>=0.9.7.1         # Swagger UI（Swagger 2.0）
```

### 7.3 文字エンコーディング

- 入力: UTF-8（BOM 付き）/ UTF-8 / CP932(Shift-JIS) / Latin-1 を順に試行（`parsers/base.detect_encoding` および `read_file`）
- 出力: UTF-8

---

## 8. エラーハンドリング

### 8.1 想定エラー

| エラー種別         | 原因                 | 対処                           |
| ------------------ | -------------------- | ------------------------------ |
| FileNotFoundError  | ファイルが存在しない | エラーメッセージを表示して終了 |
| ParseError         | 設定ファイル形式不正 | 該当箇所をスキップして継続     |
| EncodingError      | 文字コード不正       | UTF-8 以外を試行               |
| UnsupportedVersion | 未対応バージョン     | 警告を出力して継続             |

### 8.2 ログ出力

CLI（利用者の端末で実行）:

```text
[INFO] ファイル読み込み: fortigate.conf
[INFO] 機器タイプ: FortiGate
[INFO] パース完了: 120ポリシー、45オブジェクト
[WARNING] 解析警告: 2 件
[INFO] HTML出力: output.html
```

Web アプリ（サーバー）は設定由来の値をログに出さない:

```text
[WARNING] [services.conversion] 設定の解析で警告が 2 件ありました
[INFO] [exporters.pdf] PDF生成完了 (html=0.04s, render=0.95s, total=0.99s)
[ERROR] [routes.convert] 変換エラー: ValueError (vdom.py:412)
```

---

## 9. 制限事項

### 9.1 取得できない情報

- シリアル番号（設定ファイルに含まれない）
- ライセンス情報（設定ファイルに含まれない）
- パスワード（マスク化されている）
- 実行コンフィグと起動コンフィグの差分

### 9.2 対応バージョン

- FortiOS: 6.0 以上を推奨（6.x/7.x/8.0 で動作確認）
- PAN-OS: 10.0 以上を推奨（10.x/11.x で動作確認）

### 9.3 パフォーマンス

- 大規模設定（10,000 ポリシー以上）では処理時間が増加する可能性あり

---

## 10. 今後の拡張予定

### 10.1 Phase 1（初期リリース）

- [x] FortiGate HTML パーサー（既存実装を統合）
- [x] Palo Alto HTML パーサー（既存実装を統合）
- [x] 統合 CLI ツール
- [x] 基本 HTML テンプレート
- [x] Web インターフェース（ファイルアップロード・ダウンロード）
- [x] 機器概要セクション（シリアル番号、ライセンス情報含む）
- [x] システム設定セクション（管理アクセス詳細含む）
- [x] ネットワーク設定セクション（インターフェース役割、DHCP除外IP含む）
- [x] ファイアウォールポリシーセクション（NAT、セキュリティプロファイル、Local-inポリシー含む）
- [x] NAT設定セクション（VIP/DNATポートフォワーディング詳細含む）
- [x] VPN設定セクション（IPsec Phase1/Phase2詳細、SSL-VPN認証方法含む）
- [x] セキュリティプロファイルセクション（UTM詳細情報含む）
- [x] ログ・監視設定セクション（ログ転送種別、SNMPトラップ送信先含む）

### 10.2 Phase 2

- [x] Excel 出力対応
- [x] Excel 出力: VDOM/vsys単位出力対応、モダンスタイリング（VDOMごとの色テーマ、タブ色、視覚的区別）
- [x] PDF 出力対応
- [x] Internet Service対応
- [x] CIDR表記への自動変換
- [x] インタラクティブツールチップ機能
- [x] 複数ファイルのHAクラスタ処理（Web/CLI）
- [ ] 差分比較機能
- [ ] バッチ処理（複数ファイル同時変換）

### 10.3 Phase 3

- [ ] 他ベンダー対応（Cisco ASA、Juniper SRX 等）
- [x] Docker コンテナ化
  - [x] マルチステージビルド対応
  - [x] 日本語フォント対応
  - [x] タイムゾーン設定（環境変数で指定可能、デフォルト: Asia/Tokyo）
  - [x] ファイルシステムベースのメタデータ管理（複数ワーカー対応）
- [x] REST API 化（Swagger UI / ジョブ管理API を含む）

---

## 11. 付録

### 11.1 参照ドキュメント

- FortiOS CLI Reference
- PAN-OS CLI Reference Guide
- Flask Documentation
- 既存実装: `fg-parameter-sheet/`, `pa-parameter-sheet/`

### 11.2 更新履歴

| バージョン | 日付       | 変更内容                           |
| ---------- | ---------- | ---------------------------------- |
| 1.4        | 2026-04-01 | ドキュメント統合                   |
|            |            | - `docs/` 内の章別 Markdown を本ファイル（SPECIFICATION.md）に統合し正本を一本化 |
|            |            | - アーキテクチャに機器識別（TextFSM）・`textfsm` 依存・モジュールツリーを反映 |
| 1.3        | 2026-04-03 | （履歴）章別ファイル分割（本版で統合済み） |
|            |            | - 機器識別（TextFSM）・実装に合わせた追記 |
| 1.2        | 2026-01-25 | API拡充・ドキュメント更新          |
|            |            | - Swagger UIの定義をSwagger 2.0に統一（`/apidocs`） |
|            |            | - ジョブ管理API追加（`/api/v1/jobs` など） |
|            |            | - 仕様JSON提供（`/api/v1/spec`, `/swagger.json`） |
|            |            | - ヘルスチェック追加（`/health`, `/health/ready`, `/health/live`） |
|            |            | - 入力検証強化（サイズ/内容/パラメータ） |
| 1.1        | 2026-01-23 | 現状版（v1.1）                     |
|            |            | - Excel出力: クラスタメンバーシートに「モデル」「OSバージョン」「HA管理IP」列を追加 |
|            |            | - Excel出力: ルーティングシートにOSPF、OSPFv6、BGP、ポリシールート対応を追加 |
|            |            | - Excel出力: HA設定に「HA管理ステータス」を追加 |
|            |            | - HTML出力: CSSスタイリングを改善（ユニバーサルデザイン、視認性向上） |
| 1.0        | 2026-01-20 | 現状版（v1.0）                     |
|            |            | - HTML / PDF / Excel 出力 |
|            |            | - Web: 非同期生成＋進捗表示（`/upload_async` + `/api/progress/<file_id>`） |
|            |            | - Web/CLI: 複数ファイル（HAクラスタ）対応（`ha_mode=auto/cluster/single`） |
|            |            | - FortiGate: IPv6（interface/route/address）対応 |
|            |            | - ルーティング: blackhole/discard対応 |
|            |            | - Excel出力機能実装（openpyxl使用） |
|            |            | - Excel出力: VDOM/vsys単位出力対応、モダンスタイリング（VDOMごとの色テーマ、タブ色、視覚的区別） |
|            |            | - Internet Service対応（FortiGateのinternet-service-name） |
|            |            | - CIDR表記への自動変換（アドレス、インターフェース、ルート） |
|            |            | - インタラクティブツールチップ機能（インターフェース、アドレス、サービス、セキュリティプロファイル、Internet Service） |
|            |            | - テストスイートの充実（pytest使用） |
|            |            | - FortiGate / Palo Alto設定ファイルパーサー実装 |
|            |            | - HTML / PDF出力機能実装 |
|            |            | - Web インターフェース実装（ファイルアップロード・ダウンロード） |
|            |            | - 機器概要セクション（シリアル番号、ライセンス情報含む） |
|            |            | - システム設定セクション（管理アクセス詳細含む） |
|            |            | - ネットワーク設定セクション（インターフェース役割、DHCP除外IP含む） |
|            |            | - ファイアウォールポリシーセクション（NAT、セキュリティプロファイル、Local-inポリシー含む） |
|            |            | - NAT設定セクション（VIP/DNATポートフォワーディング詳細含む） |
|            |            | - VPN設定セクション（IPsec Phase1/Phase2詳細、SSL-VPN認証方法含む） |
|            |            | - セキュリティプロファイルセクション（UTM詳細情報含む） |
|            |            | - ログ・監視設定セクション（ログ転送種別、SNMPトラップ送信先含む） |
|            |            | - PDF出力: A3横向き対応、フォントサイズ最適化（9pt/7pt）、日本語フォント対応 |
|            |            | - Docker: マルチステージビルド、日本語フォント対応、タイムゾーン設定、複数ワーカー対応 |
|            |            | - アプリケーションID解決機能（static/data/appid.csv によるマッピング） |
|            |            | - Internet Service 対応、CIDR表記への自動変換、インタラクティブツールチップ |
|            |            | - FortiGate: IPv6 対応、ルーティング: blackhole/discard 対応 |
|            |            | - pytest によるテスト、Docker 対応 |
