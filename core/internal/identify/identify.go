// Package identify ports WallScribe parsers/device_identification.py and
// TextFSM header patterns as Go regular expressions.
package identify

import (
	"regexp"
	"strings"
)

type DeviceType string

const (
	FortiGate DeviceType = "FortiGate"
	PaloAlto  DeviceType = "PaloAlto"
	JunosEX   DeviceType = "JunosEX"
	Unknown   DeviceType = "Unknown"
)

type Identified struct {
	Type      DeviceType
	Model     string
	OSVersion string
}

// From fortigate_header.textfsm / header_config_version.textfsm
var (
	reFGIdentify = regexp.MustCompile(
		`(?m)^#config-version=(?P<prefix>FGT|FG)(?P<code>[A-Z0-9]+)-(?P<ver>\d+(?:\.\d+){1,3})-FW`,
	)
	// Fallback from FortiGateParser._parse_header_line
	reFGHeaderFallback = regexp.MustCompile(
		`#config-version=(?P<model>[A-Z0-9]+)-(?P<version>\d+(?:\.\d+){1,3})-FW(?:-build(?P<build>\d+))?`,
	)
	reOpmode = regexp.MustCompile(`:opmode=(\d+)`)
	reVDOM   = regexp.MustCompile(`:vdom=(\d+)`)
	// Compact model code: 60F1 → FortiGate 6001F style (WallScribe format_fortigate_display_model)
	reCompactModel = regexp.MustCompile(`^(\d{2})([A-Z])(\d)$`)

	// paloalto_config_root.textfsm
	rePAConfigRoot = regexp.MustCompile(`(?i)<config\s+[^>]*version="([^"]+)"`)
	// paloalto_set_cli probe (simplified from paloalto_text._SET_CLI_PROBE)
	rePASetCLI = regexp.MustCompile(`(?m)^set\s+(deviceconfig|shared|vsys|network|rulebase)\b`)

	reJunosSetHost = regexp.MustCompile(`(?m)^set\s+system\s+host-name\s+`)
	reJunosSetL2   = regexp.MustCompile(`(?m)^set\s+(vlans|interfaces)\s+`)
	reJunosEthSW   = regexp.MustCompile(`(?i)ethernet-switching`)
	reJunosEXModel = regexp.MustCompile(`(?i)\b(EX\d{2,4}[A-Z0-9-]*)\b`)
	reJunosVersion = regexp.MustCompile(`(?m)^(?:set\s+)?version\s+(\S+);?\s*$`)
)

// FormatFortiGateDisplayModel ports format_fortigate_display_model.
func FormatFortiGateDisplayModel(modelCode string) string {
	if modelCode == "" {
		return "FortiGate"
	}
	if m := reCompactModel.FindStringSubmatch(modelCode); m != nil {
		return "FortiGate " + m[1] + "0" + m[3] + m[2]
	}
	return "FortiGate " + modelCode
}

func Identify(content string) *Identified {
	sample := content
	if len(sample) > 65536 {
		sample = sample[:65536]
	}

	if m := reFGIdentify.FindStringSubmatch(sample); m != nil {
		return &Identified{
			Type:      FortiGate,
			Model:     FormatFortiGateDisplayModel(m[2]),
			OSVersion: m[3],
		}
	}
	if strings.Contains(sample, "#config-version=") || strings.Contains(sample, "config system global") {
		if m := reFGHeaderFallback.FindStringSubmatch(sample); m != nil {
			code := m[1]
			switch {
			case strings.HasPrefix(code, "FGT"):
				code = code[3:]
			case strings.HasPrefix(code, "FG"):
				code = code[2:]
			}
			return &Identified{
				Type:      FortiGate,
				Model:     FormatFortiGateDisplayModel(code),
				OSVersion: m[2],
			}
		}
		return &Identified{Type: FortiGate, Model: "FortiGate"}
	}

	if rePASetCLI.MatchString(sample) {
		return &Identified{Type: PaloAlto, Model: "PA Series"}
	}
	if m := rePAConfigRoot.FindStringSubmatch(sample); m != nil {
		return &Identified{Type: PaloAlto, Model: "PA Series", OSVersion: m[1]}
	}
	if strings.Contains(sample, "<config ") && (strings.Contains(sample, "<devices>") || strings.Contains(sample, "<vsys>")) {
		return &Identified{Type: PaloAlto, Model: "PA Series"}
	}

	if id := identifyJunosEX(sample); id != nil {
		return id
	}
	return nil
}

func identifyJunosEX(sample string) *Identified {
	hasHost := reJunosSetHost.MatchString(sample) ||
		(strings.Contains(sample, "system {") && strings.Contains(sample, "host-name"))
	hasL2 := reJunosEthSW.MatchString(sample) ||
		strings.Contains(sample, "set vlans ") ||
		strings.Contains(sample, "vlans {") ||
		(reJunosSetL2.MatchString(sample) && reJunosEthSW.MatchString(sample))
	// Accept Junos switch configs that look like EX (L2 or explicit EX model).
	modelHint := "EX Series"
	if m := reJunosEXModel.FindStringSubmatch(sample); m != nil {
		modelHint = m[1]
		hasL2 = true
	}
	if !hasHost && !strings.Contains(sample, "set interfaces ") && !strings.Contains(sample, "interfaces {") {
		return nil
	}
	if !hasL2 && !hasHost {
		return nil
	}
	// Prefer Junos when set system host-name exists (not Palo Alto set CLI).
	if reJunosSetHost.MatchString(sample) || (strings.Contains(sample, "system {") && strings.Contains(sample, "host-name")) {
		id := &Identified{Type: JunosEX, Model: modelHint}
		if m := reJunosVersion.FindStringSubmatch(sample); m != nil {
			id.OSVersion = strings.TrimSuffix(m[1], ";")
		}
		return id
	}
	if hasL2 && (strings.Contains(sample, "set interfaces ") || strings.Contains(sample, "interfaces {")) {
		return &Identified{Type: JunosEX, Model: modelHint}
	}
	return nil
}

// ParseHeaderExtras extracts opmode / vdom flags from a #config-version line.
func ParseHeaderExtras(line string) (opmode string, vdomEnabled bool, build string) {
	if m := reOpmode.FindStringSubmatch(line); m != nil {
		opmode = m[1]
	}
	if m := reVDOM.FindStringSubmatch(line); m != nil {
		vdomEnabled = m[1] == "1"
	}
	if m := reFGHeaderFallback.FindStringSubmatch(line); m != nil && m[3] != "" {
		build = m[3]
	}
	return
}
