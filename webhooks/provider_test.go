package webhooks

import (
	"strings"
	"testing"

	"github.com/zatrano/framework/v2/kernel"
)

func TestRegisterRejectsMissingSecret(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", "")
	t.Setenv("WEBHOOK_URL", "")
	err := (&ServiceProvider{}).Register(kernel.NewApplication(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "WEBHOOK_SECRET") {
		t.Fatalf("err=%v", err)
	}
}

func TestRegisterRejectsPlaceholderSecret(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", strings.Join([]string{"zatrano", "webhook", "secret"}, "-"))
	t.Setenv("WEBHOOK_URL", "https://example.test/hook")
	err := (&ServiceProvider{}).Register(kernel.NewApplication(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "WEBHOOK_SECRET") {
		t.Fatalf("err=%v", err)
	}
}

func TestRegisterAcceptsConfiguredSecret(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", "application-webhook-secret")
	t.Setenv("WEBHOOK_URL", "https://example.test/hook")
	app := kernel.NewApplication(t.TempDir())
	if err := (&ServiceProvider{}).Register(app); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Make("webhooks"); err != nil {
		t.Fatal(err)
	}
}
