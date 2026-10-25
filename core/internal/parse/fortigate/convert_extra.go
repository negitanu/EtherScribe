package fortigate

import (
	"sort"
	"strings"

	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/util"
)

// convertObjects ports WallScribe converters/objects.py (address/service + groups).
func objectNames(n Node) []string {
	names := make([]string, 0, len(n))
	for name, raw := range n {
		if _, ok := raw.(Node); ok && name != "_name" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
func objectTags(n Node) []string { return uniqueAppend(n.List("tag"), n.List("tags")) }
func convertObjects(cfg *model.Config, scope Node, vdom string) {
	for _, block := range []string{"firewall address", "firewall address6"} {
		addrs := scope.Map(block)
		for _, name := range objectNames(addrs) {
			an := addrs.Map(name)
			kind := firstNonEmpty(an.String("type"), "subnet")
			value := ""
			switch kind {
			case "fqdn":
				value = an.String("fqdn")
			case "iprange":
				start, end := firstNonEmpty(an.String("start-ip6"), an.String("start-ip")), firstNonEmpty(an.String("end-ip6"), an.String("end-ip"))
				if start != "" && end != "" {
					value = start + "-" + end
				}
			default:
				value = util.IPToCIDR(firstNonEmpty(an.String("ip6"), an.String("subnet"), an.String("agent-id")))
			}
			cfg.Objects.Addresses = append(cfg.Objects.Addresses, model.AddressObject{Name: firstNonEmpty(an.String("_name"), name), Type: kind, Value: value, VDOM: vdom, Description: an.String("comment"), Tags: objectTags(an)})
		}
	}
	for _, block := range []string{"firewall addrgrp", "firewall addrgrp6"} {
		groups := scope.Map(block)
		for _, name := range objectNames(groups) {
			gn := groups.Map(name)
			cfg.Objects.AddressGroups = append(cfg.Objects.AddressGroups, model.NamedMembers{Name: firstNonEmpty(gn.String("_name"), name), Members: gn.List("member"), VDOM: vdom, Description: gn.String("comment"), Tags: objectTags(gn)})
		}
	}
	services := scope.Map("firewall service custom")
	for _, name := range objectNames(services) {
		sn := services.Map(name)
		ports := []string{strings.Join(sn.List("tcp-portrange"), ", "), strings.Join(sn.List("udp-portrange"), ", "), strings.Join(sn.List("sctp-portrange"), ", ")}
		var protocols, values []string
		for i, p := range ports {
			if p != "" {
				protocols = append(protocols, []string{"TCP", "UDP", "SCTP"}[i])
				values = uniqueAppend(values, []string{p})
			}
		}
		var icmp []string
		if v := sn.String("icmptype"); v != "" {
			icmp = append(icmp, "type="+v)
		}
		if v := sn.String("icmpcode"); v != "" {
			icmp = append(icmp, "code="+v)
		}
		if len(icmp) > 0 {
			protocols = append(protocols, "ICMP")
			values = append(values, strings.Join(icmp, ","))
		}
		if len(protocols) == 0 {
			protocols = []string{firstNonEmpty(strings.ToUpper(sn.String("protocol")), "IP")}
		}
		cfg.Objects.Services = append(cfg.Objects.Services, model.ServiceObject{Name: firstNonEmpty(sn.String("_name"), name), Protocol: strings.Join(protocols, "/"), Port: strings.Join(values, ", "), TCPPort: ports[0], UDPPort: ports[1], SCTPPort: ports[2], VDOM: vdom, Description: sn.String("comment"), Tags: objectTags(sn)})
	}
	groups := scope.Map("firewall service group")
	for _, name := range objectNames(groups) {
		gn := groups.Map(name)
		cfg.Objects.ServiceGroups = append(cfg.Objects.ServiceGroups, model.NamedMembers{Name: firstNonEmpty(gn.String("_name"), name), Members: gn.List("member"), VDOM: vdom, Description: gn.String("comment"), Tags: objectTags(gn)})
	}
}

// convertNAT ports WallScribe convert_nat (VIP + IP pool).
func convertNAT(cfg *model.Config, scope Node, vdom string) {
	if vips := scope.Map("firewall vip"); vips != nil {
		for name, raw := range vips {
			if name == "_name" {
				continue
			}
			vn, ok := raw.(Node)
			if !ok {
				continue
			}
			mapped := vn.String("mappedip")
			if vals := vn.List("mappedip"); len(vals) > 1 {
				mapped = strings.Join(vals, ", ")
			}
			cfg.NATPolicies = append(cfg.NATPolicies, model.NATPolicy{
				Name: firstNonEmpty(vn.String("_name"), name), NATType: "vip",
				ExternalInterface: vn.String("extintf"), ExternalIP: vn.String("extip"),
				InternalIP: mapped, OriginalDestination: vn.String("extip"),
				TranslatedDestination: mapped, Interface: vn.String("extintf"),
				PortForward:  vn.String("portforward") == "enable",
				OriginalPort: vn.String("extport"), TranslatedPort: vn.String("mappedport"),
				VDOM: vdom, Description: vn.String("comment"),
			})
		}
	}
	if pools := scope.Map("firewall ippool"); pools != nil {
		for name, raw := range pools {
			if name == "_name" {
				continue
			}
			pn, ok := raw.(Node)
			if !ok {
				continue
			}
			start, end := pn.String("startip"), pn.String("endip")
			translated := start
			if start != "" && end != "" && start != end {
				translated = start + "-" + end
			}
			cfg.NATPolicies = append(cfg.NATPolicies, model.NATPolicy{
				Name: firstNonEmpty(pn.String("_name"), name), NATType: "ippool",
				PoolType:         firstNonEmpty(pn.String("type"), "overload"),
				TranslatedSource: translated, Interface: pn.String("associated-interface"),
				VDOM: vdom, Description: pn.String("comments"),
			})
		}
	}
}

// convertVPN ports WallScribe converters/vpn.py (IPsec phase1/2). PSK raw values are not stored.
func convertVPN(cfg *model.Config, scope Node, vdom string) {
	addPhase1 := func(section string) {
		entries := scope.Map(section)
		if entries == nil {
			return
		}
		for name, raw := range entries {
			if name == "_name" {
				continue
			}
			pn, ok := raw.(Node)
			if !ok {
				continue
			}
			proposal := pn.String("proposal")
			if list := pn.List("proposal"); len(list) > 0 {
				proposal = strings.Join(list, ", ")
			}
			enc, auth := parseProposal(proposal)
			cfg.VPN.IPSecPhase1 = append(cfg.VPN.IPSecPhase1, model.IPSecPhase1{
				Name:          firstNonEmpty(pn.String("_name"), name),
				RemoteGateway: pn.String("remote-gw"), Interface: pn.String("interface"),
				IKEVersion: pn.String("ike-version"), Proposal: proposal,
				Encryption: enc, Authentication: auth, DHGroup: pn.String("dhgrp"),
				Lifetime:      pn.String("keylife"),
				PSKConfigured: pn.String("psksecret") != "",
				VDOM:          vdom,
			})
		}
	}
	addPhase2 := func(section string) {
		entries := scope.Map(section)
		if entries == nil {
			return
		}
		for name, raw := range entries {
			if name == "_name" {
				continue
			}
			pn, ok := raw.(Node)
			if !ok {
				continue
			}
			proposal := pn.String("proposal")
			if list := pn.List("proposal"); len(list) > 0 {
				proposal = strings.Join(list, ", ")
			}
			enc, auth := parseProposal(proposal)
			cfg.VPN.IPSecPhase2 = append(cfg.VPN.IPSecPhase2, model.IPSecPhase2{
				Name: firstNonEmpty(pn.String("_name"), name), Phase1Name: pn.String("phase1name"),
				Proposal: proposal, Encryption: enc, Authentication: auth,
				PFS: pn.String("pfs"), Lifetime: pn.String("keylifeseconds"),
				LocalSubnet:  util.IPToCIDR(pn.String("src-subnet")),
				RemoteSubnet: util.IPToCIDR(pn.String("dst-subnet")),
				VDOM:         vdom,
			})
		}
	}
	addPhase1("vpn ipsec phase1-interface")
	addPhase1("vpn ipsec phase1")
	addPhase2("vpn ipsec phase2-interface")
	addPhase2("vpn ipsec phase2")
	addSSLVPN(cfg, scope, vdom)
}

func addSSLVPN(cfg *model.Config, scope Node, vdom string) {
	if settings := scope.Map("vpn ssl settings"); settings != nil && len(settings) > 0 {
		pools := settings.String("tunnel-ip-pools")
		if list := settings.List("tunnel-ip-pools"); len(list) > 0 {
			pools = strings.Join(list, ", ")
		}
		var groups []string
		if rules := settings.Map("authentication-rule"); rules != nil {
			for id, raw := range rules {
				if id == "_name" {
					continue
				}
				rn, ok := raw.(Node)
				if !ok {
					continue
				}
				groups = append(groups, rn.List("groups")...)
			}
		}
		port := settings.String("port")
		if port == "" {
			port = "443"
		}
		cfg.VPN.SSLVPN = append(cfg.VPN.SSLVPN, model.SSLVPNSettings{
			ListenPort:      port,
			ListenInterface: settings.String("source-interface"),
			TunnelIPPool:    pools,
			UserGroups:      groups,
			Mode:            "tunnel",
			VDOM:            vdom,
		})
	}
	if portals := scope.Map("vpn ssl web portal"); portals != nil {
		for name, raw := range portals {
			if name == "_name" {
				continue
			}
			pn, ok := raw.(Node)
			if !ok {
				continue
			}
			if pn.String("tunnel-mode") == "" && pn.String("web-mode") == "" {
				continue
			}
			mode := "web"
			if pn.String("tunnel-mode") == "enable" {
				mode = "tunnel"
			} else if pn.String("web-mode") == "enable" {
				mode = "web"
			}
			cfg.VPN.SSLVPN = append(cfg.VPN.SSLVPN, model.SSLVPNSettings{
				Portal: firstNonEmpty(pn.String("_name"), name),
				Mode:   mode,
				VDOM:   vdom,
			})
		}
	}
}

func parseProposal(proposals string) (encryption, authentication string) {
	proposals = strings.TrimSpace(proposals)
	if proposals == "" {
		return "", ""
	}
	// WallScribe: first token before '-' is enc, second is auth (aes256-sha256)
	first := strings.Split(proposals, ",")[0]
	first = strings.TrimSpace(first)
	parts := strings.SplitN(first, "-", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return first, ""
}

func parseAction(action string) string {
	action = strings.TrimSpace(action)
	// WallScribe: empty / unset → deny
	if action == "" || action == "unset" {
		return "deny"
	}
	if action == "accept" {
		return "allow"
	}
	return action
}

func extractSecurityProfiles(pn Node) []string {
	var out []string
	if v := pn.String("profile-group"); v != "" {
		out = append(out, "Group:"+v)
	}
	if v := pn.String("av-profile"); v != "" {
		out = append(out, "AV:"+v)
	}
	if v := pn.String("webfilter-profile"); v != "" {
		out = append(out, "Web:"+v)
	}
	if v := pn.String("application-list"); v != "" {
		out = append(out, "App:"+v)
	}
	if v := pn.String("ips-sensor"); v != "" {
		out = append(out, "IPS:"+v)
	}
	if v := pn.String("ssl-ssh-profile"); v != "" {
		out = append(out, "SSL:"+v)
	}
	return out
}
