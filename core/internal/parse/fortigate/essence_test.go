package fortigate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/negitanu/etherscribe/internal/model"
)

func TestParseEssenceFixture(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "fortigate-essence.conf"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse(string(b), "fortigate-essence.conf")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DeviceInfo.Hostname != "FW-ESSENCE" {
		t.Fatalf("hostname=%q", cfg.DeviceInfo.Hostname)
	}
	if len(cfg.Objects.Addresses) < 2 {
		t.Fatalf("addresses=%d", len(cfg.Objects.Addresses))
	}
	if len(cfg.Objects.AddressGroups) < 1 {
		t.Fatalf("addrgrp=%d", len(cfg.Objects.AddressGroups))
	}
	if len(cfg.NATPolicies) < 2 {
		t.Fatalf("nat=%d", len(cfg.NATPolicies))
	}
	if len(cfg.VPN.IPSecPhase1) != 1 || !cfg.VPN.IPSecPhase1[0].PSKConfigured {
		t.Fatalf("vpn p1=%+v", cfg.VPN.IPSecPhase1)
	}
	if cfg.VPN.IPSecPhase1[0].RemoteGateway == "" {
		t.Fatal("missing remote-gw")
	}
	// Ensure raw ENC secret not stored on model beyond presence flag
	for _, p1 := range cfg.VPN.IPSecPhase1 {
		if p1.Proposal == "ENC hidden-psk-value" {
			t.Fatal("psk leaked into proposal")
		}
	}
	if len(cfg.Policies) != 1 || cfg.Policies[0].Action != "allow" {
		t.Fatalf("policy=%+v", cfg.Policies)
	}
	if len(cfg.Policies[0].SecurityProfiles) == 0 {
		t.Fatal("expected security profile tags")
	}
	if !hasProfile(cfg, "antivirus", "default") || !hasProfile(cfg, "ssl-inspection", "deep-inspection") {
		t.Fatalf("profiles=%+v", cfg.SecurityProfiles)
	}
	foundAV := false
	for _, p := range cfg.SecurityProfilesDetail.Antivirus {
		if p.Name == "default" {
			if p.ScanMode != "quick" || !containsStr(p.Protocols, "http") {
				t.Fatalf("av=%+v", p)
			}
			foundAV = true
		}
	}
	if !foundAV {
		t.Fatal("missing antivirus detail")
	}
	foundCat := false
	for _, p := range cfg.SecurityProfilesDetail.Webfilter {
		if p.Name == "default" && containsStr(p.Categories, "26") {
			foundCat = true
		}
	}
	if !foundCat {
		t.Fatalf("webfilter=%+v", cfg.SecurityProfilesDetail.Webfilter)
	}
	var known, unknown bool
	for _, p := range cfg.SecurityProfilesDetail.AppControl {
		if p.Name != "default" {
			continue
		}
		if !containsStr(p.Categories, "23") {
			t.Fatalf("app categories=%v", p.Categories)
		}
		for _, app := range p.Applications {
			switch app.AppID {
			case "900001":
				if app.AppName != "ExampleChat" || app.Category != "Collaboration" || app.Risk != "2" || app.Action != "block" {
					t.Fatalf("resolved app=%+v", app)
				}
				known = true
			case "900002":
				if app.AppName != "900002" {
					t.Fatalf("unknown app=%+v", app)
				}
				unknown = true
			}
		}
	}
	if !known || !unknown {
		t.Fatalf("app control=%+v", cfg.SecurityProfilesDetail.AppControl)
	}
	foundRule := false
	for _, p := range cfg.SecurityProfilesDetail.IPS {
		if p.Name == "default" && containsStr(p.Signatures, "1001") {
			foundRule = true
		}
	}
	if !foundRule {
		t.Fatalf("ips=%+v", cfg.SecurityProfilesDetail.IPS)
	}
	modes := map[string]string{}
	for _, p := range cfg.SecurityProfilesDetail.SSLInspection {
		modes[p.Name] = p.Mode
	}
	if modes["deep-inspection"] != "deep-inspection" || modes["certificate-inspection"] != "certificate-inspection" {
		t.Fatalf("ssl=%v", modes)
	}
	for _, c := range cfg.Coverage {
		if strings.Contains(c.Path, "antivirus profile") {
			t.Fatalf("supported profile marked unsupported: %+v", c)
		}
	}
	if cfg.HA.Mode != "active-passive" || !cfg.HA.PasswordConfigured || len(cfg.HA.HeartbeatInterfaces) < 2 {
		t.Fatalf("ha=%+v", cfg.HA)
	}
	if cfg.HA.GroupName != "CLUSTER1" || cfg.HA.Priority != "200" {
		t.Fatalf("ha group=%+v", cfg.HA)
	}
	if len(cfg.Logging.SyslogServers) != 1 || cfg.Logging.SyslogServers[0].Server != "192.0.2.50" {
		t.Fatalf("syslog=%+v", cfg.Logging.SyslogServers)
	}
	if len(cfg.Logging.SNMP) != 1 || !cfg.Logging.SNMP[0].CommunityConfigured {
		t.Fatalf("snmp=%+v", cfg.Logging.SNMP)
	}
	if cfg.Logging.FortiAnalyzerServer != "192.0.2.60" {
		t.Fatalf("faz=%+v", cfg.Logging)
	}
	if len(cfg.VPN.SSLVPN) < 2 {
		t.Fatalf("sslvpn=%+v", cfg.VPN.SSLVPN)
	}
	if len(cfg.DHCPServers) != 1 || cfg.DHCPServers[0].StartIP != "192.0.2.100" {
		t.Fatalf("dhcp=%+v", cfg.DHCPServers)
	}
	if len(cfg.Routing.OSPF) != 1 || cfg.Routing.OSPF[0].RouterID != "1.1.1.1" {
		t.Fatalf("ospf=%+v", cfg.Routing.OSPF)
	}
	if len(cfg.Routing.BGP) != 1 || cfg.Routing.BGP[0].ASNumber != "65001" || len(cfg.Routing.BGP[0].Neighbors) != 1 {
		t.Fatalf("bgp=%+v", cfg.Routing.BGP)
	}
	if len(cfg.Routing.PolicyRoutes) != 1 {
		t.Fatalf("policy routes=%+v", cfg.Routing.PolicyRoutes)
	}
}

func hasProfile(cfg *model.Config, kind, name string) bool {
	for _, p := range cfg.SecurityProfiles {
		if p.ProfileType == kind && p.Name == name {
			return true
		}
	}
	return false
}

func containsStr(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
