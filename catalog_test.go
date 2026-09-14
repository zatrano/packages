package packages

import (
	"testing"

	"github.com/zatrano/framework/v2/kernel"
)

func TestCatalogToolkitAddons(t *testing.T) {
	if len(Catalog) != 10 {
		t.Fatalf("catalog=%d want 10 (3 experimental intelligence + 7 toolkit)", len(Catalog))
	}
	seen := map[string]bool{}
	for _, p := range Catalog {
		if p.Description == "" {
			t.Errorf("%s missing description", p.Name)
		}
		if seen[p.Name] {
			t.Errorf("duplicate %s", p.Name)
		}
		seen[p.Name] = true
		switch p.Name {
		case "ai", "rag", "agent":
			if p.Layer != kernel.LayerIntelligence || p.Stability != "experimental" {
				t.Errorf("%s want intelligence experimental, layer=%s stability=%s", p.Name, p.Layer, p.Stability)
			}
		default:
			if p.Layer != kernel.LayerAddon || p.Kind != kernel.KindLibrary {
				t.Errorf("%s layer=%s kind=%s", p.Name, p.Layer, p.Kind)
			}
		}
	}
	for _, name := range []string{
		"ai", "rag", "agent",
		"toolkit/arr", "toolkit/color", "toolkit/date", "toolkit/html",
		"toolkit/money", "toolkit/num", "toolkit/str",
	} {
		if !seen[name] {
			t.Errorf("missing %s", name)
		}
	}
}
