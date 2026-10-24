package export

// Section keys shared by CLI / GUI / JSON.
var SectionKeys = []string{
	"device_info", "system_settings", "network", "vlans", "lag", "stp",
	"objects", "policies", "nat", "vpn", "security_profiles", "ha", "logging",
}

type Preset struct {
	Key         string
	Label       string
	Description string
	Sections    []string
}

var Presets = map[string]Preset{
	"overview": {
		Key: "overview", Label: "概要", Description: "ホスト・IF・ポリシー",
		Sections: []string{"device_info", "network", "policies"},
	},
	"connectivity": {
		Key: "connectivity", Label: "接続", Description: "IF・経路・NAT・VPN",
		Sections: []string{"device_info", "network", "nat", "vpn"},
	},
	"policy": {
		Key: "policy", Label: "ポリシー監査", Description: "オブジェクト・ルール・NAT",
		Sections: []string{"device_info", "objects", "policies", "nat", "security_profiles"},
	},
	"ops": {
		Key: "ops", Label: "運用", Description: "運用引き継ぎ向け",
		Sections: []string{"device_info", "system_settings", "network", "objects", "policies", "nat", "vpn", "security_profiles", "ha", "logging"},
	},
	"all": {
		Key: "all", Label: "すべて", Description: "全項目",
		Sections: append([]string{}, SectionKeys...),
	},
	"switch-overview": {
		Key: "switch-overview", Label: "スイッチ概要", Description: "ホスト・IF・VLAN",
		Sections: []string{"device_info", "network", "vlans"},
	},
	"switch-l2": {
		Key: "switch-l2", Label: "スイッチ L2", Description: "VLAN・LAG・STP",
		Sections: []string{"device_info", "network", "vlans", "lag", "stp"},
	},
	"switch-l3": {
		Key: "switch-l3", Label: "スイッチ L3", Description: "IRB・経路・VRF",
		Sections: []string{"device_info", "network", "vlans"},
	},
	"switch-ops": {
		Key: "switch-ops", Label: "スイッチ運用", Description: "運用引き継ぎ向け",
		Sections: []string{"device_info", "system_settings", "network", "vlans", "lag", "stp", "ha", "logging"},
	},
	"switch-all": {
		Key: "switch-all", Label: "スイッチすべて", Description: "スイッチ向け全項目（FW除外）",
		Sections: []string{"device_info", "system_settings", "network", "vlans", "lag", "stp", "ha", "logging"},
	},
}

func ResolveSections(preset string, sections []string) ([]string, error) {
	if len(sections) > 0 {
		return sections, nil
	}
	if preset == "" {
		return nil, nil // all
	}
	p, ok := Presets[preset]
	if !ok {
		return nil, ErrUnknownPreset(preset)
	}
	return append([]string{}, p.Sections...), nil
}

type ErrUnknownPreset string

func (e ErrUnknownPreset) Error() string { return "未知のプリセットです: " + string(e) }

func SectionSelected(sections []string, key string) bool {
	if len(sections) == 0 {
		return true
	}
	for _, s := range sections {
		if s == key {
			return true
		}
		if s == "device" && (key == "device_info" || key == "system_settings" || key == "ha" || key == "logging") {
			return true
		}
		if s == "policy" && (key == "policies" || key == "nat") {
			return true
		}
	}
	return false
}
