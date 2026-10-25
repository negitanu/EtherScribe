package fortigate

import (
	"github.com/negitanu/etherscribe/internal/model"
	"github.com/negitanu/etherscribe/internal/util"
)

func enrichInterface(cfg *model.Config, iface *model.Interface, in Node) {
	iface.Context = iface.VDOM
	iface.RoutingContext = in.String("vrf")
	iface.AddressMode = firstNonEmpty(in.String("mode"), "static")
	if in.String("mtu-override") == "enable" {
		iface.MTU = in.String("mtu")
	}
	add := func(address string) {
		if address == "" {
			return
		}
		for _, a := range iface.Addresses {
			if a == address {
				return
			}
		}
		iface.Addresses = append(iface.Addresses, address)
	}
	add(iface.IPAddress)
	if in.String("secondary-IP") == "enable" {
		secondary := in.Map("secondaryip")
		for _, id := range EditNames(secondary) {
			if entry := secondary.Map(id); entry != nil {
				add(util.IPToCIDR(entry.String("ip")))
			}
		}
	}
	if ipv6 := in.Map("ipv6"); ipv6 != nil {
		add(ipv6.String("ip6-address"))
		for _, section := range []string{"ip6-extra-addr"} {
			entries := ipv6.Map(section)
			for _, id := range EditNames(entries) {
				if e := entries.Map(id); e != nil {
					add(firstNonEmpty(e.String("prefix"), e.String("_name")))
				}
			}
		}
	}
	if iface.IPAddress == "" && len(iface.Addresses) > 0 {
		iface.IPAddress = iface.Addresses[0]
	}
	if in.String("type") == "aggregate" {
		iface.InterfaceType = "lag"
		mode := firstNonEmpty(in.String("lacp-mode"), "active")
		cfg.LAGs = append(cfg.LAGs, model.LAG{Name: iface.Name, Members: in.List("member"), LACPEnabled: mode != "static", Context: iface.VDOM, Description: iface.Description})
	}
}
