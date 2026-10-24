package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/negitanu/etherscribe/internal/identify"
)

func testdata(name string) string {
	return filepath.Join("..", "..", "testdata", name)
}

func TestConvertFortiGateJSON(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.json")
	res, err := ConvertFile(ConvertRequest{
		Path: testdata("fortigate-vdom.conf"), Output: out, Format: "json", Preset: "policy",
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "ENC ") {
		t.Fatal("secret leaked")
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["coverage"]; !ok {
		t.Fatal("coverage missing")
	}
	if res.Summary["hostname"] != "FW-VDOM-DEMO" {
		t.Fatalf("summary=%v", res.Summary)
	}
}

func TestConvertPaloAlto(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.json")
	_, err := ConvertFile(ConvertRequest{Path: testdata("paloalto-vsys.xml"), Output: out, Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestConvertJunosEX(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.json")
	res, err := ConvertFile(ConvertRequest{
		Path: testdata("junos-ex-els.set"), Output: out, Format: "json", Preset: "switch-overview",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary["deviceType"] != identify.JunosEX {
		t.Fatalf("deviceType=%v", res.Summary["deviceType"])
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"role": "switch"`) {
		t.Fatalf("expected switch role in %s", raw)
	}
	if !strings.Contains(string(raw), "VLAN10") {
		t.Fatal("expected VLAN10")
	}
}

func TestConvertJunosEXBrace(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.json")
	_, err := ConvertFile(ConvertRequest{Path: testdata("junos-ex-els.conf"), Output: out, Format: "json", Preset: "switch-l3"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCompareDetectsAddedPolicy(t *testing.T) {
	dir := t.TempDir()
	before := testdata("fortigate-vdom.conf")
	afterPath := filepath.Join(dir, "after.conf")
	raw, _ := os.ReadFile(before)
	extra := string(raw) + `
config vdom
 edit branch
  config firewall policy
   edit 99
    set name new-rule
    set srcintf lan
    set dstintf transit0
    set srcaddr all
    set dstaddr all
    set service ALL
    set action accept
   next
  end
 next
end
`
	_ = os.WriteFile(afterPath, []byte(extra), 0o644)
	out := filepath.Join(dir, "diff.json")
	res, err := CompareFiles(CompareRequest{Before: before, After: afterPath, Output: out, Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary["changes"].(int) < 1 {
		t.Fatalf("expected changes, got %v", res.Summary)
	}
}

func TestCompareCrossVendorFails(t *testing.T) {
	_, err := CompareFiles(CompareRequest{
		Before: testdata("fortigate-vdom.conf"),
		After:  testdata("paloalto-vsys.xml"),
		Output: filepath.Join(t.TempDir(), "x.json"),
		Format: "json",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
