package seo_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	stdhttp "net/http"

	"github.com/zatrano/framework/v2/kernel/http"
	"github.com/zatrano/framework/v2/kernel/routing"
	"github.com/zatrano/packages/docs"
	"github.com/zatrano/packages/seo"
)

func TestSitemapXMLAndRobots(t *testing.T) {
	s := seo.New("http://localhost:8080")
	s.Add("/").Add("/documentation")
	xml := s.XML()
	if !strings.Contains(xml, "<loc>http://localhost:8080/</loc>") {
		t.Fatal(xml)
	}
	if !strings.Contains(xml, "<loc>http://localhost:8080/documentation</loc>") {
		t.Fatal(xml)
	}
	robots := s.Robots()
	if !strings.Contains(robots, "Sitemap: http://localhost:8080/sitemap.xml") {
		t.Fatal(robots)
	}
	if strings.Contains(robots, "LLMs-Txt") {
		t.Fatal("robots.txt must stay Lighthouse-valid")
	}
}

func TestSecurityTxt(t *testing.T) {
	s := seo.NewWithOptions(seo.Options{
		BaseURL: "https://example.test",
		Security: seo.Security{
			ContactEmail: "security@example.test",
			Expires:      time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
			Canonical:    "https://example.test/.well-known/security.txt",
		},
	})
	txt := s.SecurityTxt()
	if !strings.Contains(txt, "Contact: mailto:security@example.test") {
		t.Fatal(txt)
	}
	if !strings.Contains(txt, "Expires:") {
		t.Fatal(txt)
	}
}

func TestViewDataIsGeneric(t *testing.T) {
	s := seo.NewWithOptions(seo.Options{
		BaseURL:     "https://example.test",
		Name:        "Acme",
		Description: "Acme docs",
		SameAs:      []string{"https://github.com/acme"},
	})
	data := s.ViewData(seo.Meta{Title: "Home", Path: "/"})
	if data["ogTitle"] != "Home" {
		t.Fatalf("%#v", data)
	}
	jsonLd, _ := data["jsonLd"].(string)
	if strings.Contains(jsonLd, "ZATRANO") {
		t.Fatal("package JSON-LD must not hardcode ZATRANO")
	}
	if !strings.Contains(jsonLd, "Acme") {
		t.Fatal(jsonLd)
	}
}

func TestLLMsTxtAndPlugin(t *testing.T) {
	s := seo.NewWithOptions(seo.Options{
		BaseURL:     "https://example.test",
		Name:        "Acme",
		Description: "Acme platform",
	})
	s.Configure(func(o *seo.Options) {
		o.LLMs = seo.LLMs{
			Intro: "Services resolve with From(app).",
			Line:  "v1.0.0",
			KeyPages: []seo.Link{
				{Title: "Home", Path: "/", Note: "Overview."},
			},
		}
	})
	txt := s.LLMsTxt()
	if !strings.Contains(txt, "From(app)") || !strings.Contains(txt, "# Acme") {
		t.Fatal(txt)
	}
	plugin := s.PluginJSON()
	if plugin["name_for_human"] != "Acme" {
		t.Fatalf("%#v", plugin)
	}
}

func TestRegisterRoutes(t *testing.T) {
	s := seo.NewWithOptions(seo.Options{
		BaseURL: "http://localhost:8080",
		Name:    "Acme",
		Humans:  "# TEAM\n",
	})
	s.Add("/docs")
	router := routing.New()
	s.Register(router)

	raw, _ := stdhttp.NewRequest(stdhttp.MethodGet, "/sitemap.xml", nil)
	resp := router.Dispatch(http.NewRequest(raw))
	if resp == nil || resp.StatusCode() != 200 || !strings.Contains(string(resp.Content()), "/docs") {
		t.Fatalf("sitemap %#v body=%s", resp, respBody(resp))
	}

	raw, _ = stdhttp.NewRequest(stdhttp.MethodGet, "/llms.txt", nil)
	resp = router.Dispatch(http.NewRequest(raw))
	if resp == nil || resp.StatusCode() != 200 {
		t.Fatalf("llms %#v", resp)
	}

	raw, _ = stdhttp.NewRequest(stdhttp.MethodGet, "/llm.txt", nil)
	resp = router.Dispatch(http.NewRequest(raw))
	if resp == nil || resp.StatusCode() != 301 {
		t.Fatalf("llm redirect %#v", resp)
	}

	raw, _ = stdhttp.NewRequest(stdhttp.MethodGet, "/humans.txt", nil)
	resp = router.Dispatch(http.NewRequest(raw))
	if resp == nil || !strings.Contains(string(resp.Content()), "TEAM") {
		t.Fatalf("humans %#v", resp)
	}
}

func TestAttachDocsFillsSitemapAndLLMsFull(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "installation.md"), []byte("# Installation\n\nGo get"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "navigation.json"), []byte(`{
  "sections": [{"title": "Start", "pages": [{"title": "Installation", "slug": "installation"}]}]
}`), 0o644)
	repo := docs.New(dir)
	s := seo.NewWithOptions(seo.Options{BaseURL: "https://example.test", Name: "Acme"})
	s.AttachDocs(repo)
	xml := s.XML()
	if !strings.Contains(xml, "/docs/installation") {
		t.Fatal(xml)
	}
	full := s.LLMsFullTxt()
	if !strings.Contains(full, "Installation") {
		t.Fatal(full)
	}
}

func TestSEODoesNotImportAI(t *testing.T) {
	out, err := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", "github.com/zatrano/packages/seo").CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v", out, err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		switch line {
		case "github.com/zatrano/packages/ai",
			"github.com/zatrano/packages/rag",
			"github.com/zatrano/packages/agent",
			"github.com/zatrano/packages/workflow":
			t.Fatalf("forbidden import %s", line)
		}
	}
}

func respBody(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	return string(resp.Content())
}
