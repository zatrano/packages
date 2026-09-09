package testing_test

import (
	"slices"
	"testing"

	"github.com/zatrano/framework/v2/bootstrap/addons"

	_ "github.com/zatrano/packages/agent"
	_ "github.com/zatrano/packages/apitoken"
	_ "github.com/zatrano/packages/auth"
	_ "github.com/zatrano/packages/backup"
	_ "github.com/zatrano/packages/broadcasting"
	_ "github.com/zatrano/packages/cache"
	_ "github.com/zatrano/packages/flash"
	_ "github.com/zatrano/packages/notification"
	_ "github.com/zatrano/packages/orm"
	_ "github.com/zatrano/packages/queue"
	_ "github.com/zatrano/packages/rag"
	_ "github.com/zatrano/packages/redisx"
)

func TestAuthRequiresClosure(t *testing.T) {
	meta, ok := addons.Lookup("auth")
	if !ok {
		t.Fatal("auth must be registered")
	}
	if !containsAll(meta.Requires, "hashing", "database", "session") {
		t.Fatalf("auth Requires=%v", meta.Requires)
	}
	if containsAny(meta.Requires, "cache", "notification", "authorization", "events") {
		t.Fatalf("optional names must not be Requires: %v", meta.Requires)
	}
	if !containsAll(meta.Optional, "cache", "notification", "authorization", "events") {
		t.Fatalf("auth Optional=%v", meta.Optional)
	}

	got, err := addons.Expand([]string{"auth"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := metaNames(got)
	if !containsAll(names, "auth", "hashing", "database", "session") {
		t.Fatalf("auth closure=%v", names)
	}
}

func TestApitokenTransitiveRequires(t *testing.T) {
	meta, ok := addons.Lookup("apitoken")
	if !ok {
		t.Fatal("apitoken must be registered")
	}
	if !containsAll(meta.Requires, "auth") {
		t.Fatalf("apitoken Requires=%v", meta.Requires)
	}
	if containsAny(meta.Requires, "database") {
		t.Fatal("apitoken must not require database (memory store exists)")
	}
	if !containsAll(meta.Optional, "database") {
		t.Fatalf("apitoken Optional=%v", meta.Optional)
	}

	got, err := addons.Expand([]string{"apitoken"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := metaNames(got)
	if !containsAll(names, "apitoken", "auth", "hashing", "database", "session") {
		t.Fatalf("apitoken transitive closure=%v", names)
	}
}

func TestOptionalNotInRequires(t *testing.T) {
	cases := map[string][]string{
		"queue":        {"database", "cache"},
		"notification": {"view", "broadcasting", "localization", "database"},
		"backup":       {"database"},
		"broadcasting": {"auth"},
		"database":     {"events"},
	}
	for name, optional := range cases {
		meta, ok := addons.Lookup(name)
		if !ok {
			t.Fatalf("%s must be registered", name)
		}
		if containsAny(meta.Requires, optional...) {
			t.Fatalf("%s Requires=%v must not include Optional %v", name, meta.Requires, optional)
		}
		if !containsAll(meta.Optional, optional...) {
			t.Fatalf("%s Optional=%v want %v", name, meta.Optional, optional)
		}
	}
}

func TestExpandMissingRequires(t *testing.T) {
	lookup := func(name string) (addons.Meta, bool) {
		if name == "apitoken" {
			return addons.Meta{Name: "apitoken", Requires: []string{"auth"}}, true
		}
		return addons.Meta{}, false
	}
	_, err := addons.Expand([]string{"apitoken"}, lookup)
	if err == nil {
		t.Fatal("expected missing Requires error")
	}
}

func TestExpandCycleDetection(t *testing.T) {
	_, err := addons.OrderMetas([]addons.Meta{
		{Name: "a", Requires: []string{"b"}},
		{Name: "b", Requires: []string{"a"}},
	})
	if err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestExpandDeterministic(t *testing.T) {
	a, err := addons.Expand([]string{"auth", "flash"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := addons.Expand([]string{"flash", "auth"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	an, bn := metaNames(a), metaNames(b)
	slices.Sort(an)
	slices.Sort(bn)
	if !slices.Equal(an, bn) {
		t.Fatalf("closure not deterministic: %v vs %v", an, bn)
	}
}

func TestLibrariesNotRegistered(t *testing.T) {
	for _, name := range []string{"rag", "agent", "redisx"} {
		if _, ok := addons.Lookup(name); ok {
			t.Fatalf("%s must not register as an addon", name)
		}
	}
}

func TestNegativePathMissingRequires(t *testing.T) {
	cases := []struct {
		name     string
		meta     addons.Meta
		selected string
	}{
		{name: "auth without hashing", selected: "auth", meta: addons.Meta{Name: "auth", Requires: []string{"hashing", "database", "session"}}},
		{name: "auth without database", selected: "auth", meta: addons.Meta{Name: "auth", Requires: []string{"hashing", "database", "session"}}},
		{name: "auth without session", selected: "auth", meta: addons.Meta{Name: "auth", Requires: []string{"hashing", "database", "session"}}},
		{name: "orm without database", selected: "orm", meta: addons.Meta{Name: "orm", Requires: []string{"database"}}},
		{name: "flash without session", selected: "flash", meta: addons.Meta{Name: "flash", Requires: []string{"session"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lookup := func(name string) (addons.Meta, bool) {
				if name == tc.meta.Name {
					return tc.meta, true
				}
				return addons.Meta{}, false
			}
			if _, err := addons.Expand([]string{tc.selected}, lookup); err == nil {
				t.Fatal("expected missing Requires to fail")
			}
		})
	}
}

func TestNegativePathOptionalAbsent(t *testing.T) {
	cases := []addons.Meta{
		{Name: "auth", Requires: []string{"hashing", "database", "session"}, Optional: []string{"cache", "notification", "authorization", "events"}},
		{Name: "queue", Optional: []string{"database", "cache"}},
		{Name: "notification", Optional: []string{"view", "broadcasting", "localization", "database"}},
		{Name: "backup", Optional: []string{"database"}},
		{Name: "broadcasting", Optional: []string{"auth"}},
	}
	for _, meta := range cases {
		t.Run(meta.Name, func(t *testing.T) {
			lookup := func(name string) (addons.Meta, bool) {
				if name == meta.Name {
					return meta, true
				}
				for _, req := range meta.Requires {
					if name == req {
						return addons.Meta{Name: req}, true
					}
				}
				return addons.Meta{}, false
			}
			got, err := addons.Expand([]string{meta.Name}, lookup)
			if err != nil {
				t.Fatal(err)
			}
			names := metaNames(got)
			if containsAny(names, meta.Optional...) {
				t.Fatalf("optional names leaked into closure: %v", names)
			}
		})
	}
}

func metaNames(metas []addons.Meta) []string {
	out := make([]string, 0, len(metas))
	for _, m := range metas {
		out = append(out, m.Name)
	}
	return out
}

func containsAll(have []string, want ...string) bool {
	set := map[string]bool{}
	for _, n := range have {
		set[n] = true
	}
	for _, n := range want {
		if !set[n] {
			return false
		}
	}
	return true
}

func containsAny(have []string, names ...string) bool {
	set := map[string]bool{}
	for _, n := range have {
		set[n] = true
	}
	for _, n := range names {
		if set[n] {
			return true
		}
	}
	return false
}
