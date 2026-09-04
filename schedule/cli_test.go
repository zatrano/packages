package schedule

import (
	"testing"

	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/kernel"
)

func TestScheduleCLIRegistered(t *testing.T) {
	meta, ok := addons.Lookup("schedule")
	if !ok || meta.CLI == nil {
		t.Fatal("schedule Meta.CLI is not registered")
	}
	cmds := meta.CLI(kernel.NewApplication(t.TempDir()))
	want := map[string]bool{"schedule:run": false, "schedule:list": false}
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
