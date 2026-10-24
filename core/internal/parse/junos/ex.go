package junos

import (
	"strings"

	"github.com/negitanu/etherscribe/internal/identify"
	"github.com/negitanu/etherscribe/internal/model"
)

// ParseEX converts Junos EX (ELS) brace or set config into the unified model.
func ParseEX(content, filename string) (*model.Config, error) {
	cfg := &model.Config{
		SourceFile: filename,
		DeviceInfo: model.DeviceInfo{
			DeviceType: identify.JunosEX, Role: model.RoleSwitch, VendorFamily: "junos-ex",
			Model: "EX Series", OperationMode: "Switch",
		},
		VirtualContexts: []model.VirtualContext{{ID: model.DefaultContext, Type: model.ContextDefault}},
		System:          model.SystemSettings{HTTPSPort: "443", SSHPort: "22"},
		Objects:         model.Objects{},
	}
	if id := identify.Identify(content); id != nil {
		if id.Model != "" {
			cfg.DeviceInfo.Model = id.Model
		}
		if id.OSVersion != "" {
			cfg.DeviceInfo.OSVersion = id.OSVersion
			cfg.DeviceInfo.OSVersionSource = "version"
		}
	}

	stmts, nErrs := Normalize(content)
	cfg.ParseErrors = append(cfg.ParseErrors, nErrs...)

	ifaces := map[string]*model.Interface{}
	vlans := map[string]*model.VLAN{}
	lags := map[string]*model.LAG{}
	contexts := map[string]model.VirtualContext{model.DefaultContext: cfg.VirtualContexts[0]}
	seenTop := map[string]struct{}{}

	ensureIface := func(name string) *model.Interface {
		if i, ok := ifaces[name]; ok {
			return i
		}
		i := &model.Interface{Name: name, Status: "up", Context: model.DefaultContext}
		switch {
		case strings.HasPrefix(name, "ae"):
			i.InterfaceType = "lag"
		case strings.HasPrefix(name, "irb.") || name == "irb":
			i.InterfaceType = "irb"
		case strings.HasPrefix(name, "ge-") || strings.HasPrefix(name, "xe-") || strings.HasPrefix(name, "et-"):
			i.InterfaceType = "physical"
		default:
			i.InterfaceType = "l3"
		}
		ifaces[name] = i
		return i
	}

	for _, st := range stmts {
		if st.Op == OpDelete {
			cfg.Coverage = append(cfg.Coverage, model.CoverageItem{
				Path: strings.Join(st.Path, " "), Reason: "delete 文はモデルへ未適用",
			})
			continue
		}
		if len(st.Path) == 0 {
			continue
		}
		seenTop[st.Path[0]] = struct{}{}

		switch {
		case PathHas(st, "system", "host-name") && len(st.Path) >= 3:
			cfg.DeviceInfo.Hostname = st.Path[2]
		case PathHas(st, "system", "time-zone") && len(st.Path) >= 3:
			cfg.System.Timezone = st.Path[2]
		case PathHas(st, "system", "name-server") && len(st.Path) >= 3:
			if cfg.System.DNSPrimary == "" {
				cfg.System.DNSPrimary = st.Path[2]
			} else if cfg.System.DNSSecondary == "" {
				cfg.System.DNSSecondary = st.Path[2]
			}
		case PathHas(st, "system", "ntp", "server") && len(st.Path) >= 4:
			cfg.System.NTPServers = append(cfg.System.NTPServers, st.Path[3])
		case PathHas(st, "version") && len(st.Path) >= 2:
			cfg.DeviceInfo.OSVersion = st.Path[1]
			cfg.DeviceInfo.OSVersionSource = "version"
		case PathHas(st, "interfaces"):
			mapInterface(st, ensureIface, lags)
		case PathHas(st, "vlans"):
			mapVLAN(st, vlans)
		case PathHas(st, "protocols", "rstp") || PathHas(st, "protocols", "mstp") || PathHas(st, "protocols", "vstp"):
			mapSTP(st, cfg)
		case PathHas(st, "routing-instances"):
			mapRoutingInstance(st, cfg, contexts, ensureIface)
		case PathHas(st, "routing-options", "static", "route"):
			mapStaticRoute(st, cfg, model.DefaultContext)
		case PathHas(st, "protocols", "ospf"):
			mapOSPF(st, cfg, model.DefaultContext)
		case PathHas(st, "protocols", "bgp"):
			mapBGP(st, cfg, model.DefaultContext)
		}
	}

	for _, i := range ifaces {
		cfg.Interfaces = append(cfg.Interfaces, *i)
	}
	for _, v := range vlans {
		cfg.VLANs = append(cfg.VLANs, *v)
	}
	for _, l := range lags {
		cfg.LAGs = append(cfg.LAGs, *l)
	}
	cfg.VirtualContexts = cfg.VirtualContexts[:0]
	for _, c := range contexts {
		cfg.VirtualContexts = append(cfg.VirtualContexts, c)
	}

	knownTop := map[string]struct{}{
		"system": {}, "interfaces": {}, "vlans": {}, "protocols": {},
		"routing-options": {}, "routing-instances": {}, "version": {},
		"snmp": {}, "chassis": {}, "groups": {}, "apply-groups": {},
	}
	for top := range seenTop {
		if _, ok := knownTop[top]; !ok {
			cfg.Coverage = append(cfg.Coverage, model.CoverageItem{
				Path: top, Reason: "未対応または部分対応のトップレベル",
			})
		}
	}
	return cfg, nil
}

func mapInterface(st Statement, ensure func(string) *model.Interface, lags map[string]*model.LAG) {
	// interfaces <if> …
	if len(st.Path) < 2 {
		return
	}
	name := st.Path[1]
	// brace form: interfaces irb unit 10 family inet address … → irb.10
	if name == "irb" && len(st.Path) >= 4 && st.Path[2] == "unit" {
		name = "irb." + st.Path[3]
	}
	iface := ensure(name)
	if st.Op == OpDeactivate {
		iface.Status = "down"
	}
	base := st.Path[1]
	// interfaces <if> description …
	if len(st.Path) >= 4 && st.Path[2] == "description" && base != "irb" {
		iface.Description = strings.Join(st.Path[3:], " ")
		return
	}
	// unit index depends on whether we rewrote irb
	unitIdx := 2
	if base == "irb" && strings.HasPrefix(name, "irb.") {
		unitIdx = 2 // path still interfaces irb unit N …
	}
	// interfaces <if> unit <u> family ethernet-switching …
	if len(st.Path) >= unitIdx+5 && st.Path[unitIdx] == "unit" && st.Path[unitIdx+2] == "family" && st.Path[unitIdx+3] == "ethernet-switching" {
		leaf := st.Path[unitIdx+4]
		switch leaf {
		case "interface-mode":
			if len(st.Path) > unitIdx+5 {
				iface.SwitchportMode = st.Path[unitIdx+5]
			}
		case "vlan":
			if len(st.Path) > unitIdx+5 && st.Path[unitIdx+5] == "members" {
				members := cleanMembers(st.Path[unitIdx+6:])
				if iface.SwitchportMode == "trunk" || len(members) > 1 {
					iface.SwitchportMode = "trunk"
					iface.TrunkVLANs = append(iface.TrunkVLANs, members...)
				} else if len(members) == 1 {
					if iface.SwitchportMode == "" {
						iface.SwitchportMode = "access"
					}
					iface.AccessVLAN = members[0]
				}
			}
		}
		return
	}
	// interfaces <if> unit <u> family inet address x/y  (also irb)
	if len(st.Path) >= unitIdx+6 && st.Path[unitIdx] == "unit" && st.Path[unitIdx+2] == "family" && st.Path[unitIdx+3] == "inet" && st.Path[unitIdx+4] == "address" {
		iface.IPAddress = st.Path[unitIdx+5]
		if iface.InterfaceType == "physical" {
			iface.InterfaceType = "l3"
		}
		return
	}
	// interfaces <if> ether-options 802.3ad aeX
	if len(st.Path) >= 5 && st.Path[2] == "ether-options" && st.Path[3] == "802.3ad" {
		ae := st.Path[4]
		iface.LAGParent = ae
		lag := lags[ae]
		if lag == nil {
			lag = &model.LAG{Name: ae, LACPEnabled: true, Context: model.DefaultContext}
			lags[ae] = lag
		}
		lag.Members = appendUnique(lag.Members, name)
		ensure(ae)
		return
	}
	if len(st.Path) >= 5 && st.Path[2] == "gigether-options" && st.Path[3] == "802.3ad" {
		ae := st.Path[4]
		iface.LAGParent = ae
		lag := lags[ae]
		if lag == nil {
			lag = &model.LAG{Name: ae, LACPEnabled: true, Context: model.DefaultContext}
			lags[ae] = lag
		}
		lag.Members = appendUnique(lag.Members, name)
		ensure(ae)
	}
}

func mapVLAN(st Statement, vlans map[string]*model.VLAN) {
	// vlans <name> vlan-id N
	if len(st.Path) < 2 {
		return
	}
	name := st.Path[1]
	v := vlans[name]
	if v == nil {
		v = &model.VLAN{Name: name, Context: model.DefaultContext}
		vlans[name] = v
	}
	if len(st.Path) >= 4 && st.Path[2] == "vlan-id" {
		v.VLANID = st.Path[3]
	}
	if len(st.Path) >= 4 && st.Path[2] == "l3-interface" {
		v.L3Interface = st.Path[3]
	}
	if len(st.Path) >= 4 && st.Path[2] == "description" {
		v.Description = strings.Join(st.Path[3:], " ")
	}
}

func mapSTP(st Statement, cfg *model.Config) {
	if cfg.STP == nil {
		cfg.STP = &model.STPSettings{Enabled: true, Context: model.DefaultContext}
	}
	if PathHas(st, "protocols", "rstp") {
		cfg.STP.Protocol = "rstp"
	} else if PathHas(st, "protocols", "mstp") {
		cfg.STP.Protocol = "mstp"
	} else if PathHas(st, "protocols", "vstp") {
		cfg.STP.Protocol = "vstp"
	}
	if st.Op == OpDeactivate {
		cfg.STP.Enabled = false
	}
}

func mapRoutingInstance(st Statement, cfg *model.Config, contexts map[string]model.VirtualContext, ensure func(string) *model.Interface) {
	// routing-instances <name> …
	if len(st.Path) < 2 {
		return
	}
	name := st.Path[1]
	if _, ok := contexts[name]; !ok {
		contexts[name] = model.VirtualContext{ID: name, Type: model.ContextRoutingInstance}
	}
	if len(st.Path) >= 4 && st.Path[2] == "instance-type" {
		c := contexts[name]
		c.Description = st.Path[3]
		contexts[name] = c
	}
	if len(st.Path) >= 4 && st.Path[2] == "interface" {
		iface := ensure(st.Path[3])
		iface.Context = name
	}
	// routing-instances <name> routing-options static route …
	if len(st.Path) >= 6 && st.Path[2] == "routing-options" && st.Path[3] == "static" && st.Path[4] == "route" {
		sub := Statement{Op: st.Op, Path: append([]string{"routing-options", "static", "route"}, st.Path[5:]...), Raw: st.Raw}
		mapStaticRoute(sub, cfg, name)
	}
	if len(st.Path) >= 4 && st.Path[2] == "protocols" && st.Path[3] == "ospf" {
		sub := Statement{Op: st.Op, Path: st.Path[2:], Raw: st.Raw}
		mapOSPF(sub, cfg, name)
	}
	if len(st.Path) >= 4 && st.Path[2] == "protocols" && st.Path[3] == "bgp" {
		sub := Statement{Op: st.Op, Path: st.Path[2:], Raw: st.Raw}
		mapBGP(sub, cfg, name)
	}
}

func mapStaticRoute(st Statement, cfg *model.Config, context string) {
	// routing-options static route <prefix> next-hop <nh>
	// path after normalize may be routing-options static route PREFIX next-hop NH
	idx := indexOf(st.Path, "route")
	if idx < 0 || idx+1 >= len(st.Path) {
		return
	}
	prefix := st.Path[idx+1]
	gw := ""
	iface := ""
	for i := idx + 2; i < len(st.Path)-1; i++ {
		switch st.Path[i] {
		case "next-hop":
			gw = st.Path[i+1]
		case "qualified-next-hop":
			gw = st.Path[i+1]
		}
	}
	cfg.Routes = append(cfg.Routes, model.Route{
		Name: prefix, Destination: prefix, Gateway: gw, Interface: iface,
		Context: context, VDOM: context, RouteType: "static", Enabled: st.Op != OpDeactivate,
	})
}

func mapOSPF(st Statement, cfg *model.Config, context string) {
	var ospf *model.OSPFSettings
	for i := range cfg.Routing.OSPF {
		if cfg.Routing.OSPF[i].Context == context {
			ospf = &cfg.Routing.OSPF[i]
			break
		}
	}
	if ospf == nil {
		cfg.Routing.OSPF = append(cfg.Routing.OSPF, model.OSPFSettings{Context: context, VDOM: context})
		ospf = &cfg.Routing.OSPF[len(cfg.Routing.OSPF)-1]
	}
	if PathHas(st, "protocols", "ospf", "router-id") && len(st.Path) >= 4 {
		ospf.RouterID = st.Path[3]
	}
	// area <id> interface <if>
	if len(st.Path) >= 6 && st.Path[2] == "area" && st.Path[4] == "interface" {
		areaID := st.Path[3]
		found := false
		for i := range ospf.Areas {
			if ospf.Areas[i].AreaID == areaID {
				found = true
				break
			}
		}
		if !found {
			ospf.Areas = append(ospf.Areas, model.OSPFArea{AreaID: areaID})
		}
		ospf.Interfaces = append(ospf.Interfaces, model.OSPFInterface{
			Name: st.Path[5], Interface: st.Path[5], Area: areaID,
		})
	}
}

func mapBGP(st Statement, cfg *model.Config, context string) {
	var bgp *model.BGPSettings
	for i := range cfg.Routing.BGP {
		if cfg.Routing.BGP[i].Context == context {
			bgp = &cfg.Routing.BGP[i]
			break
		}
	}
	if bgp == nil {
		cfg.Routing.BGP = append(cfg.Routing.BGP, model.BGPSettings{Context: context, VDOM: context})
		bgp = &cfg.Routing.BGP[len(cfg.Routing.BGP)-1]
	}
	// protocols bgp group <g> peer-as N / neighbor IP
	if len(st.Path) >= 5 && st.Path[2] == "group" {
		if len(st.Path) >= 6 && st.Path[4] == "peer-as" {
			// store as on group via neighbor later; keep AS on settings if empty
			if bgp.ASNumber == "" {
				bgp.ASNumber = st.Path[5]
			}
		}
		if len(st.Path) >= 6 && st.Path[4] == "neighbor" {
			ip := st.Path[5]
			remote := ""
			if len(st.Path) >= 8 && st.Path[6] == "peer-as" {
				remote = st.Path[7]
			}
			bgp.Neighbors = append(bgp.Neighbors, model.BGPNeighbor{
				IP: ip, RemoteAS: remote, Activate: true,
			})
		}
	}
	if PathHas(st, "protocols", "bgp", "local-as") && len(st.Path) >= 4 {
		bgp.ASNumber = st.Path[3]
	}
}

func cleanMembers(tokens []string) []string {
	var out []string
	for _, t := range tokens {
		t = strings.Trim(t, "[]")
		if t == "" || t == "[" || t == "]" {
			continue
		}
		out = append(out, t)
	}
	return out
}

func appendUnique(list []string, v string) []string {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}

func indexOf(path []string, want string) int {
	for i, p := range path {
		if p == want {
			return i
		}
	}
	return -1
}
