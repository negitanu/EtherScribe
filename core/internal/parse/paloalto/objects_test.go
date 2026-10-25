package paloalto

import (
	"github.com/negitanu/etherscribe/internal/model"
	"os"
	"reflect"
	"testing"
)

func TestSharedAndVsysObjectParity(t *testing.T) {
	b, err := os.ReadFile("../../../testdata/paloalto-object-parity.xml")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse(string(b), "paloalto-object-parity.xml")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Objects.Addresses) != 6 {
		t.Fatalf("addresses=%+v", cfg.Objects.Addresses)
	}
	objects := map[string]model.AddressObject{}
	for _, a := range cfg.Objects.Addresses {
		objects[a.VDOM+"/"+a.Name] = a
	}
	for key, want := range map[string]string{"shared/Shared LAN": "192.0.2.0/24", "vsys1/Shared LAN": "203.0.113.0/24", "vsys2/Shared LAN": "10.0.0.0/24", "shared/Shared Range": "198.51.100.10-198.51.100.20", "vsys1/Web Server": "www.example.com", "vsys1/IPv6 LAN": "2001:db8:1::/64"} {
		if got := objects[key].Value; got != want {
			t.Errorf("%s=%q want %q", key, got, want)
		}
	}
	if objects["shared/Shared LAN"].Type != "subnet" || objects["shared/Shared Range"].Type != "iprange" || !reflect.DeepEqual(objects["shared/Shared LAN"].Tags, []string{"共有 タグ"}) {
		t.Fatal("address normalization/tags lost")
	}
	groups := map[string]model.NamedMembers{}
	for _, g := range cfg.Objects.AddressGroups {
		groups[g.Name] = g
	}
	if !reflect.DeepEqual(groups["Shared Group"].Members, []string{"Shared LAN", "Shared Range"}) {
		t.Fatal("shared group missing")
	}
	if groups["Dynamic Group"].DynamicFilter != "'production' and 'web'" || groups["Dynamic Group"].Description != "動的グループの説明" || !reflect.DeepEqual(groups["Dynamic Group"].Tags, []string{"自動"}) {
		t.Fatalf("dynamic group=%+v", groups["Dynamic Group"])
	}
	services := map[string]model.ServiceObject{}
	for _, s := range cfg.Objects.Services {
		services[s.Name] = s
	}
	for name, proto := range map[string]string{"Shared DNS": "UDP", "Web TLS": "TCP"} {
		s := services[name]
		if s.Protocol != proto || s.SourcePort != "1024-65535" || s.Description == "" || len(s.Tags) == 0 {
			t.Fatalf("service=%+v", s)
		}
	}
	if len(cfg.Objects.ServiceGroups) != 2 || cfg.Objects.ServiceGroups[0].Description == "" {
		t.Fatalf("service groups=%+v", cfg.Objects.ServiceGroups)
	}
}
