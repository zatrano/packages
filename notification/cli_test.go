package notification

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/kernel"
)

func TestNotificationCLIRegistered(t *testing.T) {
	meta, ok := addons.Lookup("notification")
	if !ok || meta.CLI == nil {
		t.Fatal("notification Meta.CLI is not registered")
	}
	app := kernel.NewApplication(t.TempDir())
	cmds := meta.CLI(app)
	found := false
	for _, c := range cmds {
		if c.Name == "make:notification" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("missing make:notification")
	}
}

func TestMakeNotificationWritesFile(t *testing.T) {
	dir := t.TempDir()
	cmd := &MakeNotificationCommand{app: kernel.NewApplication(dir)}
	if err := cmd.Handle([]string{"OrderShipped"}); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "app", "notifications", "order_shipped.go")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected %s: %v", want, err)
	}
}
