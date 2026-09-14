package packages

import (
	"testing"

	"github.com/zatrano/framework/v2/kernel"
)

func TestCatalogToolkitAddons(t *testing.T) {
	if len(Catalog) != 7 {
		t.Fatalf("toolkit catalog=%d want 7", len(Catalog))
	}
	seen := map[string]bool{}
	for _, p := range Catalog {
		if p.Layer != kernel.LayerAddon || p.Kind != kernel.KindLibrary {
			t.Errorf("%s layer=%s kind=%s", p.Name, p.Layer, p.Kind)
		}
		if p.Description == "" {
			t.Errorf("%s missing description", p.Name)
		}
		if seen[p.Name] {
			t.Errorf("duplicate %s", p.Name)
		}
		seen[p.Name] = true
	}
	for _, name := range []string{
		"toolkit/arr", "toolkit/color", "toolkit/date", "toolkit/html",
		"toolkit/money", "toolkit/num", "toolkit/str",
	} {
		if !seen[name] {
			t.Errorf("missing %s", name)
		}
	}
}
