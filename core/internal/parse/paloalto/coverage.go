package paloalto

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"

	"github.com/negitanu/etherscribe/internal/model"
)

type coverageNode struct {
	XMLName  xml.Name
	Children []coverageNode `xml:",any"`
}

// Only structural element paths are retained. Entry names, attributes and text
// (including credentials and device identifiers) never enter coverage notes.
func collectPACoverage(content string, cfg *model.Config) {
	var root coverageNode
	if xml.Unmarshal([]byte(content), &root) != nil {
		return
	}
	rules := map[string]string{
		"":                                                       "devices shared",
		"devices":                                                "entry",
		"devices/entry":                                          "network vsys deviceconfig",
		"devices/entry/deviceconfig":                             "system",
		"devices/entry/deviceconfig/system":                      "hostname ip-address netmask default-gateway domain timezone dns-setting ntp-servers",
		"devices/entry/network":                                  "interface vlan virtual-router ike dhcp",
		"devices/entry/network/interface":                        "ethernet aggregate-ethernet vlan loopback tunnel",
		"devices/entry/network/ike":                              "gateway",
		"devices/entry/network/virtual-router":                   "entry",
		"devices/entry/network/virtual-router/entry":             "routing-table interface",
		"devices/entry/vsys":                                     "entry",
		"devices/entry/vsys/entry":                               "address address-group service service-group zone import rulebase log-settings",
		"devices/entry/vsys/entry/rulebase":                      "security nat",
		"devices/entry/vsys/entry/rulebase/security":             "rules",
		"devices/entry/vsys/entry/rulebase/security/rules":       "entry",
		"devices/entry/vsys/entry/rulebase/security/rules/entry": "from to source destination application service action disabled log-end log-start description schedule source-user category tag profile-setting",
		"shared": "address address-group service service-group application log-settings",
	}
	for _, kind := range []string{"ethernet", "aggregate-ethernet"} {
		prefix := "devices/entry/network/interface/" + kind
		rules[prefix] = "entry"
		rules[prefix+"/entry"] = "layer3 layer2 virtual-wire tap aggregate-group comment link-state lacp"
		for _, mode := range []string{"layer3", "layer2", "virtual-wire"} {
			rules[prefix+"/entry/"+mode] = "ip units dhcp-client ipv6 mtu"
		}
	}
	for _, prefix := range []string{"shared", "devices/entry/vsys/entry"} {
		for _, kind := range []string{"address", "address-group", "service", "service-group"} {
			rules[prefix+"/"+kind] = "entry"
		}
		rules[prefix+"/address/entry"] = "ip-netmask ip-range fqdn description tag"
		rules[prefix+"/address-group/entry"] = "static dynamic description tag"
		rules[prefix+"/service/entry"] = "protocol description tag"
		rules[prefix+"/service-group/entry"] = "members description tag"
	}
	rules["devices/entry/deviceconfig/system/dns-setting"] = "servers"
	rules["devices/entry/deviceconfig/system/dns-setting/servers"] = "primary secondary"
	rules["devices/entry/deviceconfig/system/ntp-servers"] = "primary-ntp-server secondary-ntp-server"
	for _, kind := range []string{"primary-ntp-server", "secondary-ntp-server"} {
		rules["devices/entry/deviceconfig/system/ntp-servers/"+kind] = "ntp-server-address"
	}
	for _, kind := range []string{"vlan", "loopback", "tunnel"} {
		prefix := "devices/entry/network/interface/" + kind
		rules[prefix] = "units"
		rules[prefix+"/units"] = "entry"
		rules[prefix+"/units/entry"] = "ip comment dhcp-client ipv6 mtu"
	}
	for _, kind := range []string{"ethernet", "aggregate-ethernet"} {
		for _, mode := range []string{"layer3", "layer2", "virtual-wire"} {
			prefix := "devices/entry/network/interface/" + kind + "/entry/" + mode
			rules[prefix+"/units"] = "entry"
			rules[prefix+"/units/entry"] = "tag ip comment dhcp-client ipv6 mtu"
		}
	}
	prefix := "devices/entry/network/virtual-router/entry/routing-table"
	rules[prefix] = "ip ipv6"
	for _, family := range []string{"ip", "ipv6"} {
		path := prefix + "/" + family
		rules[path] = "static-route"
		rules[path+"/static-route"] = "entry"
		rules[path+"/static-route/entry"] = "destination interface metric admin-dist nexthop no-install"
		rules[path+"/static-route/entry/nexthop"] = "ip-address ipv6-address next-vr discard"
	}
	rules["devices/entry/network/vlan"] = "entry"
	rules["devices/entry/network/vlan/entry"] = "interface virtual-interface"
	rules["devices/entry/vsys/entry/zone"] = "entry"
	rules["devices/entry/vsys/entry/zone/entry"] = "network"
	rules["devices/entry/vsys/entry/zone/entry/network"] = "layer2 layer3 virtual-wire tap"
	prefix = "devices/entry/vsys/entry/rulebase/security/rules/entry/profile-setting"
	rules[prefix] = "group profiles"
	rules[prefix+"/profiles"] = "virus spyware vulnerability url-filtering file-blocking wildfire-analysis data-filtering"
	prefix = "devices/entry/vsys/entry/rulebase/nat"
	rules[prefix] = "rules"
	rules[prefix+"/rules"] = "entry"
	rules[prefix+"/rules/entry"] = "from to source destination service to-interface description source-translation destination-translation"
	rules[prefix+"/rules/entry/source-translation"] = "dynamic-ip-and-port static-ip"
	for _, scope := range []string{"shared", "devices/entry/vsys/entry"} {
		rules[scope+"/service/entry/protocol"] = "tcp udp"
	}

	// Address nodes are only partly modeled; keep other IPv6 controls in coverage.
	for path, children := range rules {
		if strings.Contains(" "+children+" ", " ipv6 ") && strings.Contains(path, "/interface/") {
			rules[path+"/ipv6"] = "enabled address"
			rules[path+"/ipv6/address"] = "entry"
			rules[path+"/ipv6/address/entry"] = ""
		}
	}

	dhcp := "devices/entry/network/dhcp"
	rules[dhcp] = "interface"
	rules[dhcp+"/interface"] = "entry"
	rules[dhcp+"/interface/entry"] = "server relay"
	server := dhcp + "/interface/entry/server"
	rules[server] = "mode probe-ip option ip-pool reserved"
	rules[server+"/option"] = "lease gateway subnet-mask dns"
	rules[server+"/option/lease"] = "timeout unlimited"
	rules[server+"/option/dns"] = "primary secondary"
	rules[server+"/reserved"] = "entry"
	rules[server+"/reserved/entry"] = "mac description"
	relay := dhcp + "/interface/entry/relay"
	rules[relay] = "ip"
	rules[relay+"/ip"] = "enabled server"
	for _, scope := range []string{"shared", "devices/entry/vsys/entry"} {
		p := scope + "/log-settings"
		rules[p] = "syslog"
		rules[p+"/syslog"] = "entry"
		rules[p+"/syslog/entry"] = "server"
		rules[p+"/syslog/entry/server"] = "entry"
		rules[p+"/syslog/entry/server/entry"] = "server transport port format facility"
	}

	counts := map[string]int{}
	var audit func(coverageNode, string)
	audit = func(node coverageNode, path string) {
		allowed, ok := rules[path]
		if !ok {
			return
		}
		for _, child := range node.Children {
			next := child.XMLName.Local
			if path != "" {
				next = path + "/" + next
			}
			matched := false
			for _, key := range strings.Fields(allowed) {
				if child.XMLName.Local == key {
					matched = true
					break
				}
			}
			if !matched {
				counts[next]++
				continue
			}
			audit(child, next)
		}
	}
	audit(root, "")
	keys := make([]string, 0, len(counts))
	for path := range counts {
		keys = append(keys, path)
	}
	sort.Strings(keys)
	for _, path := range keys {
		cfg.Coverage = append(cfg.Coverage, model.CoverageItem{Path: path, Reason: "未対応または部分対応の設定項目", Note: fmt.Sprintf("出現数 %d（設定値は非表示）", counts[path])})
	}
}
