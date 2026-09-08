package social

import (
	"errors"
	"testing"

	"github.com/zatrano/framework/v2/kernel"
)

func TestRegisterDoesNotInventPlaceholderCredentials(t *testing.T) {
	t.Setenv("GITHUB_CLIENT_ID", "")
	t.Setenv("GITHUB_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	SetAllowStubProviders(true)
	t.Cleanup(func() { SetAllowStubProviders(true) })

	app := kernel.NewApplication(t.TempDir())
	if err := (&ServiceProvider{}).Register(app); err != nil {
		t.Fatal(err)
	}
	mgr := From(app)
	if mgr == nil {
		t.Fatal("social manager missing")
	}
	p, err := mgr.Driver("github")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.(*githubProvider); ok {
		t.Fatal("Register must not treat empty credentials as a live GitHub client")
	}
	if _, ok := p.(*StubProvider); !ok {
		t.Fatalf("expected stub in non-production, got %T", p)
	}
}

func TestDisallowedStubsRejectEmptyCredentials(t *testing.T) {
	SetAllowStubProviders(false)
	t.Cleanup(func() { SetAllowStubProviders(true) })

	p := GitHub(Config{})
	_, err := p.UserFromCode("demo")
	if !errors.Is(err, ErrStubNotAllowedInProduction) {
		t.Fatalf("err=%v", err)
	}
}
