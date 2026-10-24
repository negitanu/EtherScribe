package diff

import (
	"strings"
	"testing"
)

func TestHTMLDiffEscapesConfigValues(t *testing.T) {
	b, err := RenderHTML(&Report{BeforeFile: "<img src=x>", Changes: []Change{{Kind: Added, Entity: "interface", Key: "<script>", After: map[string]any{"alias": "</pre><script>alert(1)</script>"}}}})
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, "<script>") || strings.Contains(s, "<img src=x>") {
		t.Fatal("unescaped input")
	}
	if !strings.Contains(s, "<table>") || !strings.Contains(s, "追加") {
		t.Fatal("missing readable diff")
	}
}
