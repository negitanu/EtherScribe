package fortigate

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestImplicitVDOMEndPreservesFollowingSections(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/fortigate-vdom-boundaries.conf")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		cfg, err := Parse(string(data), "synthetic.conf")
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Coverage) != 1 || cfg.Coverage[0].Path != "vdom/*/system sdwan" {
			t.Fatalf("unsupported VDOM block missing from coverage: %+v", cfg.Coverage)
		}
		if len(cfg.ParseErrors) != 0 {
			t.Fatalf("unexpected warnings: %v", cfg.ParseErrors)
		}
		if len(cfg.Objects.Addresses) != 2 || len(cfg.Objects.Services) != 1 || len(cfg.Routes) != 1 || len(cfg.Policies) != 4 {
			t.Fatalf("lost following VDOM: addresses=%d services=%d routes=%d policies=%d", len(cfg.Objects.Addresses), len(cfg.Objects.Services), len(cfg.Routes), len(cfg.Policies))
		}
		var ids []string
		for _, p := range cfg.Policies {
			ids = append(ids, p.PolicyID)
		}
		if !reflect.DeepEqual(ids, []string{"20", "3", "100", "7"}) {
			t.Fatalf("evaluation order=%v", ids)
		}
		if !cfg.Policies[0].LogStart || cfg.Policies[1].Description != "更新しても順序を変えない" || cfg.Policies[3].VDOM != "Context B" {
			t.Fatal("policy context/update was lost")
		}
	}
}
func TestNonVDOMMissingNextStillWarns(t *testing.T) {
	_, errors := ParseLines(strings.Split("config firewall address\nedit broken\nset subnet 192.0.2.0 255.255.255.0\nend", "\n"))
	if len(errors) == 0 {
		t.Fatal("non-VDOM malformed boundary was accepted")
	}
}

func TestEditOrderMetadataDoesNotClobberObjectNames(t *testing.T) {
	tree, errs := ParseLines(strings.Split("config firewall address\nedit _edit_order\nset subnet 192.0.2.0 255.255.255.0\nnext\nedit second\nset subnet 198.51.100.0 255.255.255.0\nnext\nend", "\n"))
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	addresses := tree.Map("firewall address")
	if !reflect.DeepEqual(EditNames(addresses), []string{"_edit_order", "second"}) || addresses.Map("_edit_order").String("subnet") == "" {
		t.Fatal("order metadata overwrote a real name")
	}
}
