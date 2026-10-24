package export

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/negitanu/etherscribe/internal/model"
)

var reportOrder = []string{"deviceInfo", "virtualContexts", "systemSettings", "interfaces", "vlans", "lags", "stp", "routes", "routing", "dhcpServers", "objects", "firewallPolicies", "natPolicies", "vpn", "securityProfiles", "securityProfilesDetail", "ha", "logging", "coverage"}
var reportLabels = map[string]string{
	"deviceInfo": "機器概要", "virtualContexts": "仮想コンテキスト", "systemSettings": "システム設定", "interfaces": "インターフェース", "vlans": "VLAN", "lags": "LAG", "stp": "STP", "routes": "ルート", "routing": "動的ルーティング", "dhcpServers": "DHCP", "objects": "オブジェクト", "firewallPolicies": "ポリシー", "natPolicies": "NAT", "vpn": "VPN", "securityProfiles": "セキュリティプロファイル", "securityProfilesDetail": "セキュリティ詳細", "ha": "HA", "logging": "ログ・監視", "coverage": "解析範囲",
	"hostname": "ホスト名", "model": "機種", "osVersion": "OS バージョン", "name": "名前", "deviceType": "機器種別", "role": "役割", "interfaceType": "種別", "ipAddress": "IP アドレス", "context": "コンテキスト", "vdom": "VDOM / vsys", "status": "状態", "description": "説明", "destination": "宛先", "gateway": "ゲートウェイ", "interface": "インターフェース", "switchportMode": "ポートモード", "accessVlan": "アクセス VLAN", "trunkVlans": "トランク VLAN", "nativeVlan": "ネイティブ VLAN", "lagParent": "所属 LAG", "vlanId": "VLAN ID", "members": "メンバー", "lacpEnabled": "LACP 有効", "policyId": "ID", "sourceAddress": "送信元アドレス", "destinationAddress": "宛先アドレス", "sourceInterface": "送信元 IF", "destinationInterface": "宛先 IF", "service": "サービス", "action": "動作", "path": "パス", "reason": "理由", "note": "備考",
}

func reportLabel(key string) string {
	if s, ok := reportLabels[key]; ok {
		return s
	}
	return key
}

// reportTables uses the same selected, scrubbed model as JSON so no format silently drops fields.
func reportTables(cfg *model.Config, sections []string) ([]sheetData, error) {
	data, err := RenderJSON(cfg, sections)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	var tables []sheetData
	for _, key := range reportOrder {
		value := doc[key]
		section := map[string]string{"deviceInfo": "device_info", "virtualContexts": "device_info", "systemSettings": "system_settings", "interfaces": "network", "routes": "network", "routing": "network", "dhcpServers": "network", "vlans": "vlans", "lags": "lag", "stp": "stp", "objects": "objects", "firewallPolicies": "policies", "natPolicies": "nat", "vpn": "vpn", "securityProfiles": "security_profiles", "securityProfilesDetail": "security_profiles", "ha": "ha", "logging": "logging"}[key]
		if key != "coverage" && !SectionSelected(sections, section) {
			continue
		}
		rows := [][]string{{"項目", "値"}}
		if records, ok := value.([]any); ok && len(records) > 0 {
			columns := map[string]bool{}
			flattened := make([]map[string]string, 0, len(records))
			for _, record := range records {
				flat := map[string]string{}
				flattenReport("", record, flat)
				flattened = append(flattened, flat)
				for k := range flat {
					columns[k] = true
				}
			}
			keys := make([]string, 0, len(columns))
			for k := range columns {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			// Put identifying fields first while retaining deterministic order for all remaining fields.
			for i := len([]string{"name", "policyId", "path"}) - 1; i >= 0; i-- {
				priority := []string{"name", "policyId", "path"}[i]
				for j, k := range keys {
					if k == priority {
						keys = append([]string{k}, append(keys[:j], keys[j+1:]...)...)
						break
					}
				}
			}
			header := make([]string, len(keys))
			for i, k := range keys {
				header[i] = reportLabel(k)
			}
			rows = [][]string{header}
			for _, flat := range flattened {
				row := make([]string, len(keys))
				for i, k := range keys {
					row[i] = flat[k]
				}
				rows = append(rows, row)
			}
		} else {
			flat := map[string]string{}
			flattenReport("", value, flat)
			keys := make([]string, 0, len(flat))
			for k := range flat {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				rows = append(rows, []string{reportLabel(k), flat[k]})
			}
		}
		if len(rows) == 1 {
			rows = append(rows, []string{"該当する設定はありません", "—"})
		}
		tables = append(tables, sheetData{Name: reportLabel(key), Rows: rows})
	}
	if len(cfg.ParseErrors) > 0 {
		tables = append(tables, sheetData{Name: "解析警告", Rows: [][]string{{"項目", "件数"}, {"解析警告（設定値を含む可能性があるため本文は非表示）", fmt.Sprint(len(cfg.ParseErrors))}}})
	}
	return tables, nil
}
func flattenReport(prefix string, value any, out map[string]string) {
	switch v := value.(type) {
	case map[string]any:
		for k, child := range v {
			path := k
			if prefix != "" {
				path = prefix + "." + k
			}
			flattenReport(path, child, out)
		}
	case []any:
		scalar := true
		for _, child := range v {
			switch child.(type) {
			case map[string]any, []any:
				scalar = false
			}
		}
		if scalar {
			parts := make([]string, len(v))
			for i, child := range v {
				parts[i] = fmt.Sprint(child)
			}
			out[prefix] = strings.Join(parts, ", ")
		} else {
			for i, child := range v {
				flattenReport(fmt.Sprintf("%s[%d]", prefix, i+1), child, out)
			}
		}
	case nil:
	default:
		out[prefix] = fmt.Sprint(v)
	}
}
