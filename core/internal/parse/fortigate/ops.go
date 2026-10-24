package fortigate

import (
	"strings"

	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/util"
)

// convertHA ports WallScribe converters/misc.convert_ha. Raw HA password is never stored.
func convertHA(cfg *model.Config, global Node) {
	ha := global.Map("system ha")
	if ha == nil || len(ha) == 0 {
		return
	}
	mode := parseHAMode(ha.String("mode"))
	priority := ha.String("priority")
	if priority == "" {
		priority = "128"
	}
	hbInterval := ha.String("hb-interval")
	if hbInterval == "" {
		hbInterval = "2"
	}
	hbLost := ha.String("hb-lost-threshold")
	if hbLost == "" {
		hbLost = "6"
	}
	names, detail := parseHBDev(ha)
	cfg.HA = model.HASettings{
		Mode:                      mode,
		GroupID:                   ha.String("group-id"),
		GroupName:                 ha.String("group-name"),
		Priority:                  priority,
		MonitorInterfaces:         ha.List("monitor"),
		HeartbeatInterfaces:       names,
		HeartbeatInterfacesDetail: detail,
		HAMgmtInterfaces:          parseHAMgmt(ha),
		HAMgmtStatus:              ha.String("ha-mgmt-status") == "enable",
		Preempt:                   ha.String("override") == "enable",
		SessionPickup:             ha.String("session-pickup") == "enable",
		HBInterval:                hbInterval,
		HBLostThreshold:           hbLost,
		Encryption:                ha.String("encryption") == "enable",
		Authentication:            ha.String("authentication") == "enable",
		PasswordConfigured:        ha.String("password") != "",
	}
}

func parseHAMode(mode string) string {
	switch mode {
	case "a-p":
		return "active-passive"
	case "a-a":
		return "active-active"
	case "":
		return "standalone"
	default:
		return mode
	}
}

func parseHBDev(ha Node) ([]string, []model.HAHeartbeatInterface) {
	parts := ha.List("hbdev")
	if len(parts) == 0 {
		if s := ha.String("hbdev"); s != "" {
			parts = strings.Fields(strings.ReplaceAll(s, `"`, ""))
		}
	}
	var names []string
	var detail []model.HAHeartbeatInterface
	for i := 0; i < len(parts); {
		iface := strings.Trim(parts[i], `"`)
		prio := ""
		if i+1 < len(parts) && isDigits(parts[i+1]) {
			prio = parts[i+1]
			i += 2
		} else {
			i++
		}
		if iface == "" {
			continue
		}
		names = append(names, iface)
		detail = append(detail, model.HAHeartbeatInterface{Interface: iface, Priority: prio})
	}
	return names, detail
}

func parseHAMgmt(ha Node) []model.HAManagementInterface {
	mgmt := ha.Map("ha-mgmt-interfaces")
	if mgmt == nil {
		return nil
	}
	var out []model.HAManagementInterface
	for id, raw := range mgmt {
		if id == "_name" {
			continue
		}
		mn, ok := raw.(Node)
		if !ok {
			continue
		}
		dst := util.IPToCIDR(joinIPMask(mn.List("dst")))
		if dst == "" {
			dst = util.IPToCIDR(mn.String("dst"))
		}
		out = append(out, model.HAManagementInterface{
			ID:        firstNonEmpty(mn.String("_name"), id),
			Interface: mn.String("interface"),
			Dst:       dst,
			Gateway:   mn.String("gateway"),
		})
	}
	return out
}

func joinIPMask(parts []string) string {
	return strings.Join(parts, " ")
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// convertLogging ports WallScribe converters/misc.convert_logging.
// SNMP community strings are never stored.
func convertLogging(cfg *model.Config, global Node) {
	if syslog := global.Map("log syslogd setting"); syslog != nil && len(syslog) > 0 {
		status := "disabled"
		if syslog.String("status") == "enable" {
			status = "enabled"
		}
		port := syslog.String("port")
		if port == "" {
			port = "514"
		}
		cfg.Logging.SyslogServers = append(cfg.Logging.SyslogServers, model.SyslogServer{
			Server:   syslog.String("server"),
			Port:     port,
			Facility: syslog.String("facility"),
			Status:   status,
		})
	}

	snmpEnabled := false
	if sysinfo := global.Map("system snmp sysinfo"); sysinfo != nil {
		snmpEnabled = sysinfo.String("status") == "enable"
	}
	if communities := global.Map("system snmp community"); communities != nil {
		for name, raw := range communities {
			if name == "_name" {
				continue
			}
			cn, ok := raw.(Node)
			if !ok {
				continue
			}
			hosts := parseSNMPHosts(cn.Map("hosts"))
			trapHosts := []string(nil)
			if cn.String("trap-status") == "enable" {
				trapHosts = hosts
			}
			cfg.Logging.SNMP = append(cfg.Logging.SNMP, model.SNMPSettings{
				Enabled:             snmpEnabled,
				CommunityConfigured: cn.String("name") != "" || name != "",
				Hosts:               hosts,
				TrapHosts:           trapHosts,
				Version:             "v2c",
			})
		}
	}
	if users := global.Map("system snmp user"); users != nil {
		for name, raw := range users {
			if name == "_name" {
				continue
			}
			un, ok := raw.(Node)
			if !ok {
				continue
			}
			hosts := parseSNMPHosts(un.Map("hosts"))
			trapHosts := []string(nil)
			if un.String("trap-status") == "enable" {
				trapHosts = hosts
			}
			cfg.Logging.SNMP = append(cfg.Logging.SNMP, model.SNMPSettings{
				Enabled:   snmpEnabled,
				Hosts:     hosts,
				TrapHosts: trapHosts,
				Version:   "v3",
				Username:  firstNonEmpty(un.String("_name"), name),
			})
		}
	}
	if faz := global.Map("log fortianalyzer setting"); faz != nil && len(faz) > 0 {
		cfg.Logging.FortiAnalyzerServer = faz.String("server")
		if faz.String("status") == "enable" {
			cfg.Logging.FortiAnalyzerStatus = "enabled"
		} else if faz.String("status") != "" {
			cfg.Logging.FortiAnalyzerStatus = "disabled"
		}
	}
}

func parseSNMPHosts(hosts Node) []string {
	if hosts == nil {
		return nil
	}
	var out []string
	for id, raw := range hosts {
		if id == "_name" {
			continue
		}
		hn, ok := raw.(Node)
		if !ok {
			continue
		}
		ip := ""
		if list := hn.List("ip"); len(list) > 0 {
			ip = list[0]
		} else if s := hn.String("ip"); s != "" {
			ip = strings.Fields(s)[0]
		}
		if ip != "" {
			out = append(out, strings.Trim(ip, `"`))
		}
	}
	return out
}
