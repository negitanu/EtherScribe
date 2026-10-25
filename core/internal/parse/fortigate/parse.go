package fortigate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/negitanu/etherscribe/internal/identify"
	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/util"
)

// Parse converts FortiGate .conf text into the unified model (MVP subset).
func Parse(content, filename string) (*model.Config, error) {
	cfg := &model.Config{
		SourceFile: filename,
		DeviceInfo: model.DeviceInfo{
			DeviceType: identify.FortiGate, Role: model.RoleFirewall, VendorFamily: "fortigate",
			OperationMode: "Unknown",
		},
		System:  model.SystemSettings{HTTPSPort: "443", SSHPort: "22"},
		Objects: model.Objects{},
	}

	lines, sepErrs := SeparateLogicalLines(content)
	cfg.ParseErrors = append(cfg.ParseErrors, sepErrs...)

	header := Node{}
	var body []string
	for _, line := range lines {
		if strings.HasPrefix(line, "#config-version=") {
			parseHeaderLine(line, header)
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		body = append(body, line)
	}

	tree, treeErrs := ParseLines(body)
	cfg.ParseErrors = append(cfg.ParseErrors, treeErrs...)

	globalBlock := tree.Map("global")
	vdomBlock := tree.Map("vdom")
	global := Node{}
	for k, v := range tree {
		if k == "global" || k == "vdom" {
			continue
		}
		global[k] = v
	}
	if globalBlock != nil {
		for k, v := range globalBlock {
			global[k] = v
		}
	}
	if vdomBlock == nil {
		vdomBlock = Node{}
	}

	applyHeader(cfg, header)
	if id := identify.Identify(content); id != nil {
		if cfg.DeviceInfo.Model == "" {
			cfg.DeviceInfo.Model = id.Model
		}
		if cfg.DeviceInfo.OSVersion == "" {
			cfg.DeviceInfo.OSVersion = id.OSVersion
		}
	}

	convertDevice(cfg, global)
	convertSystem(cfg, global)
	convertInterfaces(cfg, global)
	convertRoutes(cfg, global, "root")
	convertPolicies(cfg, global, "root")
	convertObjects(cfg, global, "root")
	convertNAT(cfg, global, "root")
	convertVPN(cfg, global, "root")
	convertSecurity(cfg, global, "root")
	convertDHCP(cfg, global, "root")
	convertOSPF(cfg, global, "root")
	convertBGP(cfg, global, "root")
	convertPolicyRoutes(cfg, global, "root")
	convertHA(cfg, global)
	convertLogging(cfg, global)

	var vdoms []string
	for _, name := range EditNames(vdomBlock) {
		raw := vdomBlock[name]
		if name == "_name" {
			continue
		}
		vn, ok := raw.(Node)
		if !ok {
			continue
		}
		vdoms = append(vdoms, name)
		convertRoutes(cfg, vn, name)
		convertPolicies(cfg, vn, name)
		convertObjects(cfg, vn, name)
		convertNAT(cfg, vn, name)
		convertVPN(cfg, vn, name)
		convertSecurity(cfg, vn, name)
		convertDHCP(cfg, vn, name)
		convertOSPF(cfg, vn, name)
		convertBGP(cfg, vn, name)
		convertPolicyRoutes(cfg, vn, name)
		convertSyslogSettings(cfg, vn, name)
	}
	sort.Strings(vdoms)
	if len(vdoms) > 0 {
		cfg.DeviceInfo.VDOMList = vdoms
		cfg.DeviceInfo.VDOMEnabled = true
	} else if len(cfg.DeviceInfo.VDOMList) == 0 {
		cfg.DeviceInfo.VDOMList = []string{"root"}
	}

	// Coverage: top-level config keys we did not convert deeply
	known := map[string]struct{}{
		"system global": {}, "system dns": {}, "system ntp": {}, "system admin": {},
		"system interface": {}, "system settings": {}, "system vdom-link": {},
		"router static": {}, "router static6": {}, "firewall policy": {},
		"firewall address": {}, "firewall address6": {}, "firewall addrgrp6": {}, "firewall addrgrp": {}, "firewall service custom": {},
		"firewall service group": {}, "firewall vip": {}, "firewall ippool": {},
		"vpn ipsec phase1-interface": {}, "vpn ipsec phase1": {},
		"vpn ipsec phase2-interface": {}, "vpn ipsec phase2": {},
		"antivirus profile": {}, "webfilter profile": {}, "ips sensor": {},
		"firewall ssl-ssh-profile": {}, "application list": {},
		"system ha": {}, "log syslogd setting": {}, "log fortianalyzer setting": {},
		"system snmp sysinfo": {}, "system snmp community": {}, "system snmp user": {},
		"vpn ssl settings": {}, "vpn ssl web portal": {},
		"system dhcp server": {}, "router ospf": {}, "router ospf6": {},
		"router bgp": {}, "router policy": {},
	}
	coverageCounts := map[string]int{}
	inspect := func(scope Node, prefix string) {
		for key, raw := range scope {
			if _, ok := raw.(Node); !ok {
				continue
			}
			if _, ok := known[key]; !ok {
				coverageCounts[prefix+key]++
			}
		}
	}
	for _, name := range []string{"syslogd", "syslogd2", "syslogd3", "syslogd4"} {
		for _, suffix := range []string{"setting", "override-setting"} {
			known["log "+name+" "+suffix] = struct{}{}
		}
	}
	inspect(global, "global/")

	for _, name := range EditNames(vdomBlock) {
		inspect(vdomBlock.Map(name), "vdom/*/")
	}
	paths := make([]string, 0, len(coverageCounts))
	for path := range coverageCounts {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		cfg.Coverage = append(cfg.Coverage, model.CoverageItem{Path: path, Reason: "未対応または部分対応の設定ブロック", Note: fmt.Sprintf("出現数 %d（設定値は非表示）", coverageCounts[path])})
	}
	return cfg, nil
}

func parseHeaderLine(line string, header Node) {
	id := identify.Identify(line + "\n")
	if id != nil {
		header["model"] = id.Model
		header["version"] = id.OSVersion
	}
	opmode, vdom, build := identify.ParseHeaderExtras(line)
	if opmode != "" {
		header["opmode"] = opmode
	}
	header["vdom_enabled"] = vdom
	if build != "" {
		header["build"] = build
	}
}

func applyHeader(cfg *model.Config, header Node) {
	if m, ok := header["model"].(string); ok {
		cfg.DeviceInfo.Model = m
	}
	if v, ok := header["version"].(string); ok && v != "" {
		cfg.DeviceInfo.OSVersion = v
		cfg.DeviceInfo.OSVersionSource = "config-version header"
	}
	if op, ok := header["opmode"].(string); ok {
		if op == "0" {
			cfg.DeviceInfo.OperationMode = "NAT/Route"
		} else {
			cfg.DeviceInfo.OperationMode = "Transparent"
		}
	}
	if ve, ok := header["vdom_enabled"].(bool); ok {
		cfg.DeviceInfo.VDOMEnabled = ve
	}
}

func convertDevice(cfg *model.Config, global Node) {
	sg := global.Map("system global")
	if sg == nil {
		return
	}
	cfg.DeviceInfo.Hostname = strings.Trim(sg.String("hostname"), `"`)
}

func convertSystem(cfg *model.Config, global Node) {
	if dns := global.Map("system dns"); dns != nil {
		cfg.System.DNSPrimary = dns.String("primary")
		cfg.System.DNSSecondary = dns.String("secondary")
	}
	if sg := global.Map("system global"); sg != nil {
		cfg.System.Timezone = sg.String("timezone")
		if p := sg.String("admin-https-port"); p != "" {
			cfg.System.HTTPSPort = p
		}
		if p := sg.String("admin-ssh-port"); p != "" {
			cfg.System.SSHPort = p
		}
	}
	if ntp := global.Map("system ntp"); ntp != nil {
		if servers := ntp.Map("ntpserver"); servers != nil {
			for _, raw := range servers {
				if sn, ok := raw.(Node); ok {
					if s := sn.String("server"); s != "" {
						cfg.System.NTPServers = append(cfg.System.NTPServers, strings.Trim(s, `"`))
					}
				}
			}
		}
	}
	if admins := global.Map("system admin"); admins != nil {
		for name, raw := range admins {
			if name == "_name" {
				continue
			}
			an, ok := raw.(Node)
			if !ok {
				continue
			}
			u := model.Admin{
				Username: an.String("_name"),
				Profile:  an.String("accprofile"),
				VDOM:     an.String("vdom"),
				Password: an.String("password"),
			}
			if u.Username == "" {
				u.Username = name
			}
			cfg.System.AdminUsers = append(cfg.System.AdminUsers, u)
		}
	}
}

func convertInterfaces(cfg *model.Config, global Node) {
	ifaces := global.Map("system interface")
	if ifaces == nil {
		return
	}
	for _, name := range EditNames(ifaces) {
		raw := ifaces[name]
		if name == "_name" {
			continue
		}
		in, ok := raw.(Node)
		if !ok {
			continue
		}
		vlan := in.String("vlanid")
		itype := in.String("type")
		if itype == "" {
			itype = "physical"
		}
		if vlan != "" {
			itype = "vlan"
		}
		iface := model.Interface{
			Name:            firstNonEmpty(in.String("_name"), name),
			Alias:           in.String("alias"),
			InterfaceType:   itype,
			ParentInterface: in.String("interface"),
			VLANID:          vlan,
			VDOM:            firstNonEmpty(in.String("vdom"), "root"),
			Role:            in.String("role"),
			Description:     in.String("description"),
			Status:          "up",
		}
		if in.String("status") == "down" {
			iface.Status = "down"
		}
		if ip := in.String("ip"); ip != "" {
			iface.IPAddress = util.IPToCIDR(ip)
		}
		enrichInterface(cfg, &iface, in)
		iface.AllowedAccess = in.List("allowaccess")
		cfg.Interfaces = append(cfg.Interfaces, iface)
	}
	for _, lag := range cfg.LAGs {
		for _, member := range lag.Members {
			for i := range cfg.Interfaces {
				if cfg.Interfaces[i].Name == member {
					cfg.Interfaces[i].LAGParent = lag.Name
				}
			}
		}
	}

}

func convertRoutes(cfg *model.Config, scope Node, vdom string) {
	add := func(section, routeType string) {
		routes := scope.Map(section)
		if routes == nil {
			return
		}
		for _, id := range EditNames(routes) {
			raw := routes[id]
			if id == "_name" {
				continue
			}
			rn, ok := raw.(Node)
			if !ok {
				continue
			}
			r := model.Route{
				Name:        firstNonEmpty(rn.String("_name"), id),
				Destination: util.IPToCIDR(rn.String("dst")),
				Gateway:     rn.String("gateway"),
				Interface:   rn.String("device"),
				Distance:    rn.String("distance"),
				VDOM:        vdom,
				Context:     vdom,
				RouteType:   routeType,
				Enabled:     rn.String("status") != "disable",
			}
			r.RoutingContext = rn.String("vrf")
			r.DestinationObject = rn.String("dstaddr")
			r.Priority = rn.String("priority")
			r.Description = rn.String("comment")
			r.Action = "forward"
			if rn.String("blackhole") == "enable" {
				r.Action = "discard"
			}
			if r.Destination == "" && r.DestinationObject == "" {
				if section == "router static6" {
					r.Destination = "::/0"
				} else {
					r.Destination = "0.0.0.0/0"
				}
			}
			cfg.Routes = append(cfg.Routes, r)
		}
	}
	add("router static", "static")
	add("router static6", "static6")
}

func convertPolicies(cfg *model.Config, scope Node, vdom string) {
	policies := scope.Map("firewall policy")
	if policies == nil {
		return
	}
	for _, id := range EditNames(policies) {
		raw := policies[id]
		if id == "_name" {
			continue
		}
		pn, ok := raw.(Node)
		if !ok {
			continue
		}
		log := pn.String("logtraffic")
		p := model.Policy{
			PolicyID:             firstNonEmpty(pn.String("policyid"), id),
			Name:                 pn.String("name"),
			SourceInterface:      pn.List("srcintf"),
			DestinationInterface: pn.List("dstintf"),
			SourceAddress:        pn.List("srcaddr"),
			DestinationAddress:   pn.List("dstaddr"),
			Service:              pn.List("service"),
			Action:               parseAction(pn.String("action")),
			NATEnabled:           pn.String("nat") == "enable",
			LogStart:             pn.String("logtraffic-start") == "enable",
			LogEnabled:           log == "enable" || log == "all" || log == "utm",
			SecurityProfiles:     extractSecurityProfiles(pn),
			VDOM:                 vdom,
			Enabled:              pn.String("status") != "disable",
			Description:          pn.String("comments"),
			Schedule:             pn.String("schedule"),
		}
		cfg.Policies = append(cfg.Policies, p)
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
