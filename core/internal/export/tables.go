package export

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/negitanu/etherscribe/internal/model"
)

// Table shape follows the model, not the fields present in an individual JSON record.
// Values always come from RenderJSON so the existing secret scrubber is shared.
type reportRecord struct {
	value                 map[string]any
	parent, vdom, context string
}
type reportField struct {
	key string
	typ reflect.Type
}

var reportOrder = []string{"deviceInfo", "virtualContexts", "systemSettings", "interfaces", "vlans", "lags", "stp", "routes", "routing", "dhcpServers", "objects", "firewallPolicies", "natPolicies", "vpn", "securityProfiles", "securityProfilesDetail", "ha", "logging", "coverage"}
var reportSections = map[string]string{"deviceInfo": "device_info", "virtualContexts": "device_info", "systemSettings": "system_settings", "interfaces": "network", "routes": "network", "routing": "network", "dhcpServers": "network", "vlans": "vlans", "lags": "lag", "stp": "stp", "objects": "objects", "firewallPolicies": "policies", "natPolicies": "nat", "vpn": "vpn", "securityProfiles": "security_profiles", "securityProfilesDetail": "security_profiles", "ha": "ha", "logging": "logging"}

func reportTables(cfg *model.Config, sections []string) ([]sheetData, error) {
	data, err := RenderJSON(cfg, sections)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err = json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	types := map[string]reflect.Type{}
	typ := reflect.TypeOf(model.Config{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		types[jsonKey(f)] = f.Type
	}
	var tables []sheetData
	for _, key := range reportOrder {
		if key != "coverage" && !SectionSelected(sections, reportSections[key]) {
			continue
		}
		typ := types[key]
		single := typ.Kind() != reflect.Slice
		tables = append(tables, buildReportTables(key, reportLabel(key), typ, toReportRecords(doc[key]), single)...)
	}
	if len(cfg.ParseErrors) > 0 {
		tables = append(tables, sheetData{Name: "解析警告", Rows: [][]string{{"項目", "件数"}, {"解析警告（設定値を含む可能性があるため本文は非表示）", fmt.Sprint(len(cfg.ParseErrors))}}})
	}
	return tables, nil
}
func jsonKey(f reflect.StructField) string { return strings.Split(f.Tag.Get("json"), ",")[0] }
func recordType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}
	return typ
}
func toReportRecords(value any) []reportRecord {
	var records []reportRecord
	if items, ok := value.([]any); ok {
		for _, item := range items {
			if row, ok := item.(map[string]any); ok {
				records = append(records, reportRecord{value: row})
			}
		}
	} else if row, ok := value.(map[string]any); ok {
		records = append(records, reportRecord{value: row})
	}
	return records
}
func reportFields(typ reflect.Type) (scalar, children []reportField) {
	typ = recordType(typ)
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		field := reportField{key: jsonKey(f), typ: f.Type}
		if recordType(f.Type).Kind() == reflect.Struct {
			children = append(children, field)
		} else {
			scalar = append(scalar, field)
		}
	}
	return
}
func orderedReportFields(path string, fields []reportField) []reportField {
	priorities := map[string][]string{
		"routing.bgp.neighbors": {"ip", "remoteAs", "localAs", "group", "description", "updateSource", "importPolicies", "exportPolicies"},
		"routes":                {"name", "destination", "gateway", "interface", "distance", "metric", "priority", "action", "description", "routeType", "enabled", "context", "vdom"},
		"interfaces":            {"name", "interfaceType", "mode", "ipAddress", "addresses", "addressMode", "mtu", "routingContext", "vlanId", "context", "vdom", "zone", "lagParent", "description"},
		"vlans":                 {"name", "vlanId", "members", "l3Interface", "context", "description"},
		"firewallPolicies":      {"policyId", "name", "sourceInterface", "destinationInterface", "sourceAddress", "destinationAddress", "applications", "service", "action", "enabled", "natEnabled", "logEnabled", "logStart", "securityProfiles", "sourceUsers", "urlCategories", "tags", "schedule", "description", "vdom"},
		"objects.addresses":     {"name", "type", "value", "description", "tags", "vdom"},
		"objects.addressGroups": {"name", "members", "description", "dynamicFilter", "tags", "vdom"},
		"objects.services":      {"name", "protocol", "port", "sourcePort", "tcpPort", "udpPort", "sctpPort", "description", "tags", "vdom"},
		"objects.serviceGroups": {"name", "members", "description", "tags", "dynamicFilter", "vdom"},
	}
	order, ok := priorities[path]
	if !ok {
		return fields
	}
	rank := map[string]int{}
	for i, k := range order {
		rank[k] = i
	}
	sort.SliceStable(fields, func(i, j int) bool {
		a, ok := rank[fields[i].key]
		if !ok {
			a = len(order)
		}
		b, ok := rank[fields[j].key]
		if !ok {
			b = len(order)
		}
		return a < b
	})
	return fields
}
func buildReportTables(path, title string, typ reflect.Type, records []reportRecord, single bool) []sheetData {
	if strings.HasPrefix(path, "objects.") {
		sort.SliceStable(records, func(i, j int) bool {
			a, b := records[i].value, records[j].value
			return fmt.Sprint(a["vdom"], "\x00", a["name"]) < fmt.Sprint(b["vdom"], "\x00", b["name"])
		})
	}
	scalar, children := reportFields(typ)
	scalar = orderedReportFields(path, scalar)
	var tables []sheetData
	if len(scalar) > 0 {
		table := sheetData{Name: title}
		if single {
			table.Rows = [][]string{{"項目", "値"}}
			for _, f := range scalar {
				var value any
				if len(records) > 0 {
					value = reportFieldValue(records[0].value, f)
				}
				table.Rows = append(table.Rows, []string{columnLabel(path, f.key), reportValue(f.key, value)})
			}
		} else {
			parentColumns := false
			for _, r := range records {
				parentColumns = parentColumns || r.parent != ""
			}
			header := []string{}
			if parentColumns {
				header = append(header, "親設定", "親 VDOM / vsys", "親コンテキスト")
			}
			for _, f := range scalar {
				header = append(header, columnLabel(path, f.key))
			}
			table.Rows = [][]string{header}
			for _, r := range records {
				row := []string{}
				if parentColumns {
					row = append(row, r.parent, r.vdom, r.context)
				}
				for _, f := range scalar {
					row = append(row, reportValue(f.key, reportFieldValue(r.value, f)))
				}
				table.Rows = append(table.Rows, row)
			}
			if len(records) == 0 {
				row := make([]string, len(header))
				row[0] = "該当する設定はありません"
				table.Rows = append(table.Rows, row)
			}
		}
		tables = append(tables, table)
	}
	for _, child := range children {
		var nested []reportRecord
		for _, r := range records {
			childRecords := toReportRecords(r.value[child.key])
			for _, cr := range childRecords {
				cr.parent = r.parent
				// Only actual parent records contribute identity; container structs add no artificial row/index.
				if !single {
					id := reportIdentity(r.value)
					if id != "" {
						if cr.parent != "" {
							cr.parent += " / "
						}
						cr.parent += title + ": " + id
					}
				}
				cr.vdom = firstReportString(r.value["vdom"], r.vdom)
				cr.context = firstReportString(r.value["context"], r.context)
				nested = append(nested, cr)
			}
		}
		childPath := path + "." + child.key
		// Object inventories always retain their four stable tables, including empty states.
		if len(nested) == 0 && path != "objects" {
			continue
		}
		childTitle := title + "・" + reportLabel(child.key)
		if path == "objects" || path == "vpn" || path == "routing" || path == "securityProfilesDetail" {
			childTitle = reportLabel(child.key)
		}
		tables = append(tables, buildReportTables(childPath, childTitle, child.typ, nested, false)...)
	}
	if len(tables) == 0 {
		tables = append(tables, sheetData{Name: title, Rows: [][]string{{"項目", "値"}, {"該当する設定はありません", ""}}})
	}
	return tables
}
func reportFieldValue(values map[string]any, field reportField) any {
	value, exists := values[field.key]
	if !exists && field.typ.Kind() == reflect.Bool {
		return false
	}
	return value
}
func reportIdentity(value map[string]any) string {
	var parts []string
	for _, k := range []string{"name", "policyId", "interface", "asNumber", "routerId", "areaId", "username", "id"} {
		if s, ok := value[k].(string); ok && s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " / ")
}
func firstReportString(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}
func columnLabel(path, key string) string {
	if path == "routes" && key == "distance" {
		return "管理距離 / Preference"
	}
	if path == "routing.bgp.neighbors" && key == "group" {
		return "BGP グループ"
	}
	if path == "interfaces" && key == "addresses" {
		return "IP アドレス一覧"
	}
	if key == "name" {
		switch path {
		case "objects.addresses", "objects.services":
			return "オブジェクト名"
		case "objects.addressGroups", "objects.serviceGroups":
			return "グループ名"
		}
	}
	return reportLabel(key)
}
func reportValue(key string, value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case bool:
		if strings.HasSuffix(key, "Configured") {
			if v {
				return "設定あり"
			}
			return "なし"
		}
		if v {
			return "有効"
		}
		return "無効"
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, reportValue(key, item))
		}
		return strings.Join(parts, ", ")
	case string:
		if v == "<redacted>" {
			return "秘匿済み"
		}
		if key == "status" || key == "fortianalyzerStatus" {
			if v == "enable" {
				return "有効"
			}
			if v == "disable" {
				return "無効"
			}
		}
		return v
	default:
		return fmt.Sprint(v)
	}
}

var reportLabels = map[string]string{
	"ipv6Enabled": "IPv6 有効", "destinationObject": "宛先オブジェクト", "group": "グループ", "mtu": "MTU", "routingContext": "ルーティング所属", "localAs": "ローカル AS", "importPolicies": "受信ポリシー", "exportPolicies": "送信ポリシー",
	"deviceInfo": "機器概要", "virtualContexts": "仮想コンテキスト", "systemSettings": "システム設定", "interfaces": "インターフェース", "vlans": "VLAN", "lags": "LAG", "stp": "STP", "routes": "ルート", "routing": "動的ルーティング", "dhcpServers": "DHCP", "objects": "オブジェクト", "firewallPolicies": "ポリシー", "natPolicies": "NAT", "vpn": "VPN", "securityProfiles": "セキュリティプロファイル", "securityProfilesDetail": "セキュリティ詳細", "ha": "HA", "logging": "ログ・監視", "coverage": "解析範囲",
	"hostname": "ホスト名", "model": "機種", "osVersion": "OS バージョン", "name": "名前", "deviceType": "機器種別", "role": "役割", "interfaceType": "種別", "ipAddress": "IP アドレス", "context": "コンテキスト", "vdom": "VDOM / vsys", "status": "状態", "description": "説明", "destination": "宛先", "gateway": "ゲートウェイ", "interface": "インターフェース", "switchportMode": "ポートモード", "accessVlan": "アクセス VLAN", "trunkVlans": "トランク VLAN", "nativeVlan": "ネイティブ VLAN", "lagParent": "所属 LAG", "vlanId": "VLAN ID", "members": "メンバー", "lacpEnabled": "LACP 有効", "policyId": "ID", "sourceAddress": "送信元アドレス", "destinationAddress": "宛先アドレス", "sourceInterface": "送信元 IF", "destinationInterface": "宛先 IF", "service": "サービス", "action": "動作", "path": "パス", "reason": "理由", "note": "備考",
	"type": "タイプ", "value": "値", "tags": "タグ", "dynamicFilter": "動的フィルター", "protocol": "プロトコル", "port": "ポート", "sourcePort": "送信元ポート", "tcpPort": "TCP ポート", "udpPort": "UDP ポート", "sctpPort": "SCTP ポート",
	"addresses": "アドレスオブジェクト", "addressGroups": "アドレスグループ", "services": "サービスオブジェクト", "serviceGroups": "サービスグループ", "adminUsers": "管理者アカウント", "ipsecPhase1": "IPsec Phase1", "ipsecPhase2": "IPsec Phase2", "sslVpn": "SSL-VPN", "ospf": "OSPF", "ospf6": "OSPFv6", "bgp": "BGP", "policyRoutes": "ポリシールート", "neighbors": "ネイバー", "networks": "広告ネットワーク", "areas": "エリア", "redistributes": "再配送", "applications": "アプリケーション", "antivirus": "アンチウイルス", "webfilter": "Web フィルター", "appControl": "アプリケーション制御", "ips": "IPS", "sslInspection": "SSL インスペクション", "syslogServers": "Syslog サーバー", "snmp": "SNMP", "heartbeatInterfacesDetail": "ハートビート詳細", "haMgmtInterfaces": "HA 管理インターフェース",
	"vendorFamily": "ベンダー系列", "operationMode": "動作モード", "vdomEnabled": "VDOM / vsys 有効", "vdomList": "VDOM / vsys 一覧", "osVersionSource": "OS バージョンの取得元", "id": "ID", "l3Interface": "L3 インターフェース", "bridges": "ブリッジ", "enabled": "状態", "dnsPrimary": "プライマリ DNS", "dnsSecondary": "セカンダリ DNS", "timezone": "タイムゾーン", "httpsPort": "HTTPS ポート", "sshPort": "SSH ポート", "ntpServers": "NTP サーバー", "username": "ユーザー名", "profile": "管理者プロファイル", "trustHosts": "接続許可ホスト", "password": "パスワード", "alias": "エイリアス", "zone": "ゾーン", "allowedAccess": "管理アクセス", "parentInterface": "親インターフェース", "distance": "管理距離", "routeType": "経路種別", "natEnabled": "NAT 有効", "logEnabled": "ログ有効", "schedule": "スケジュール", "natType": "NAT 種別", "externalInterface": "外部インターフェース", "externalIp": "外部 IP", "internalIp": "内部 IP", "translatedSource": "変換後の送信元", "originalSource": "変換前の送信元", "originalDestination": "変換前の宛先", "translatedDestination": "変換後の宛先", "portForward": "ポート転送", "originalPort": "変換前ポート", "translatedPort": "変換後ポート", "poolType": "IP プール種別",
	"portal": "ポータル", "listenPort": "待受ポート", "listenInterface": "待受インターフェース", "tunnelIpPool": "トンネル IP プール", "userGroups": "ユーザーグループ", "mode": "モード", "remoteGateway": "対向ゲートウェイ", "ikeVersion": "IKE バージョン", "proposal": "プロポーザル", "encryption": "暗号化", "authentication": "認証", "dhGroup": "DH グループ", "lifetime": "ライフタイム", "pskConfigured": "事前共有鍵", "phase1Name": "Phase1 名", "localSubnet": "ローカルサブネット", "remoteSubnet": "リモートサブネット", "pfs": "PFS", "profileType": "プロファイル種別", "scanMode": "スキャンモード", "protocols": "対象プロトコル", "categories": "カテゴリ", "appId": "アプリケーション ID", "appName": "アプリケーション名", "category": "カテゴリ", "risk": "リスク", "signatures": "シグネチャ", "startIp": "開始 IP", "endIp": "終了 IP", "netmask": "サブネットマスク", "dnsServers": "DNS サーバー", "excludeIps": "除外 IP", "leaseTime": "リース時間",
	"areaId": "エリア ID", "areaType": "エリア種別", "area": "エリア", "cost": "コスト", "priority": "優先度", "helloInterval": "Hello 間隔", "deadInterval": "Dead 間隔", "networkType": "ネットワーク種別", "passive": "パッシブ", "metric": "メトリック", "metricType": "メトリック種別", "routeMap": "ルートマップ", "routerId": "ルーター ID", "defaultInformationOriginate": "デフォルト経路の広告", "defaultMetric": "デフォルトメトリック", "passiveInterfaces": "パッシブインターフェース", "ip": "IP アドレス", "remoteAs": "対向 AS", "updateSource": "更新元", "ebgpMultihop": "EBGP マルチホップ", "nextHopSelf": "Next Hop Self", "softReconfiguration": "ソフト再構成", "routeMapIn": "受信ルートマップ", "routeMapOut": "送信ルートマップ", "activate": "有効化", "shutdown": "シャットダウン", "prefix": "プレフィックス", "asNumber": "AS 番号", "seqNum": "シーケンス番号", "src": "送信元", "srcNegate": "送信元の否定", "dst": "宛先", "dstNegate": "宛先の否定", "inputDevice": "入力インターフェース", "outputDevice": "出力インターフェース", "comments": "コメント",
	"groupId": "グループ ID", "groupName": "グループ名", "monitorInterfaces": "監視インターフェース", "heartbeatInterfaces": "ハートビートインターフェース", "haMgmtStatus": "HA 管理有効", "preempt": "プリエンプト", "sessionPickup": "セッション引継ぎ", "hbInterval": "ハートビート間隔", "hbLostThreshold": "ハートビート損失閾値", "passwordConfigured": "パスワード", "server": "サーバー", "facility": "ファシリティ", "communityConfigured": "コミュニティ", "hosts": "ホスト", "trapHosts": "トラップ送信先", "version": "バージョン", "fortianalyzerServer": "FortiAnalyzer サーバー", "fortianalyzerStatus": "FortiAnalyzer 状態", "sourceFile": "入力ファイル", "parseErrors": "解析警告", "encoding": "文字コード",
	"managementIp": "管理 IP", "managementNetmask": "管理サブネットマスク", "managementGateway": "管理ゲートウェイ", "domain": "ドメイン", "sourceUsers": "送信元ユーザー", "urlCategories": "URL カテゴリ", "logStart": "開始時ログ", "addressMode": "IP アドレス取得方式",
}

func reportLabel(key string) string {
	if s, ok := reportLabels[key]; ok {
		return s
	}
	return key
}
