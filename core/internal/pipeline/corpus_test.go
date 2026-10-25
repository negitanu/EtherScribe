package pipeline

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestComplexSyntheticCorpus(t *testing.T) {
	cases := []struct {
		file, host                                   string
		interfaces, routes, policies, addresses, nat int
	}{
		{"corpus-fortios-complex.conf", "CORPUS-FG", 4, 2, 4, 4, 2},
		{"corpus-junos-complex.set", "CORPUS-EX", 9, 3, 0, 0, 0},
		{"corpus-paloalto-complex.xml", "CORPUS-PA", 2, 1, 4, 3, 2},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(testdata(tc.file))
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := ParseContent(string(data), tc.file)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.DeviceInfo.Hostname != tc.host {
				t.Fatalf("hostname=%q", cfg.DeviceInfo.Hostname)
			}
			got := []int{len(cfg.Interfaces), len(cfg.Routes), len(cfg.Policies), len(cfg.Objects.Addresses), len(cfg.NATPolicies)}
			want := []int{tc.interfaces, tc.routes, tc.policies, tc.addresses, tc.nat}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("interfaces/routes/policies/addresses/NAT: got %v want %v", got, want)
			}
			if len(cfg.ParseErrors) > 0 {
				t.Fatalf("unexpected parse warnings: %v", cfg.ParseErrors)
			}
			switch tc.host {
			case "CORPUS-FG":
				ids, contexts := []string{}, []string{}
				for _, p := range cfg.Policies {
					ids = append(ids, p.PolicyID)
					contexts = append(contexts, p.VDOM)
				}
				if !reflect.DeepEqual(ids, []string{"90", "4", "90", "4"}) || !reflect.DeepEqual(contexts, []string{"Tenant A", "Tenant A", "Tenant B", "Tenant B"}) {
					t.Fatalf("policy order/context=%v %v", ids, contexts)
				}
				for _, a := range cfg.Objects.Addresses {
					if a.Name == "App Network" && a.Value != "192.0.2.0/24" && a.Value != "198.51.100.0/24" {
						t.Fatalf("CIDR=%q", a.Value)
					}
				}
				if len(cfg.Objects.AddressGroups) != 2 || len(cfg.Objects.Services) != 2 {
					t.Fatal("context objects lost")
				}
			case "CORPUS-EX":
				if len(cfg.VLANs) != 3 || len(cfg.LAGs) != 1 || cfg.STP == nil || cfg.STP.Protocol != "rstp" {
					t.Fatal("L2 topology lost")
				}
				contexts := map[string]bool{}
				for _, r := range cfg.Routes {
					contexts[r.Context] = true
				}
				if !contexts["BLUE"] || !contexts["GREEN"] {
					t.Fatalf("route contexts=%v", contexts)
				}
			case "CORPUS-PA":
				contexts := []string{}
				for _, p := range cfg.Policies {
					contexts = append(contexts, p.VDOM)
				}
				if !reflect.DeepEqual(contexts, []string{"vsys1", "vsys1", "vsys2", "vsys2"}) {
					t.Fatalf("vsys/order=%v", contexts)
				}
			}
			for _, format := range []string{"json", "html", "excel"} {
				result, err := ConvertFile(ConvertRequest{Path: testdata(tc.file), Format: format, Preset: "all"})
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Bytes) == 0 {
					t.Fatalf("empty %s", format)
				}
				if format == "json" && !json.Valid(result.Bytes) {
					t.Fatal("invalid JSON")
				}
			}
		})
	}
}

func TestFutureCorpusFormatsAreRejected(t *testing.T) {
	for _, file := range []string{"corpus-cisco-ios.conf", "corpus-cisco-nxos.conf", "corpus-paloalto-cli.set"} {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(testdata(file))
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := ParseContent(string(data), file)
			if err == nil || cfg != nil {
				t.Fatal("unsupported format accepted")
			}
		})
	}
}
