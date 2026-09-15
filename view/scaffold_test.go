package view_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zatrano/framework/v2/kernel"
	"github.com/zatrano/packages/view"
)

func TestWriteStarterViews(t *testing.T) {
	dir := t.TempDir()
	app := kernel.NewApplication(dir)
	if err := view.WriteStarterViews(app); err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(dir, "app", "views", "layout", "app.html"),
		filepath.Join(dir, "app", "views", "web", "welcome.html"),
	}
	for _, path := range want {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
	if err := os.WriteFile(want[0], []byte("kept"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := view.WriteStarterViews(app); err != nil {
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

func TestWriteStarterViewsSwitchesStarterHome(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "app", "http", "controllers", "web", "home_controller.go")
	if err := os.MkdirAll(filepath.Dir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package web

import "github.com/zatrano/framework/v2/kernel/http"

type HomeController struct{}

func (c *HomeController) Index(req *http.Request) *http.Response {
	return http.HTML("<!DOCTYPE html><html><head><title>Demo</title></head><body><h1>Demo</h1></body></html>")
}
`
	if err := os.WriteFile(home, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := view.WriteStarterViews(kernel.NewApplication(dir)); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(home)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if !strings.Contains(text, `http.View("web.welcome")`) {
		t.Fatalf("expected View home, got:\n%s", text)
	}
	if strings.Contains(text, "http.HTML(") {
		t.Fatal("starter http.HTML must be replaced")
	}
}

func TestWriteStarterViewsLeavesCustomHome(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "app", "http", "controllers", "web", "home_controller.go")
	if err := os.MkdirAll(filepath.Dir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package web

import "github.com/zatrano/framework/v2/kernel/http"

type HomeController struct{}

func (c *HomeController) Index(req *http.Request) *http.Response {
	return http.View("web.dashboard")
}
`
	if err := os.WriteFile(home, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := view.WriteStarterViews(kernel.NewApplication(dir)); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(home)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `http.View("web.dashboard")`) {
		t.Fatal("custom home must stay")
	}
}
