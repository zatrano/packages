package hashid

import (
	"strings"
	"testing"

	"github.com/zatrano/framework/v2/kernel"
)

func TestRegisterRejectsMissingSalt(t *testing.T) {
	t.Setenv("HASHID_SALT", "")
	t.Setenv("APP_KEY", "")
	err := (&ServiceProvider{}).Register(kernel.NewApplication(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "HASHID_SALT") {
		t.Fatalf("err=%v", err)
	}
}

func TestRegisterRejectsPlaceholderSalt(t *testing.T) {
	t.Setenv("HASHID_SALT", "zatrano")
	t.Setenv("APP_KEY", "")
	err := (&ServiceProvider{}).Register(kernel.NewApplication(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "HASHID_SALT") {
		t.Fatalf("err=%v", err)
	}
}

func TestRegisterAcceptsConfiguredSalt(t *testing.T) {
	t.Setenv("HASHID_SALT", "application-hashid-salt")
	app := kernel.NewApplication(t.TempDir())
	if err := (&ServiceProvider{}).Register(app); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Make("hashid"); err != nil {
		t.Fatal(err)
	}
}
