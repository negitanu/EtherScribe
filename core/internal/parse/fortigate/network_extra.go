package fortigate

import (
	"strings"

	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/util"
)

// convertDHCP ports WallScribe converters/network.convert_dhcp.
func convertDHCP(cfg *model.Config, scope Node, vdom string) {
	servers := scope.Map("system dhcp server")
	if servers == nil {
		return
	}
	for id, raw := range servers {
		if id == "_name" {
			continue
		}
		dn, ok := raw.(Node)
		if !ok {
			continue
		}
		dhcp := model.DHCPServer{
			Interface: dn.String("interface"),
			Netmask:   dn.String("netmask"),
			Gateway:   dn.String("default-gateway"),
			LeaseTime: dn.String("lease-time"),
			VDOM:      vdom,
			Context:   vdom,
		}
		if ranges := dn.Map("ip-range"); ranges != nil {
			for rid, rr := range ranges {
				if rid == "_name" {
					continue
				}
				rn, ok := rr.(Node)
				if !ok {
					continue
				}
				dhcp.StartIP = rn.String("start-ip")
				dhcp.EndIP = rn.String("end-ip")
				break
			}
		}
		for _, key := range []string{"dns-server1", "dns-server2", "dns-server3"} {
			if dns := dn.String(key); dns != "" {
				dhcp.DNSServers = append(dhcp.DNSServers, dns)
			}
		}
		if excludes := dn.Map("exclude-range"); excludes != nil {
			for eid, er := range excludes {
				if eid == "_name" {
					continue
				}
				en, ok := er.(Node)
				if !ok {
					continue
				}
				start, end := en.String("start-ip"), en.String("end-ip")
				switch {
				case start != "" && end != "":
					dhcp.ExcludeIPs = append(dhcp.ExcludeIPs, start+"-"+end)
				case start != "":
					dhcp.ExcludeIPs = append(dhcp.ExcludeIPs, start)
				}
			}
		}
		cfg.DHCPServers = append(cfg.DHCPServers, dhcp)
	}
}

// convertOSPF ports WallScribe converters/network.convert_ospf (ospf + ospf6).
func convertOSPF(cfg *model.Config, scope Node, vdom string) {
	if ospf := scope.Map("router ospf"); ospf != nil && len(ospf) > 0 {
		cfg.Routing.OSPF = append(cfg.Routing.OSPF, parseOSPF(ospf, vdom))
	}
	if ospf6 := scope.Map("router ospf6"); ospf6 != nil && len(ospf6) > 0 {
		cfg.Routing.OSPF6 = append(cfg.Routing.OSPF6, parseOSPF(ospf6, vdom))
	}
}

func parseOSPF(ospf Node, vdom string) model.OSPFSettings {
	out := model.OSPFSettings{
		RouterID:                    ospf.String("router-id"),
		DefaultInformationOriginate: ospf.String("default-information-originate") == "enable",
		DefaultMetric:               ospf.String("default-metric"),
		Distance:                    ospf.String("distance"),
		PassiveInterfaces:           ospf.List("passive-interface"),
		VDOM:                        vdom,
		Context:                     vdom,
	}
	if areas := ospf.Map("area"); areas != nil {
		for id, raw := range areas {
			if id == "_name" {
				continue
			}
			an, ok := raw.(Node)
			if !ok {
				continue
			}
			out.Areas = append(out.Areas, model.OSPFArea{
				AreaID:         firstNonEmpty(an.String("_name"), id),
				AreaType:       firstNonEmpty(an.String("stub-type"), "normal"),
				Authentication: an.String("authentication"),
			})
		}
	}
	if ifaces := ospf.Map("ospf-interface"); ifaces != nil {
		for name, raw := range ifaces {
			if name == "_name" {
				continue
			}
			in, ok := raw.(Node)
			if !ok {
				continue
			}
			out.Interfaces = append(out.Interfaces, model.OSPFInterface{
				Name:           firstNonEmpty(in.String("_name"), name),
				Interface:      in.String("interface"),
				Area:           firstNonEmpty(in.String("area"), "0.0.0.0"),
				Cost:           in.String("cost"),
				Priority:       in.String("priority"),
				HelloInterval:  in.String("hello-interval"),
				DeadInterval:   in.String("dead-interval"),
				NetworkType:    in.String("network-type"),
				Authentication: in.String("authentication"),
				Passive:        in.String("passive") == "enable",
			})
		}
	}
	if nets := ospf.Map("network"); nets != nil {
		for id, raw := range nets {
			if id == "_name" {
				continue
			}
			nn, ok := raw.(Node)
			if !ok {
				continue
			}
			prefix := util.IPToCIDR(nn.String("prefix"))
			if list := nn.List("prefix"); len(list) > 0 {
				prefix = util.IPToCIDR(strings.Join(list, " "))
			}
			areaID := firstNonEmpty(nn.String("area"), "0.0.0.0")
			attached := false
			for i := range out.Areas {
				if out.Areas[i].AreaID == areaID {
					out.Areas[i].Networks = append(out.Areas[i].Networks, prefix)
					attached = true
					break
				}
			}
			if !attached {
				out.Areas = append(out.Areas, model.OSPFArea{AreaID: areaID, Networks: []string{prefix}})
			}
		}
	}
	for _, proto := range []string{"connected", "static", "bgp", "rip", "isis"} {
		redist := ospf.Map("redistribute " + proto)
		if redist == nil {
			if parent := ospf.Map("redistribute"); parent != nil {
				redist = parent.Map(proto)
			}
		}
		if redist == nil {
			continue
		}
		status := redist.String("status") == "enable"
		if !status && len(redist) == 0 {
			continue
		}
		out.Redistributes = append(out.Redistributes, model.OSPFRedistribute{
			Protocol:   proto,
			Status:     status,
			Metric:     redist.String("metric"),
			MetricType: redist.String("metric-type"),
			RouteMap:   redist.String("routemap"),
		})
	}
	return out
}

// convertBGP ports WallScribe converters/network.convert_bgp.
func convertBGP(cfg *model.Config, scope Node, vdom string) {
	bgpNode := scope.Map("router bgp")
	if bgpNode == nil || len(bgpNode) == 0 {
		return
	}
	bgp := model.BGPSettings{
		ASNumber: bgpNode.String("as"),
		RouterID: bgpNode.String("router-id"),
		VDOM:     vdom,
		Context:  vdom,
	}
	if neighbors := bgpNode.Map("neighbor"); neighbors != nil {
		for ip, raw := range neighbors {
			if ip == "_name" {
				continue
			}
			nn, ok := raw.(Node)
			if !ok {
				continue
			}
			bgp.Neighbors = append(bgp.Neighbors, model.BGPNeighbor{
				IP:                  firstNonEmpty(nn.String("_name"), ip),
				RemoteAS:            nn.String("remote-as"),
				Description:         nn.String("description"),
				UpdateSource:        nn.String("update-source"),
				EBGPMultihop:        nn.String("ebgp-multihop"),
				NextHopSelf:         nn.String("next-hop-self") == "enable",
				SoftReconfiguration: nn.String("soft-reconfiguration") == "enable",
				RouteMapIn:          nn.String("route-map-in"),
				RouteMapOut:         nn.String("route-map-out"),
				Activate:            nn.String("activate") != "disable",
				Shutdown:            nn.String("shutdown") == "enable",
			})
		}
	}
	if nets := bgpNode.Map("network"); nets != nil {
		for id, raw := range nets {
			if id == "_name" {
				continue
			}
			nn, ok := raw.(Node)
			if !ok {
				continue
			}
			prefix := util.IPToCIDR(nn.String("prefix"))
			if list := nn.List("prefix"); len(list) > 0 {
				prefix = util.IPToCIDR(strings.Join(list, " "))
			}
			bgp.Networks = append(bgp.Networks, model.BGPNetwork{
				Prefix: prefix, RouteMap: nn.String("route-map"),
			})
		}
	}
	for _, proto := range []string{"connected", "static", "ospf", "rip", "isis"} {
		redist := bgpNode.Map("redistribute " + proto)
		if redist == nil {
			continue
		}
		status := redist.String("status") == "enable"
		if !status && len(redist) == 0 {
			continue
		}
		bgp.Redistributes = append(bgp.Redistributes, model.BGPRedistribute{
			Protocol: proto, Status: status, RouteMap: redist.String("route-map"),
		})
	}
	cfg.Routing.BGP = append(cfg.Routing.BGP, bgp)
}

// convertPolicyRoutes ports WallScribe converters/network.convert_policy_routes.
func convertPolicyRoutes(cfg *model.Config, scope Node, vdom string) {
	routes := scope.Map("router policy")
	if routes == nil {
		return
	}
	for id, raw := range routes {
		if id == "_name" {
			continue
		}
		rn, ok := raw.(Node)
		if !ok {
			continue
		}
		src := joinCIDRs(rn.List("src"))
		if src == "" {
			src = util.IPToCIDR(rn.String("src"))
		}
		dst := joinCIDRs(rn.List("dst"))
		if dst == "" {
			dst = util.IPToCIDR(rn.String("dst"))
		}
		cfg.Routing.PolicyRoutes = append(cfg.Routing.PolicyRoutes, model.PolicyRoute{
			SeqNum:       firstNonEmpty(rn.String("_name"), id),
			Src:          src,
			SrcNegate:    rn.String("src-negate") == "enable",
			Dst:          dst,
			DstNegate:    rn.String("dst-negate") == "enable",
			Protocol:     rn.String("protocol"),
			InputDevice:  strings.Join(rn.List("input-device"), ", "),
			OutputDevice: strings.Join(rn.List("output-device"), ", "),
			Gateway:      rn.String("gateway"),
			Action:       firstNonEmpty(rn.String("action"), "permit"),
			Status:       rn.String("status") != "disable",
			Comments:     rn.String("comments"),
			VDOM:         vdom,
			Context:      vdom,
		})
	}
}

func joinCIDRs(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	var out []string
	for _, p := range parts {
		out = append(out, util.IPToCIDR(p))
	}
	return strings.Join(out, " ")
}
