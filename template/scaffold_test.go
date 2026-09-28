package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zatrano/framework/v3/core/kernel"
	"github.com/zatrano/packages/template"
)

func TestWriteStarterTemplates(t *testing.T) {
	dir := t.TempDir()
	app := kernel.NewApplication(dir)
	if err := template.WriteStarterTemplates(app); err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(dir, "templates", "layouts", "app.html"),
		filepath.Join(dir, "templates", "web", "welcome.html"),
	}
	for _, path := range want {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
	if err := os.WriteFile(want[0], []byte("kept"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := template.WriteStarterTemplates(app); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(want[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "kept" {
		t.Fatal("must not overwrite an existing layout")
	}
}

func TestWriteStarterTemplatesSwitchesStarterHome(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "app", "http", "handlers", "web", "home_handler.go")
	if err := os.MkdirAll(filepath.Dir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package web

import "github.com/zatrano/framework/v3/core/kernel/http"

type HomeHandler struct{}

func (c *HomeHandler) Index(req *http.Request) *http.Response {
	return http.HTML("<!DOCTYPE html><html><head><title>Demo</title></head><body><h1>Demo</h1></body></html>")
}
`
	if err := os.WriteFile(home, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := template.WriteStarterTemplates(kernel.NewApplication(dir)); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(home)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if !strings.Contains(text, `http.Template("web.welcome")`) {
		t.Fatalf("expected Template home, got:\n%s", text)
	}
	if strings.Contains(text, "http.HTML(") {
		t.Fatal("starter http.HTML must be replaced")
	}
}

func TestWriteStarterTemplatesLeavesCustomHome(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "app", "http", "handlers", "web", "home_handler.go")
	if err := os.MkdirAll(filepath.Dir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package web

import "github.com/zatrano/framework/v3/core/kernel/http"

type HomeHandler struct{}

func (c *HomeHandler) Index(req *http.Request) *http.Response {
	return http.Template("web.dashboard")
}
`
	if err := os.WriteFile(home, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := template.WriteStarterTemplates(kernel.NewApplication(dir)); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(home)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `http.Template("web.dashboard")`) {
		t.Fatal("custom home must stay")
	}
}
