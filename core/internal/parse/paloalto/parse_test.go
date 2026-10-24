package paloalto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseVsysObjects(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "paloalto-vsys.xml"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse(string(b), "paloalto-vsys.xml")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.DeviceInfo.VDOMList) != 2 {
		t.Fatalf("vsys=%v", cfg.DeviceInfo.VDOMList)
	}
	byName := map[string]string{}
	vdom := map[string]string{}
	kind := map[string]string{}
	for _, a := range cfg.Objects.Addresses {
		byName[a.Name] = a.Value
		vdom[a.Name] = a.VDOM
		kind[a.Name] = a.Type
	}
	if byName["LAN"] != "10.0.0.0/24" || vdom["LAN"] != "vsys1" || kind["LAN"] != "ip-netmask" {
		t.Fatalf("LAN value=%q vdom=%q type=%q", byName["LAN"], vdom["LAN"], kind["LAN"])
	}
	if kind["WEB"] != "fqdn" || byName["WEB"] != "www.example.com" || vdom["WEB"] != "vsys1" {
		t.Fatalf("WEB=%q %q %q", kind["WEB"], byName["WEB"], vdom["WEB"])
	}
	if byName["REMOTE"] != "203.0.113.0/24" || vdom["REMOTE"] != "vsys2" {
		t.Fatalf("REMOTE value=%q vdom=%q", byName["REMOTE"], vdom["REMOTE"])
	}
	if len(cfg.Objects.AddressGroups) != 1 || cfg.Objects.AddressGroups[0].Name != "LAN-GRP" ||
		len(cfg.Objects.AddressGroups[0].Members) != 1 || cfg.Objects.AddressGroups[0].Members[0] != "LAN" {
		t.Fatalf("groups=%+v", cfg.Objects.AddressGroups)
	}
	if len(cfg.Objects.Services) != 1 || cfg.Objects.Services[0].Protocol != "tcp" || cfg.Objects.Services[0].Port != "8080" {
		t.Fatalf("services=%+v", cfg.Objects.Services)
	}
	if len(cfg.Objects.ServiceGroups) != 1 || cfg.Objects.ServiceGroups[0].Members[0] != "HTTP-ALT" {
		t.Fatalf("svc groups=%+v", cfg.Objects.ServiceGroups)
	}
	zones := map[string]string{}
	ifaces := map[string]string{}
	for _, iface := range cfg.Interfaces {
		zones[iface.Name] = iface.Zone
		ifaces[iface.Name] = iface.VDOM
	}
	if zones["ethernet1/1"] != "trust" || ifaces["ethernet1/1"] != "vsys1" {
		t.Fatalf("eth1/1 zone=%q vdom=%q", zones["ethernet1/1"], ifaces["ethernet1/1"])
	}
	if zones["ethernet1/2"] != "untrust" || ifaces["ethernet1/2"] != "vsys2" {
		t.Fatalf("eth1/2 zone=%q vdom=%q", zones["ethernet1/2"], ifaces["ethernet1/2"])
	}
	if len(cfg.Policies) != 1 || !cfg.Policies[0].Enabled || !cfg.Policies[0].LogEnabled {
		t.Fatalf("policy=%+v", cfg.Policies)
	}
	if len(cfg.NATPolicies) != 1 || cfg.NATPolicies[0].Name != "outbound-snat" || cfg.NATPolicies[0].NATType != "snat" {
		t.Fatalf("nat=%+v", cfg.NATPolicies)
	}
	if len(cfg.VPN.IPSecPhase1) != 1 || cfg.VPN.IPSecPhase1[0].Name != "TO-PEER" || !cfg.VPN.IPSecPhase1[0].PSKConfigured {
		t.Fatalf("ike=%+v", cfg.VPN.IPSecPhase1)
	}
	if cfg.VPN.IPSecPhase1[0].RemoteGateway != "203.0.113.1" || cfg.VPN.IPSecPhase1[0].IKEVersion != "2" {
		t.Fatalf("ike detail=%+v", cfg.VPN.IPSecPhase1[0])
	}
}
