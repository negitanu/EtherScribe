package paloalto

import (
	"encoding/json"
	"github.com/negitanu/etherscribe/internal/model"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestNetworkAndOperationalFixture(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/paloalto-network-ops.xml")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse(string(data), "synthetic.xml")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Interfaces) != 12 || len(cfg.LAGs) != 1 || len(cfg.VLANs) != 1 {
		t.Fatalf("interfaces=%d lags=%d vlans=%d", len(cfg.Interfaces), len(cfg.LAGs), len(cfg.VLANs))
	}
	byName := map[string]model.Interface{}
	for _, iface := range cfg.Interfaces {
		byName[iface.Name] = iface
	}
	if byName["ethernet1/3"].Mode != "tap" || byName["ethernet1/5"].Mode != "virtual-wire" || byName["ethernet1/2"].Status != "down" {
		t.Fatal("mode or configured link state lost")
	}
	if byName["ethernet1/1"].LAGParent != "ae1" || byName["ethernet1/1"].VDOM != "vsys1" || byName["ae1"].Mode != "layer2" || !reflect.DeepEqual(cfg.LAGs[0].Members, []string{"ethernet1/1", "ethernet1/2"}) || !cfg.LAGs[0].LACPEnabled || cfg.LAGs[0].Context != "vsys1" {
		t.Fatalf("LAG=%+v", cfg.LAGs)
	}
	if byName["ae1.10"].VLANID != "10" || byName["ae1.10"].Zone != "L2-LAB" || byName["ethernet1/4.20"].ParentInterface != "ethernet1/4" {
		t.Fatal("subinterfaces or zones lost")
	}
	if len(byName["ethernet1/4"].Addresses) != 2 || byName["vlan.30"].AddressMode != "dhcp" || byName["loopback.1"].InterfaceType != "loopback" || byName["tunnel.1"].InterfaceType != "tunnel" {
		t.Fatal("addresses / logical interfaces lost")
	}
	if cfg.VLANs[0].L3Interface != "vlan.10" || cfg.VLANs[0].Context != "vsys1" || !reflect.DeepEqual(cfg.VLANs[0].Members, []string{"ae1.10"}) {
		t.Fatalf("VLAN=%+v", cfg.VLANs)
	}
	s := cfg.System
	if s.ManagementIP != "192.0.2.10" || s.ManagementNetmask != "255.255.255.0" || s.ManagementGateway != "192.0.2.1" || s.Domain != "example.test" || s.DNSPrimary != "192.0.2.53" || s.DNSSecondary != "198.51.100.53" || s.Timezone != "Asia/Tokyo" || !reflect.DeepEqual(s.NTPServers, []string{"ntp1.example.test", "ntp2.example.test"}) {
		t.Fatalf("management=%+v", s)
	}
	if len(cfg.Routes) != 3 || cfg.Routes[1].Gateway != "2001:db8::1" || cfg.Routes[1].Interface != "tunnel.1" || cfg.Routes[1].Metric != "20" || cfg.Routes[1].Distance != "" || cfg.Routes[2].Gateway != "discard" || cfg.Routes[2].Enabled {
		t.Fatalf("routes=%+v", cfg.Routes)
	}
	p := cfg.Policies[0]
	if !reflect.DeepEqual(p.Applications, []string{"web-browsing", "ssl"}) || !reflect.DeepEqual(p.Service, []string{"application-default"}) || !reflect.DeepEqual(p.SecurityProfiles, []string{"AV:AV-LAB", "IPS:IPS-LAB"}) || !p.LogStart || p.Schedule != "business-hours" || p.SourceUsers[0] != "known-user" || p.URLCategories[0] != "business-and-economy" || p.Tags[0] != "lab" {
		t.Fatalf("policy=%+v", p)
	}
	paths := map[string]bool{}
	for _, c := range cfg.Coverage {
		paths[c.Path] = true
	}
	for _, suffix := range []string{"devices/entry/deviceconfig/system/update-server", "devices/entry/network/virtual-router/entry/protocol", "devices/entry/vsys/entry/rulebase/security/rules/entry/source-hip"} {
		if !paths[suffix] {
			t.Errorf("missing partial coverage for %s", suffix)
		}
	}
	coverage, _ := json.Marshal(cfg.Coverage)
	if strings.Contains(string(coverage), "UNMODELED-HIP-LAB") || strings.Contains(string(coverage), "updates.example.test") || strings.Contains(string(coverage), "Web Rule") {
		t.Fatal("coverage contains configuration values")
	}
}
