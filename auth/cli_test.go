package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/kernel"
)

func TestAuthCLIRegistered(t *testing.T) {
	meta, ok := addons.Lookup("auth")
	if !ok || meta.CLI == nil {
		t.Fatal("auth Meta.CLI is not registered")
	}
	cmds := meta.CLI(kernel.NewApplication(t.TempDir()))
	want := map[string]bool{"make:auth": false, "make:panel": false}
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
	want := []string{
		filepath.Join(dir, "app", "views", "layout", "auth.html"),
		filepath.Join(dir, "app", "views", "layout", "mail.html"),
		filepath.Join(dir, "app", "views", "auth", "login.html"),
		filepath.Join(dir, "app", "views", "mail", "auth", "password-reset.html"),
		filepath.Join(dir, "app", "views", "mail", "auth", "verify-email.html"),
		filepath.Join(dir, "app", "views", "mail", "auth", "password-changed.html"),
	}
	for _, path := range want {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected embedded stub at %s: %v", path, err)
		}
	}
}

func TestMakeAuthWritesNestedSurfaces(t *testing.T) {
	dir := t.TempDir()
	cmd := &MakeAuthCommand{app: kernel.NewApplication(dir)}
	if err := cmd.Handle(nil); err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(dir, "app", "http", "controllers", "auth", "web", "auth_controller.go"),
		filepath.Join(dir, "app", "http", "controllers", "auth", "api", "auth_controller.go"),
		filepath.Join(dir, "app", "routes", "auth", "web", "auth.go"),
		filepath.Join(dir, "app", "routes", "auth", "api", "auth.go"),
	}
	for _, path := range want {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
	absent := []string{
		filepath.Join(dir, "app", "http", "controllers", "auth", "web", "social_auth_controller.go"),
		filepath.Join(dir, "app", "http", "controllers", "auth", "api", "social_auth_controller.go"),
		filepath.Join(dir, "app", "services", "social.go"),
		filepath.Join(dir, "app", "models", "social_account.go"),
	}
	for _, path := range absent {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("social must stay off unless --social is passed: %s", path)
		}
	}
	checks := []struct {
		rel     string
		needles []string
	}{
		{filepath.Join("app", "views", "auth", "login.html"), []string{"auth-social", "google/login", "continue_google"}},
		{filepath.Join("app", "views", "auth", "register.html"), []string{"auth-social", "google/login", "continue_google"}},
		{filepath.Join("app", "views", "layout", "auth.html"), []string{"auth-social"}},
		{filepath.Join("app", "routes", "auth", "web", "auth.go"), []string{"SocialAuthController", "google/login"}},
		{filepath.Join("app", "routes", "auth", "api", "auth.go"), []string{"SocialAuthController", "/google"}},
		{filepath.Join("app", "localization", "en", "auth.json"), []string{"continue_google", "provider_google", "social_"}},
	}
	for _, check := range checks {
		body, err := os.ReadFile(filepath.Join(dir, check.rel))
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		for _, needle := range check.needles {
			if strings.Contains(text, needle) {
				t.Fatalf("%s must not mention %q without --social", check.rel, needle)
			}
		}
	}
}

func TestMakeAuthSocialBareFlagEnablesGoogle(t *testing.T) {
	dir := t.TempDir()
	cmd := &MakeAuthCommand{app: kernel.NewApplication(dir)}
	if err := cmd.Handle([]string{"--social"}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(dir, "app", "http", "controllers", "auth", "web", "social_auth_controller.go"),
		filepath.Join(dir, "app", "http", "controllers", "auth", "api", "social_auth_controller.go"),
		filepath.Join(dir, "app", "services", "social.go"),
	}
	for _, path := range want {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
	login, err := os.ReadFile(filepath.Join(dir, "app", "views", "auth", "login.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(login), "/auth/google/login") {
		t.Fatal("expected Google login link after --social")
	}
	webRoutes, err := os.ReadFile(filepath.Join(dir, "app", "routes", "auth", "web", "auth.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(webRoutes), "SocialAuthController") {
		t.Fatal("expected social routes after --social")
	}
}

func TestMakeAuthSocialRejectsGitHub(t *testing.T) {
	cmd := &MakeAuthCommand{app: kernel.NewApplication(t.TempDir())}
	if err := cmd.Handle([]string{"--social=github"}); err == nil {
		t.Fatal("expected --social=github to fail")
	}
}

func TestMakeAuthEnablesViewAndSwitchesHome(t *testing.T) {
	dir := t.TempDir()
	writeAuthEnablementFixtures(t, dir)
	cmd := &MakeAuthCommand{app: kernel.NewApplication(dir)}
	if err := cmd.Handle(nil); err != nil {
		t.Fatal(err)
	}
	enabled, err := os.ReadFile(filepath.Join(dir, "bootstrap", "enabled.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(enabled), `"view"`) {
		t.Fatal("make:auth must enable view")
	}
	addonSrc, err := os.ReadFile(filepath.Join(dir, "bootstrap", "addons.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(addonSrc), "github.com/zatrano/packages/view") {
		t.Fatal("make:auth must blank-import view")
	}
	home, err := os.ReadFile(filepath.Join(dir, "app", "http", "controllers", "web", "home_controller.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(home), `http.View("web.welcome")`) {
		t.Fatalf("starter home must switch to View:\n%s", home)
	}
	if _, err := os.Stat(filepath.Join(dir, "app", "views", "web", "welcome.html")); err != nil {
		t.Fatal(err)
	}
}

func writeAuthEnablementFixtures(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "bootstrap"), 0o755); err != nil {
		t.Fatal(err)
	}
	enabled := `package bootstrap

var EnabledAddons = []string{
	"health",
	"auth",
}
`
	if err := os.WriteFile(filepath.Join(dir, "bootstrap", "enabled.go"), []byte(enabled), 0o644); err != nil {
		t.Fatal(err)
	}
	addons := `package bootstrap

import (
	_ "github.com/zatrano/packages/health"
	_ "github.com/zatrano/packages/auth"
)
`
	if err := os.WriteFile(filepath.Join(dir, "bootstrap", "addons.go"), []byte(addons), 0o644); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(dir, "app", "http", "controllers", "web", "home_controller.go")
	if err := os.MkdirAll(filepath.Dir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package web

import "github.com/zatrano/framework/v2/kernel/http"

type HomeController struct{}

func (c *HomeController) Index(req *http.Request) *http.Response {
	return http.HTML("<!DOCTYPE html><html><head><title>App</title></head><body><h1>App</h1></body></html>")
}
`
	if err := os.WriteFile(home, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMakePanelWritesSurface(t *testing.T) {
	dir := t.TempDir()
	cmd := &MakePanelCommand{app: kernel.NewApplication(dir)}
	if err := cmd.Handle([]string{"dashboard"}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(dir, "app", "http", "controllers", "dashboard", "home_controller.go"),
		filepath.Join(dir, "app", "routes", "dashboard", "dashboard.go"),
		filepath.Join(dir, "app", "views", "dashboard", "layouts", "dashboard.html"),
		filepath.Join(dir, "app", "views", "dashboard", "index.html"),
	}
	for _, path := range want {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
}

func TestMakePanelRejectsReservedName(t *testing.T) {
	cmd := &MakePanelCommand{app: kernel.NewApplication(t.TempDir())}
	if err := cmd.Handle([]string{"auth"}); err == nil {
		t.Fatal("expected reserved surface name to fail")
	}
}
