package junos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/negitanu/etherscribe/internal/model"
)

func TestParseEXSetFixture(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "junos-ex-els.set"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseEX(string(b), "junos-ex-els.set")
	if err != nil {
		t.Fatal(err)
	}
	assertEXCore(t, cfg)
}

func TestParseEXBraceFixture(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "junos-ex-els.conf"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseEX(string(b), "junos-ex-els.conf")
	if err != nil {
		t.Fatal(err)
	}
	assertEXCore(t, cfg)
}

func assertEXCore(t *testing.T, cfg *model.Config) {
	t.Helper()
	if cfg.DeviceInfo.Role != model.RoleSwitch || cfg.DeviceInfo.Hostname != "SW-EX-LAB" {
		t.Fatalf("device=%+v", cfg.DeviceInfo)
	}
	vlanOK := false
	for _, v := range cfg.VLANs {
		if v.Name == "VLAN10" && v.VLANID == "10" && v.L3Interface == "irb.10" {
			vlanOK = true
		}
	}
	if !vlanOK {
		t.Fatalf("vlans=%+v", cfg.VLANs)
	}
	access := false
	for _, i := range cfg.Interfaces {
		if i.Name == "ge-0/0/1" && i.SwitchportMode == "access" && i.AccessVLAN == "VLAN10" {
			access = true
		}
	}
	if !access {
		t.Fatalf("interfaces=%+v", cfg.Interfaces)
	}
	if cfg.STP == nil || cfg.STP.Protocol != "rstp" {
		t.Fatalf("stp=%+v", cfg.STP)
	}
	vrf := false
	for _, c := range cfg.VirtualContexts {
		if c.ID == "VRF-A" && c.Type == model.ContextRoutingInstance {
			vrf = true
		}
	}
	if !vrf {
		t.Fatalf("contexts=%+v", cfg.VirtualContexts)
	}
	routeVRF := false
	for _, r := range cfg.Routes {
		if r.Context == "VRF-A" && r.Destination == "192.0.2.0/24" {
			routeVRF = true
		}
	}
	if !routeVRF {
		t.Fatalf("routes=%+v", cfg.Routes)
	}
}
