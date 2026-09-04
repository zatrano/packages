package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/kernel"
)

func TestAuthCLIRegistered(t *testing.T) {
	meta, ok := addons.Lookup("auth")
	if !ok || meta.CLI == nil {
		t.Fatal("auth Meta.CLI is not registered")
	}
	cmds := meta.CLI(kernel.NewApplication(t.TempDir()))
	want := map[string]bool{"make:auth": false, "make:dashboard": false}
	for _, c := range cmds {
		if _, ok := want[c.Name]; ok {
			want[c.Name] = true
		}
	}
	for name, ok := range want {
		if !ok {
			t.Fatalf("missing command %s", name)
		}
	}
}

func TestMakeAuthViewsUsesEmbeddedStubs(t *testing.T) {
	dir := t.TempDir()
	cmd := &MakeAuthCommand{app: kernel.NewApplication(dir)}
	if err := cmd.Handle([]string{"--views"}); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "app", "views", "layouts", "auth.html")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected embedded stub at %s: %v", want, err)
	}
}
