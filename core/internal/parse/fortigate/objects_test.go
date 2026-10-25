package fortigate

import (
	"github.com/negitanu/etherscribe/internal/model"
	"os"
	"reflect"
	"testing"
)

func TestObjectParityFixture(t *testing.T) {
	b, err := os.ReadFile("../../../testdata/fortigate-object-parity.conf")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse(string(b), "fortigate-object-parity.conf")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ParseErrors) > 0 || len(cfg.Coverage) > 0 {
		t.Fatalf("errors=%v coverage=%v", cfg.ParseErrors, cfg.Coverage)
	}
	if len(cfg.Objects.Addresses) != 7 {
		t.Fatalf("addresses=%+v", cfg.Objects.Addresses)
	}
	addresses := map[string]model.AddressObject{}
	for _, a := range cfg.Objects.Addresses {
		addresses[a.VDOM+"/"+a.Name] = a
	}
	name := `拠点 "A" LAN`
	a := addresses["営業 部/"+name]
	if a.Value != "192.0.2.0/24" || a.Type != "subnet" || a.Description != "更新後の説明 <script>alert(1)</script>" || !reflect.DeepEqual(a.Tags, []string{"業務 用", "重要"}) {
		t.Fatalf("quoted/re-edited address=%+v", a)
	}
	if addresses["開発 部/"+name].Value != "203.0.113.0/24" {
		t.Fatal("same-name contexts were merged")
	}
	for name, want := range map[string]string{"Web Server": "www.example.com", "Client Range": "198.51.100.10-198.51.100.20", "Dynamic Agent": "SYNTHETIC-AGENT", "IPv6 LAN": "2001:db8:1::/64", "IPv6 Range": "2001:db8:2::10-2001:db8:2::20"} {
		if got := addresses["営業 部/"+name].Value; got != want {
			t.Errorf("%s=%q want %q", name, got, want)
		}
	}
	groups := map[string]model.NamedMembers{}
	for _, g := range cfg.Objects.AddressGroups {
		groups[g.Name] = g
	}
	if !reflect.DeepEqual(groups["業務 グループ"].Members, []string{name, "Client Range"}) {
		t.Fatalf("group=%+v", groups["業務 グループ"])
	}
	if !reflect.DeepEqual(groups["IPv6 Group"].Members, []string{"IPv6 LAN", "IPv6 Range"}) {
		t.Fatal("IPv6 group lost")
	}
	services := map[string]model.ServiceObject{}
	for _, s := range cfg.Objects.Services {
		services[s.Name] = s
	}
	s := services["DNS Mixed"]
	if s.Protocol != "TCP/UDP/SCTP" || s.TCPPort != "53, 8053-8054" || s.UDPPort != "53" || s.SCTPPort != "9000" || s.Description != "複数プロトコルの説明" {
		t.Fatalf("service=%+v", s)
	}
	if services["Ping Request"].Port != "type=8,code=0" {
		t.Fatalf("icmp=%+v", services["Ping Request"])
	}
	if len(cfg.Objects.ServiceGroups) != 1 || cfg.Objects.ServiceGroups[0].Description != "サービスグループの説明" {
		t.Fatal("service group description lost")
	}
}
func TestQuotedTokensAndTabs(t *testing.T) {
	tokens := shellSplit(`"" "LAN A" "拠点 \"B\"" 'C D'`)
	if !reflect.DeepEqual(tokens, []string{"", "LAN A", `拠点 "B"`, "C D"}) {
		t.Fatalf("tokens=%q", tokens)
	}
	lines := []string{"config firewall address", `edit "LAN A"`, "set\tsubnet\t192.0.2.0 255.255.255.0", `set comment ""`, "next", "end"}
	tree, errs := ParseLines(lines)
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	a := tree.Map("firewall address").Map("LAN A")
	if a.String("subnet") != "192.0.2.0 255.255.255.0" || a.String("comment") != "" {
		t.Fatalf("tree=%v", tree)
	}
}
