package identify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIdentifyFortiGateFixture(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fortigate-vdom.conf"))
	if err != nil {
		t.Fatal(err)
	}
	id := Identify(string(b))
	if id == nil || id.Type != FortiGate {
		t.Fatalf("expected FortiGate, got %#v", id)
	}
	if id.Model == "" || id.OSVersion == "" {
		t.Fatalf("expected model/version, got %#v", id)
	}
}

func TestIdentifyPaloAltoFixture(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "paloalto-vsys.xml"))
	if err != nil {
		t.Fatal(err)
	}
	id := Identify(string(b))
	if id == nil || id.Type != PaloAlto {
		t.Fatalf("expected PaloAlto, got %#v", id)
	}
}

func TestIdentifyUnknown(t *testing.T) {
	if Identify("hello world") != nil {
		t.Fatal("expected nil")
	}
}

func TestIdentifyJunosEXSet(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "junos-ex-els.set"))
	if err != nil {
		t.Fatal(err)
	}
	id := Identify(string(b))
	if id == nil || id.Type != JunosEX {
		t.Fatalf("expected JunosEX, got %#v", id)
	}
}

func TestIdentifyJunosEXBrace(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "junos-ex-els.conf"))
	if err != nil {
		t.Fatal(err)
	}
	id := Identify(string(b))
	if id == nil || id.Type != JunosEX {
		t.Fatalf("expected JunosEX, got %#v", id)
	}
}

func TestFormatFortiGateDisplayModel(t *testing.T) {
	if got := FormatFortiGateDisplayModel("60F"); got != "FortiGate 60F" {
		t.Fatalf("got %s", got)
	}
	if got := FormatFortiGateDisplayModel("33E1"); got != "FortiGate 3301E" {
		t.Fatalf("compact got %s", got)
	}
}
