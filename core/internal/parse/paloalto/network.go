package paloalto

import (
	"github.com/negitanu/etherscribe/internal/model"
)

func convertSystemSettings(cfg *model.Config, s systemCfg) {
	cfg.System.ManagementIP = s.ManagementIP
	cfg.System.ManagementNetmask = s.Netmask
	cfg.System.ManagementGateway = s.Gateway
	cfg.System.Domain = s.Domain
	cfg.System.DNSPrimary = s.DNS.Primary
	cfg.System.DNSSecondary = s.DNS.Secondary
	cfg.System.Timezone = s.Timezone
	for _, server := range []string{s.NTP.Primary, s.NTP.Secondary} {
		if server != "" {
			cfg.System.NTPServers = append(cfg.System.NTPServers, server)
		}
	}
}
func ipNames(ip ipBlk) []string {
	var out []string
	for _, e := range ip.Entries {
		out = append(out, e.Name)
	}
	return out
}
func setAddresses(iface *model.Interface, ip ipBlk, dhcp string) {
	iface.Addresses = ipNames(ip)
	if len(iface.Addresses) > 0 {
		iface.IPAddress = iface.Addresses[0]
		iface.AddressMode = "static"
	}
	if dhcp == "yes" {
		iface.AddressMode = "dhcp"
	}
}
func convertNetworkInterfaces(cfg *model.Config, n network) {
	for _, group := range []struct {
		kind    string
		entries []ifaceEntry
	}{{"physical", n.Interface.Ethernet.Entries}, {"lag", n.Interface.Aggregate.Entries}} {
		for _, e := range group.entries {
			iface := model.Interface{Name: e.Name, InterfaceType: group.kind, Description: e.Comment, Status: e.LinkState, LAGParent: e.AggregateGroup}
			if iface.Status == "" {
				iface.Status = "auto"
			}
			var layer *layer3
			switch {
			case e.Layer3 != nil:
				iface.Mode = "layer3"
				layer = e.Layer3
			case e.Layer2 != nil:
				iface.Mode = "layer2"
				layer = e.Layer2
			case e.VirtualWire != nil:
				iface.Mode = "virtual-wire"
				layer = e.VirtualWire
			case e.Tap != nil:
				iface.Mode = "tap"
			case e.AggregateGroup != "":
				iface.Mode = "aggregate-member"
			}
			if layer != nil {
				setAddresses(&iface, layer.IP, layer.DHCP.Enable)
				addIPv6(&iface, layer.IPv6)
				iface.MTU = layer.MTU
			}
			cfg.Interfaces = append(cfg.Interfaces, iface)
			if layer != nil {
				for _, u := range layer.Units {
					addUnit(cfg, u, "subinterface", e.Name, iface.Mode)
				}
			}
			if group.kind == "lag" {
				var members []string
				for _, p := range n.Interface.Ethernet.Entries {
					if p.AggregateGroup == e.Name {
						members = append(members, p.Name)
					}
				}
				cfg.LAGs = append(cfg.LAGs, model.LAG{Name: e.Name, Members: members, LACPEnabled: e.LACP.Enable == "yes", Description: e.Comment})
			}
		}
	}
	for _, group := range []struct {
		kind  string
		block unitBlk
	}{{"vlan", n.Interface.VLAN}, {"loopback", n.Interface.Loopback}, {"tunnel", n.Interface.Tunnel}} {
		for _, u := range group.block.Units {
			addUnit(cfg, u, group.kind, "", "layer3")
		}
	}
	for _, v := range n.VLAN.Entries {
		cfg.VLANs = append(cfg.VLANs, model.VLAN{Name: v.Name, L3Interface: v.VirtualInterface, Members: v.Members.Members})
	}
}
func addUnit(cfg *model.Config, u unitEntry, kind, parent, mode string) {
	iface := model.Interface{Name: u.Name, InterfaceType: kind, ParentInterface: parent, Mode: mode, VLANID: u.Tag, Description: u.Comment, Status: "auto"}
	setAddresses(&iface, u.IP, u.DHCP.Enable)
	addIPv6(&iface, u.IPv6)
	iface.MTU = u.MTU
	cfg.Interfaces = append(cfg.Interfaces, iface)
}
func zoneMembers(z zoneNetwork) []string {
	var out []string
	out = append(out, z.Layer3.Members...)
	out = append(out, z.Layer2.Members...)
	out = append(out, z.VirtualWire.Members...)
	return append(out, z.Tap.Members...)
}
func applyNetworkContexts(cfg *model.Config) {
	byName := map[string]int{}
	for i, iface := range cfg.Interfaces {
		byName[iface.Name] = i
	}
	// Imported logical units take precedence; unassigned members can inherit their parent's vsys.
	for pass := 0; pass < 2; pass++ {
		for i := range cfg.Interfaces {
			iface := &cfg.Interfaces[i]
			parent := iface.LAGParent
			if parent == "" {
				parent = iface.ParentInterface
			}
			if iface.VDOM == "" {
				if j, ok := byName[parent]; ok {
					iface.VDOM = cfg.Interfaces[j].VDOM
				}
			}
			if iface.Context == "" {
				iface.Context = iface.VDOM
			}
		}
	}

	for i := range cfg.VLANs {
		for _, iface := range cfg.Interfaces {
			if iface.Name == cfg.VLANs[i].L3Interface {
				cfg.VLANs[i].Context = iface.VDOM
			}
		}
	}
	for i := range cfg.LAGs {
		for _, iface := range cfg.Interfaces {
			if iface.Name == cfg.LAGs[i].Name {
				cfg.LAGs[i].Context = iface.VDOM
			}
		}
	}
}
func convertStaticRoutes(cfg *model.Config, vrs vrBlk) {
	for _, vr := range vrs.Entries {
		cfg.VirtualContexts = append(cfg.VirtualContexts, model.VirtualContext{ID: vr.Name, Type: model.ContextVirtualRouter})
		for i := range cfg.Interfaces {
			for _, name := range vr.Interfaces.Members {
				if cfg.Interfaces[i].Name == name {
					cfg.Interfaces[i].RoutingContext = vr.Name
				}
			}
		}
		for _, group := range []struct {
			kind  string
			items []routeEntry
		}{{"static", vr.RoutingTable.IP.Static.Entries}, {"static6", vr.RoutingTable.IPv6.Static.Entries}} {
			for _, r := range group.items {
				gw := r.Nexthop.IP
				if gw == "" {
					gw = r.Nexthop.IPv6
				}
				if gw == "" && r.Nexthop.NextVR != "" {
					gw = "next-vr:" + r.Nexthop.NextVR
				}
				if r.Nexthop.Discard != nil {
					gw = "discard"
				}
				action := "forward"
				if r.Nexthop.Discard != nil {
					action = "discard"
				} else if r.Nexthop.NextVR != "" {
					action = "next-vr"
				}
				cfg.Routes = append(cfg.Routes, model.Route{Action: action, Name: r.Name, Destination: r.Destination, Gateway: gw, Interface: r.Interface, Distance: r.AdminDistance, Metric: r.Metric, Context: vr.Name, VDOM: vr.Name, RoutingContext: vr.Name, RouteType: group.kind, Enabled: r.NoInstall != "yes"})
			}
		}
	}
}
func securityProfileNames(s profileSetting) []string {
	var out []string
	groups := []struct {
		prefix  string
		members []string
	}{{"Group:", s.Group.Members}, {"AV:", s.Profiles.Virus.Members}, {"Spyware:", s.Profiles.Spyware.Members}, {"IPS:", s.Profiles.Vulnerability.Members}, {"Web:", s.Profiles.URL.Members}, {"File:", s.Profiles.File.Members}, {"WildFire:", s.Profiles.WildFire.Members}, {"Data:", s.Profiles.Data.Members}}
	for _, g := range groups {
		for _, name := range g.members {
			out = append(out, g.prefix+name)
		}
	}
	return out
}

func addIPv6(iface *model.Interface, ipv6 ipv6Config) {
	if ipv6.Enabled != "" {
		enabled := ipv6.Enabled == "yes"
		iface.IPv6Enabled = &enabled
	}
	for _, entry := range ipv6.Address.Entries {
		found := false
		for _, existing := range iface.Addresses {
			if entry.Name == existing {
				found = true
			}
		}
		if !found {
			iface.Addresses = append(iface.Addresses, entry.Name)
		}
	}
	if iface.IPAddress == "" && len(iface.Addresses) > 0 {
		iface.IPAddress = iface.Addresses[0]
	}
	if iface.AddressMode == "" && len(iface.Addresses) > 0 {
		iface.AddressMode = "static"
	}
}
