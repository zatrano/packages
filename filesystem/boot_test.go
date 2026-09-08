package filesystem

import (
	"strings"
	"testing"

	"github.com/zatrano/framework/v2/kernel"
)

func TestBootRejectsMissingSigningKey(t *testing.T) {
	t.Setenv("APP_KEY", "")
	err := boot(kernel.NewApplication(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "APP_KEY") {
		t.Fatalf("err=%v", err)
	}
}

func TestBootRejectsPlaceholderSigningKey(t *testing.T) {
	t.Setenv("APP_KEY", "zatrano-dev-key")
	err := boot(kernel.NewApplication(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "APP_KEY") {
		t.Fatalf("err=%v", err)
	}
}

func TestBootAcceptsCISigningKey(t *testing.T) {
	t.Setenv("APP_KEY", "zatrano-dev-key-do-not-use-prod!")
	if err := boot(kernel.NewApplication(t.TempDir())); err != nil {
		t.Fatal(err)
	}
}
