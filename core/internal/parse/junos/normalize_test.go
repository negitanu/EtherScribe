package junos

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeSetAndBraceHost(t *testing.T) {
	setB, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "junos-ex-els.set"))
	if err != nil {
		t.Fatal(err)
	}
	braceB, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "junos-ex-els.conf"))
	if err != nil {
		t.Fatal(err)
	}
	setStmts, _ := Normalize(string(setB))
	braceStmts, _ := Normalize(string(braceB))
	if !hasPath(setStmts, "system", "host-name", "SW-EX-LAB") {
		t.Fatalf("set missing host-name: %+v", setStmts[:min(5, len(setStmts))])
	}
	if !hasPath(braceStmts, "system", "host-name", "SW-EX-LAB") {
		t.Fatalf("brace missing host-name")
	}
	if !hasPath(setStmts, "vlans", "VLAN10", "vlan-id", "10") {
		t.Fatal("set missing vlan")
	}
	if !hasPath(braceStmts, "vlans", "VLAN10", "vlan-id", "10") {
		t.Fatal("brace missing vlan")
	}
}

func hasPath(stmts []Statement, prefix ...string) bool {
	for _, st := range stmts {
		if PathHas(st, prefix...) && len(st.Path) >= len(prefix) {
			ok := true
			for i := range prefix {
				if st.Path[i] != prefix[i] {
					ok = false
					break
				}
			}
			if ok && len(st.Path) >= len(prefix) {
				// exact prefix match at start
				if len(prefix) == len(st.Path) || true {
					match := true
					for i := range prefix {
						if st.Path[i] != prefix[i] {
							match = false
						}
					}
					if match {
						return true
					}
				}
			}
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
