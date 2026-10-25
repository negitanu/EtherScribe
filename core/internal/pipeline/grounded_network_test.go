package pipeline

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/negitanu/etherscribe/internal/model"
)

func groundedConfig(t *testing.T, file string) *model.Config {
	t.Helper()
	data, err := os.ReadFile(testdata(file))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseContent(string(data), file)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ParseErrors) > 0 {
		t.Fatalf("parse warnings=%v", cfg.ParseErrors)
	}
	return cfg
}
func TestGroundedJunosSemantics(t *testing.T) {
	cfg := groundedConfig(t, "grounded-junos-network.set")
	peers := map[string]model.BGPNeighbor{}
	for _, b := range cfg.Routing.BGP {
		expectedAS := "64512"
		if b.Context == "BLUE" {
			expectedAS = "64517"
		}
		if b.ASNumber != expectedAS || b.RouterID != "192.0.2.1" {
			t.Fatalf("BGP=%+v", b)
		}
		for _, n := range b.Neighbors {
			peers[n.IP] = n
		}
	}
	if len(peers) != 4 {
		t.Fatalf("peers=%+v", peers)
	}
	n := peers["203.0.113.2"]
	if n.RemoteAS != "64514" || n.LocalAS != "64515" || n.UpdateSource != "192.0.2.1" || n.Group != "WAN" || !reflect.DeepEqual(n.ImportPolicies, []string{"BASE-IN", "SANITY"}) || !reflect.DeepEqual(n.ExportPolicies, []string{"EXPORT-A", "EXPORT-B"}) {
		t.Fatalf("inherited=%+v", n)
	}
	if peers["203.0.113.3"].RemoteAS != "64513" || peers["192.0.2.9"].RemoteAS != "64517" || peers["203.0.113.4"].Activate || !peers["203.0.113.4"].Shutdown {
		t.Fatalf("peer semantics=%+v", peers)
	}
	if len(cfg.Routes) != 5 {
		t.Fatalf("route count=%d", len(cfg.Routes))
	}
	for _, r := range cfg.Routes {
		switch {
		case r.Context == "BLUE":
			if r.Distance != "30" || r.Gateway != "192.0.2.9" {
				t.Fatalf("VRF route=%+v", r)
			}
		case r.Gateway == "192.0.2.3":
			if r.Distance != "25" || r.Metric != "9" || r.Description != "primary and backup" {
				t.Fatalf("qualified=%+v", r)
			}
		case r.Destination == "203.0.113.0/24":
			if r.Distance != "8" || r.Metric != "7" {
				t.Fatalf("primary=%+v", r)
			}
		case r.Destination == "192.0.2.80/32":
			if r.Enabled {
				t.Fatal("deactivated route enabled")
			}
		case r.Destination == "198.51.100.0/24":
			if r.Action != "discard" {
				t.Fatal("discard lost")
			}
		}
	}
	for _, lag := range cfg.LAGs {
		if lag.LACPEnabled != (lag.Name == "ae1") {
			t.Fatalf("LACP=%+v", lag)
		}
	}
	for _, i := range cfg.Interfaces {
		if i.Name == "irb.10" && (len(i.Addresses) != 3 || i.IPAddress != "192.0.2.1/24" || i.RoutingContext != "BLUE") {
			t.Fatalf("IRB=%+v", i)
		}
		if i.Name == "ge-0/0/2" && i.Status != "down" {
			t.Fatal("disable lost")
		}
	}
	// Running the same input repeatedly must not reorder map-backed inventories.
	first, _ := json.Marshal(cfg)
	for j := 0; j < 5; j++ {
		next, _ := json.Marshal(groundedConfig(t, "grounded-junos-network.set"))
		if !bytes.Equal(first, next) {
			t.Fatal("nondeterministic output")
		}
	}
}
func TestGroundedFortiOSNetwork(t *testing.T) {
	cfg := groundedConfig(t, "grounded-fortios-network.conf")
	for _, i := range cfg.Interfaces {
		if i.Name == "port1" {
			if !reflect.DeepEqual(i.Addresses, []string{"192.0.2.1/24", "198.51.100.1/24", "2001:db8:10::1/64", "2001:db8:20::1/64"}) || i.MTU != "9000" || i.RoutingContext != "10" {
				t.Fatalf("interface=%+v", i)
			}
		}
	}
	if len(cfg.LAGs) != 1 || cfg.LAGs[0].LACPEnabled || len(cfg.LAGs[0].Members) != 2 {
		t.Fatalf("LAG=%+v", cfg.LAGs)
	}
	if len(cfg.Routes) != 4 {
		t.Fatalf("routes=%+v", cfg.Routes)
	}
	for _, r := range cfg.Routes {
		switch r.Name {
		case "10":
			if r.RoutingContext != "10" || r.Destination != "0.0.0.0/0" || r.Priority != "50" || r.Distance != "12" || r.Description != "primary & backup" {
				t.Fatalf("route=%+v", r)
			}
		case "20":
			if r.Action != "discard" || r.Enabled {
				t.Fatalf("discard=%+v", r)
			}
		case "40":
			if r.Destination != "" || r.DestinationObject != "Branch" {
				t.Fatalf("destination object=%+v", r)
			}
		case "30":
			if r.Priority != "25" || r.RouteType != "static6" {
				t.Fatalf("IPv6=%+v", r)
			}
		}
	}
}
func TestGroundedPANOSNetwork(t *testing.T) {
	cfg := groundedConfig(t, "grounded-paloalto-network.xml")
	if len(cfg.Interfaces) != 3 {
		t.Fatalf("interfaces=%+v", cfg.Interfaces)
	}
	for _, i := range cfg.Interfaces {
		if i.VDOM != "vsys1" || i.Context != "vsys1" || i.RoutingContext != "VR-BLUE" {
			t.Fatalf("membership=%+v", i)
		}
		if i.Name == "ethernet1/1" && (i.MTU != "9000" || len(i.Addresses) != 2 || i.IPv6Enabled == nil || !*i.IPv6Enabled) {
			t.Fatalf("dual-stack=%+v", i)
		}
		if i.Name == "loopback.1" && i.IPAddress != "2001:db8::1/128" {
			t.Fatalf("loopback=%+v", i)
		}
	}
	if len(cfg.Routes) != 1 || cfg.Routes[0].Metric != "30" || cfg.Routes[0].Distance != "20" {
		t.Fatalf("route=%+v", cfg.Routes)
	}
	for _, c := range cfg.Coverage {
		if strings.HasSuffix(c.Path, "/interface") || strings.HasSuffix(c.Path, "/ipv6") || strings.HasSuffix(c.Path, "/mtu") {
			t.Fatalf("supported field in coverage: %s", c.Path)
		}
	}
}
func TestGroundedReports(t *testing.T) {
	for file, markers := range map[string][]string{
		"grounded-junos-network.set":    {"64514", "64515", "EXPORT-B", "primary and backup", "2001:db8:10::1/64"},
		"grounded-fortios-network.conf": {"2001:db8:20::1/64", "9000", "50"},
		"grounded-paloalto-network.xml": {"VR-BLUE", "2001:db8::1/128", "9000"},
	} {
		t.Run(file, func(t *testing.T) {
			for _, format := range []string{"json", "html", "excel"} {
				res, err := ConvertFile(ConvertRequest{Path: testdata(file), Format: format, Preset: "all"})
				if err != nil {
					t.Fatal(err)
				}
				text := string(res.Bytes)
				if format == "json" && !json.Valid(res.Bytes) {
					t.Fatal("invalid JSON")
				}
				if format == "excel" {
					z, err := zip.NewReader(bytes.NewReader(res.Bytes), int64(len(res.Bytes)))
					if err != nil {
						t.Fatal(err)
					}
					var all strings.Builder
					for _, f := range z.File {
						r, err := f.Open()
						if err != nil {
							t.Fatal(err)
						}
						b, err := io.ReadAll(r)
						r.Close()
						if err != nil {
							t.Fatal(err)
						}
						if strings.HasSuffix(f.Name, ".xml") {
							d := xml.NewDecoder(bytes.NewReader(b))
							for {
								_, err := d.Token()
								if err == io.EOF {
									break
								}
								if err != nil {
									t.Fatal(err)
								}
							}
						}
						all.Write(b)
					}
					text = all.String()
				}
				for _, m := range markers {
					if !strings.Contains(text, m) {
						t.Fatalf("%s missing %s", format, m)
					}
				}
			}
		})
	}
}

func TestGroundedJunosBraceSetEquivalence(t *testing.T) {
	a := groundedConfig(t, "grounded-junos-compact.conf")
	b := groundedConfig(t, "grounded-junos-compact.set")
	a.SourceFile = ""
	b.SourceFile = ""
	if !reflect.DeepEqual(a, b) {
		aj, _ := json.Marshal(a)
		bj, _ := json.Marshal(b)
		t.Fatalf("brace=%s\nset=%s", aj, bj)
	}
	if len(a.Routes) != 3 || a.Routes[2].RouteType != "static6" || a.Routes[2].Gateway != "2001:db8::2" {
		t.Fatalf("routes=%+v", a.Routes)
	}
	if a.Routing.BGP[0].Neighbors[1].Description != `edge; {test} "quoted" # literal` {
		t.Fatalf("neighbors=%+v", a.Routing.BGP[0].Neighbors)
	}
}

func TestGroundedJunosStatementOrder(t *testing.T) {
	cfg := groundedConfig(t, "grounded-junos-network.set")
	data, err := os.ReadFile(testdata("grounded-junos-network.set"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	sort.Strings(lines)
	reordered, err := ParseContent(strings.Join(lines, "\n"), "synthetic.set")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.Routing.BGP, reordered.Routing.BGP) {
		t.Fatalf("BGP changed after reorder: %+v", reordered.Routing.BGP)
	}
	order := func(routes []model.Route) {
		sort.Slice(routes, func(i, j int) bool {
			a, b := routes[i], routes[j]
			return a.Context+a.Destination+a.Gateway < b.Context+b.Destination+b.Gateway
		})
	}
	order(cfg.Routes)
	order(reordered.Routes)
	if !reflect.DeepEqual(cfg.Routes, reordered.Routes) {
		t.Fatalf("route semantics changed after reorder: %+v", reordered.Routes)
	}
}
