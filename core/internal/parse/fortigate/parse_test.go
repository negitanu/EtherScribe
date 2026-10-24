package fortigate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseVDOMFixture(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "fortigate-vdom.conf"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse(string(b), "fortigate-vdom.conf")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DeviceInfo.Hostname != "FW-VDOM-DEMO" {
		t.Fatalf("hostname=%q", cfg.DeviceInfo.Hostname)
	}
	if len(cfg.Interfaces) < 4 {
		t.Fatalf("interfaces=%d", len(cfg.Interfaces))
	}
	if len(cfg.Policies) < 2 {
		t.Fatalf("policies=%d", len(cfg.Policies))
	}
	if len(cfg.DeviceInfo.VDOMList) < 2 {
		t.Fatalf("vdoms=%v", cfg.DeviceInfo.VDOMList)
	}
}
