package seo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zatrano/framework/v2/kernel/http"
	"github.com/zatrano/framework/v2/kernel/routing"
	"github.com/zatrano/packages/docs"
)

// URL is a sitemap entry.
type URL struct {
	Loc        string
	LastMod    time.Time
	ChangeFreq string
	Priority   float64
}

// Link is a titled path used in LLM discovery lists.
type Link struct {
	Title string
	Path  string
	Note  string
}

// LLMs configures /llms.txt (not packages/ai).
type LLMs struct {
	Intro    string
	Line     string
	KeyPages []Link
	Optional []Link
	About    []string
}

// Plugin configures /.well-known/ai-plugin.json.
type Plugin struct {
	SchemaVersion       string
	NameForHuman        string
	NameForModel        string
	DescriptionForHuman string
	DescriptionForModel string
	AuthType            string
	APIType             string
	APIURL              string
	LogoURL             string
	ContactEmail        string
	LegalInfoURL        string
}

// Security holds RFC 9116 security.txt fields.
type Security struct {
	ContactEmail  string
	ContactURL    string
	Expires       time.Time
	PolicyURL     string
	Canonical     string
	PreferredLang string
}

// Options configures identity and discovery for a Site.
type Options struct {
	BaseURL             string
	Name                string
	Description         string
	Locale              string
	Language            string
	ImagePath           string
	LogoPath            string
	GTMID               string
	SameAs              []string
	DownloadURL         string
	InstallURL          string
	ApplicationCategory string
	OperatingSystem     string
	DocsPrefix          string
	RobotsExtra         []string
	Security            Security
	Humans              string
	Plugin              Plugin
	LLMs                LLMs
	ChangePasswordPath  string
}

// Site is classic crawler SEO plus LLM discovery. Resolve with From(app).
type Site struct {
	mu   sync.Mutex
	opts Options
	urls []URL
	docs *docs.Repository
}

// New creates a site with only a base URL (sitemap-style).
func New(baseURL string) *Site {
	return NewWithOptions(Options{BaseURL: baseURL})
}

// NewWithOptions creates a site with identity and discovery defaults filled in.
func NewWithOptions(opts Options) *Site {
	opts.BaseURL = strings.TrimRight(opts.BaseURL, "/")
	if opts.Name == "" {
		opts.Name = "App"
	}
	if opts.Locale == "" {
		opts.Locale = "en_US"
	}
	if opts.Language == "" {
		opts.Language = "en"
	}
	if opts.ImagePath == "" {
		opts.ImagePath = "/favicon-512.png"
	}
	if opts.LogoPath == "" {
		opts.LogoPath = opts.ImagePath
	}
	if opts.DocsPrefix == "" {
		opts.DocsPrefix = "/docs"
	}
	opts.DocsPrefix = "/" + strings.Trim(opts.DocsPrefix, "/")
	if opts.ApplicationCategory == "" {
		opts.ApplicationCategory = "WebApplication"
	}
	if opts.OperatingSystem == "" {
		opts.OperatingSystem = "Any"
	}
	return &Site{opts: opts, urls: make([]URL, 0)}
}

// Configure mutates options before Register.
func (s *Site) Configure(fn func(*Options)) *Site {
	if s == nil || fn == nil {
		return s
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.opts)
	s.opts.BaseURL = strings.TrimRight(s.opts.BaseURL, "/")
	if s.opts.DocsPrefix != "" {
		s.opts.DocsPrefix = "/" + strings.Trim(s.opts.DocsPrefix, "/")
	}
	return s
}

// Add appends a URL entry (path or absolute).
func (s *Site) Add(loc string, opts ...URL) *Site {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addLocked(loc, opts...)
	return s
}

func (s *Site) addLocked(loc string, opts ...URL) {
	resolved := s.resolve(loc)
	for i := range s.urls {
		if s.urls[i].Loc == resolved {
			if len(opts) > 0 {
				s.urls[i] = s.mergeURL(resolved, loc, opts[0])
			}
			return
		}
	}
	entry := URL{Loc: resolved, Priority: 0.5, ChangeFreq: "weekly"}
	if len(opts) > 0 {
		entry = s.mergeURL(resolved, loc, opts[0])
	} else {
		entry.LastMod = time.Now().UTC()
	}
	s.urls = append(s.urls, entry)
}

func (s *Site) mergeURL(resolved, loc string, o URL) URL {
	if o.LastMod.IsZero() {
		o.LastMod = time.Now().UTC()
	}
	if o.ChangeFreq == "" {
		o.ChangeFreq = "weekly"
	}
	if o.Priority == 0 {
		o.Priority = 0.5
	}
	o.Loc = resolved
	if loc != "" && o.Loc == "" {
		o.Loc = s.resolve(loc)
	}
	return o
}

// XML renders sitemap.xml.
func (s *Site) XML() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, u := range s.urls {
		sb.WriteString("  <url>\n")
		sb.WriteString("    <loc>" + xmlEscape(u.Loc) + "</loc>\n")
		if !u.LastMod.IsZero() {
			sb.WriteString("    <lastmod>" + u.LastMod.Format("2006-01-02") + "</lastmod>\n")
		}
		if u.ChangeFreq != "" {
			sb.WriteString("    <changefreq>" + xmlEscape(u.ChangeFreq) + "</changefreq>\n")
		}
		sb.WriteString(fmt.Sprintf("    <priority>%.1f</priority>\n", u.Priority))
		sb.WriteString("  </url>\n")
	}
	sb.WriteString("</urlset>\n")
	return sb.String()
}

// Robots renders a Lighthouse-valid robots.txt (User-agent / Allow / Disallow / Sitemap / comments).
func (s *Site) Robots(extra ...string) string {
	s.mu.Lock()
	lines := append([]string{}, s.opts.RobotsExtra...)
	s.mu.Unlock()
	lines = append(lines, extra...)
	var sb strings.Builder
	sb.WriteString("User-agent: *\n")
	sb.WriteString("Allow: /\n")
	for _, line := range lines {
		sb.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("Sitemap: " + s.baseURL() + "/sitemap.xml\n")
	return sb.String()
}

// SitemapHandler serves sitemap.xml.
func (s *Site) SitemapHandler() routing.HandlerFunc {
	return func(req *http.Request) *http.Response {
		return http.Text(s.XML()).
			Header("Content-Type", "application/xml; charset=utf-8").
			CacheFor(time.Hour)
	}
}

// RobotsHandler serves robots.txt.
func (s *Site) RobotsHandler() routing.HandlerFunc {
	return func(req *http.Request) *http.Response {
		return http.Text(s.Robots()).
			Header("Content-Type", "text/plain; charset=utf-8").
			CacheFor(time.Hour)
	}
}

// Register mounts classic crawler and LLM discovery routes. Boot does not call this.
func (s *Site) Register(router *routing.Router) {
	if s == nil || router == nil {
		return
	}
	router.Get("/sitemap.xml", s.SitemapHandler()).As("seo.sitemap")
	router.Get("/robots.txt", s.RobotsHandler()).As("seo.robots")
	router.Get("/llms.txt", s.LLMsHandler()).As("seo.llms")
	router.Get("/llm.txt", func(req *http.Request) *http.Response {
		return http.Redirect("/llms.txt", 301)
	}).As("seo.llm")
	router.Get("/llms-full.txt", s.LLMsFullHandler()).As("seo.llms-full")
	s.mu.Lock()
	humans := s.opts.Humans
	changePath := s.opts.ChangePasswordPath
	s.mu.Unlock()
	if humans != "" {
		router.Get("/humans.txt", s.HumansHandler()).As("seo.humans")
	}
	router.Get("/.well-known/security.txt", s.SecurityTxtHandler()).As("seo.security")
	router.Get("/.well-known/ai-plugin.json", s.PluginHandler()).As("seo.ai-plugin")
	if changePath != "" {
		router.Get("/.well-known/change-password", s.ChangePasswordHandler()).As("seo.change-password")
	}
}

func (s *Site) baseURL() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.opts.BaseURL
}

func (s *Site) resolve(loc string) string {
	if strings.HasPrefix(loc, "http://") || strings.HasPrefix(loc, "https://") {
		return loc
	}
	if !strings.HasPrefix(loc, "/") {
		loc = "/" + loc
	}
	return s.opts.BaseURL + loc
}

func xmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}
