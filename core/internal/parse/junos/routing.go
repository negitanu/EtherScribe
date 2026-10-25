package junos

import (
	"net"
	"sort"
	"strings"

	"github.com/negitanu/etherscribe/internal/model"
)

func prefixPath(path, prefix []string) bool {
	if len(path) < len(prefix) {
		return false
	}
	for i, v := range prefix {
		if path[i] != v {
			return false
		}
	}
	return true
}
func disabledPaths(stmts []Statement) [][]string {
	var out [][]string
	for _, st := range stmts {
		if st.Op == OpDeactivate {
			out = append(out, st.Path)
		}
	}
	return out
}
func inactive(path []string, disabled [][]string) bool {
	for _, p := range disabled {
		if prefixPath(path, p) {
			return true
		}
	}
	return false
}

// A deactivated leaf must not disable an otherwise active interface or BGP peer.
func effectiveStatements(stmts []Statement) []Statement {
	disabled := disabledPaths(stmts)
	var out []Statement
	for _, st := range stmts {
		if st.Op == OpDeactivate {
			if len(st.Path) == 2 && st.Path[0] == "interfaces" {
				out = append(out, st)
			}
			continue
		}
		if st.Op == OpSet && inactive(st.Path, disabled) {
			continue
		}
		out = append(out, st)
	}
	return out
}
func contextPath(st Statement) (string, []string) {
	if len(st.Path) >= 3 && st.Path[0] == "routing-instances" {
		return st.Path[1], st.Path[2:]
	}
	return model.DefaultContext, st.Path
}
func collectRouting(stmts []Statement, cfg *model.Config) {
	disabled := disabledPaths(stmts)
	collectStatic(stmts, disabled, cfg)
	collectBGP(stmts, disabled, cfg)
}

type staticSpec struct {
	context, prefix string
	properties      map[string]string
	hops            map[string]map[string]string
	order           []string
	path            []string
}

func collectStatic(stmts []Statement, disabled [][]string, cfg *model.Config) {
	specs := map[string]*staticSpec{}
	var order []string
	for _, st := range stmts {
		if st.Op != OpSet {
			continue
		}
		context, p := contextPath(st)
		routePathLength := len(st.Path) - len(p) + 4
		if len(p) >= 6 && prefixPath(p, []string{"routing-options", "rib", "inet6.0", "static", "route"}) {
			routePathLength += 2
			p = append([]string{"routing-options"}, p[3:]...)
		}
		if len(p) < 4 || !prefixPath(p, []string{"routing-options", "static", "route"}) {
			continue
		}
		key := context + "\x00" + p[3]
		s := specs[key]
		if s == nil {
			s = &staticSpec{context: context, prefix: p[3], properties: map[string]string{}, hops: map[string]map[string]string{}, path: st.Path[:routePathLength]}
			specs[key] = s
			order = append(order, key)
		}
		tail := p[4:]
		if len(tail) == 0 {
			continue
		}
		// Inactive attributes are excluded, while whole inactive routes stay visible as disabled.
		if inactive(st.Path, disabled) && !inactive(s.path, disabled) {
			continue
		}
		switch tail[0] {
		case "next-hop":
			for _, hop := range cleanMembers(tail[1:]) {
				if _, ok := s.hops[hop]; !ok {
					s.hops[hop] = map[string]string{}
					s.order = append(s.order, hop)
				}
			}
		case "qualified-next-hop":
			if len(tail) < 2 {
				continue
			}
			hop := tail[1]
			if _, ok := s.hops[hop]; !ok {
				s.hops[hop] = map[string]string{}
				s.order = append(s.order, hop)
			}
			if len(tail) >= 4 {
				s.hops[hop][tail[2]] = tail[3]
			}
		case "preference", "metric", "description":
			if len(tail) > 1 {
				s.properties[tail[0]] = strings.Join(tail[1:], " ")
			}
		case "discard", "reject":
			s.properties["action"] = tail[0]
		}
	}
	for _, key := range order {
		s := specs[key]
		hops := s.order
		if len(hops) == 0 {
			hops = []string{""}
		}
		for _, hop := range hops {
			value := func(k string) string {
				if v, ok := s.hops[hop][k]; ok {
					return v
				}
				return s.properties[k]
			}
			r := model.Route{Name: s.prefix, Destination: s.prefix, Context: s.context, VDOM: s.context, RoutingContext: s.context, RouteType: "static", Enabled: !inactive(s.path, disabled), Distance: value("preference"), Metric: value("metric"), Description: value("description"), Action: s.properties["action"]}
			if r.Action == "" {
				r.Action = "forward"
			}
			if net.ParseIP(hop) != nil {
				r.Gateway = hop
			} else {
				r.Interface = hop
			}
			if strings.Contains(s.prefix, ":") {
				r.RouteType = "static6"
			}
			cfg.Routes = append(cfg.Routes, r)
		}
	}
}

type bgpLevel map[string][]string
type bgpGroup struct {
	attrs     bgpLevel
	neighbors map[string]bgpLevel
	path      []string
}
type bgpScope struct {
	attrs         bgpLevel
	groups        map[string]*bgpGroup
	asn, routerID string
	exists        bool
}

func collectBGP(stmts []Statement, disabled [][]string, cfg *model.Config) {
	scopes := map[string]*bgpScope{}
	for _, st := range stmts {
		if st.Op != OpSet {
			continue
		}
		context, p := contextPath(st)
		if len(p) < 2 {
			continue
		}
		isBGP := prefixPath(p, []string{"protocols", "bgp"})
		isRouting := p[0] == "routing-options" && (p[1] == "autonomous-system" || p[1] == "router-id")
		if !isBGP && !isRouting {
			continue
		}
		scope := scopes[context]
		if scope == nil {
			scope = &bgpScope{attrs: bgpLevel{}, groups: map[string]*bgpGroup{}}
			scopes[context] = scope
		}
		if isRouting {
			if len(p) > 2 && !inactive(st.Path, disabled) {
				if p[1] == "autonomous-system" {
					scope.asn = p[2]
				} else {
					scope.routerID = p[2]
				}
			}
			continue
		}
		scope.exists = true
		tail := p[2:]
		attrs := scope.attrs
		if len(tail) >= 2 && tail[0] == "group" {
			group := scope.groups[tail[1]]
			if group == nil {
				group = &bgpGroup{attrs: bgpLevel{}, neighbors: map[string]bgpLevel{}, path: append(append([]string{}, st.Path[:len(st.Path)-len(p)]...), p[:4]...)}
				scope.groups[tail[1]] = group
			}
			attrs = group.attrs
			tail = tail[2:]
			if len(tail) >= 2 && tail[0] == "neighbor" {
				n := group.neighbors[tail[1]]
				if n == nil {
					n = bgpLevel{}
					group.neighbors[tail[1]] = n
				}
				attrs = n
				tail = tail[2:]
			}
		}
		// A disabled group/peer remains reportable, but an inactive individual attribute is not inherited.
		if len(tail) > 0 {
			scopeLen := len(st.Path) - len(tail)
			if inactive(st.Path, disabled) && !inactive(st.Path[:scopeLen], disabled) {
				continue
			}
			attrs[tail[0]] = cleanMembers(tail[1:])
		}
	}
	contexts := make([]string, 0, len(scopes))
	for c := range scopes {
		contexts = append(contexts, c)
	}
	sort.Strings(contexts)
	for _, c := range contexts {
		scope := scopes[c]
		if !scope.exists {
			continue
		}
		// Routing instances inherit the global AS/router-id unless explicitly configured.
		asn, rid := scope.asn, scope.routerID
		if global := scopes[model.DefaultContext]; global != nil {
			if asn == "" {
				asn = global.asn
			}
			if rid == "" {
				rid = global.routerID
			}
		}
		if asn == "" {
			asn = firstValue(scope.attrs["local-as"])
		}
		bgp := model.BGPSettings{ASNumber: asn, RouterID: rid, Context: c, VDOM: c}
		groups := make([]string, 0, len(scope.groups))
		for g := range scope.groups {
			groups = append(groups, g)
		}
		sort.Strings(groups)
		for _, name := range groups {
			group := scope.groups[name]
			ips := make([]string, 0, len(group.neighbors))
			for ip := range group.neighbors {
				ips = append(ips, ip)
			}
			sort.Strings(ips)
			for _, ip := range ips {
				n := group.neighbors[ip]
				value := func(key string) []string {
					for _, a := range []bgpLevel{n, group.attrs, scope.attrs} {
						if v, ok := a[key]; ok {
							return v
						}
					}
					return nil
				}
				path := append(append([]string{}, group.path...), "neighbor", ip)
				off := inactive(path, disabled)
				_, shutdown := n["shutdown"]
				_, groupShutdown := group.attrs["shutdown"]
				_, globalShutdown := scope.attrs["shutdown"]
				local := firstValue(value("local-as"))
				if local == "" {
					local = asn
				}
				remote := firstValue(value("peer-as"))
				if remote == "" && firstValue(value("type")) == "internal" {
					remote = asn
				}
				bgp.Neighbors = append(bgp.Neighbors, model.BGPNeighbor{IP: ip, Group: name, RemoteAS: remote, LocalAS: local, Description: strings.Join(value("description"), " "), UpdateSource: firstValue(value("local-address")), ImportPolicies: value("import"), ExportPolicies: value("export"), Activate: !off && !shutdown && !groupShutdown && !globalShutdown, Shutdown: off || shutdown || groupShutdown || globalShutdown})
			}
		}
		cfg.Routing.BGP = append(cfg.Routing.BGP, bgp)
	}
}
func firstValue(values []string) string {
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func finalizeInterfaces(stmts []Statement, ifaces map[string]*model.Interface, lags map[string]*model.LAG) {
	members := map[string][]string{}
	modes := map[string]string{}
	for _, st := range effectiveStatements(stmts) {
		if st.Op != OpSet || len(st.Path) < 3 || st.Path[0] != "interfaces" {
			continue
		}
		p := st.Path
		if len(p) >= 8 && p[2] == "unit" && p[4] == "family" && p[5] == "ethernet-switching" {
			if p[6] == "interface-mode" {
				modes[p[1]] = p[7]
			}
			if len(p) >= 9 && p[6] == "vlan" && p[7] == "members" {
				for _, v := range cleanMembers(p[8:]) {
					members[p[1]] = appendUnique(members[p[1]], v)
				}
			}
		}
		i := ifaces[p[1]]
		if i != nil {
			if p[2] == "disable" {
				i.Status = "down"
			}
			if p[2] == "mtu" && len(p) > 3 {
				i.MTU = p[3]
			}
			if len(p) > 3 && p[2] == "aggregated-ether-options" && p[3] == "lacp" {
				lag := lags[p[1]]
				if lag == nil {
					lag = &model.LAG{Name: p[1], Context: model.DefaultContext}
					lags[p[1]] = lag
				}
				lag.LACPEnabled = true
			}
		}
	}
	// Resolve access/trunk membership after all interface-mode statements have been read.
	for _, i := range ifaces {
		if values := members[i.Name]; len(values) > 0 {
			mode := modes[i.Name]
			if mode == "" {
				mode = "access"
				if len(values) > 1 {
					mode = "trunk"
				}
			}
			i.SwitchportMode = mode
			if mode == "trunk" {
				i.TrunkVLANs = values
				i.AccessVLAN = ""
			} else {
				i.AccessVLAN = values[0]
				i.TrunkVLANs = nil
			}
		}
		if i.SwitchportMode == "trunk" && i.AccessVLAN != "" {
			i.TrunkVLANs = appendUnique(i.TrunkVLANs, i.AccessVLAN)
			i.AccessVLAN = ""
		}
		if i.Context != model.DefaultContext {
			i.RoutingContext = i.Context
		}
	}
}
