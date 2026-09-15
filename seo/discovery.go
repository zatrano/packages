package seo

import (
	"fmt"
	"strings"
	"time"

	"github.com/zatrano/framework/v2/kernel/http"
	"github.com/zatrano/framework/v2/kernel/routing"
)

// LLMsTxt renders /llms.txt for LLM crawlers.
func (s *Site) LLMsTxt() string {
	s.mu.Lock()
	opts := s.opts
	s.mu.Unlock()
	base := strings.TrimRight(opts.BaseURL, "/")
	var b strings.Builder
	b.WriteString("# " + opts.Name + "\n\n")
	intro := strings.TrimSpace(opts.LLMs.Intro)
	if intro == "" {
		intro = opts.Description
	}
	if intro != "" {
		b.WriteString("> " + intro + "\n\n")
	}
	if len(opts.LLMs.KeyPages) > 0 {
		b.WriteString("## Key pages\n\n")
		for _, page := range opts.LLMs.KeyPages {
			b.WriteString(formatLink(base, page))
		}
		b.WriteString("\n")
	}
	b.WriteString("## About\n\n")
	b.WriteString("- Name: " + opts.Name + "\n")
	if base != "" {
		b.WriteString(fmt.Sprintf("- Site: %s\n", base))
		b.WriteString(fmt.Sprintf("- Sitemap: %s/sitemap.xml\n", base))
	}
	if opts.LLMs.Line != "" {
		b.WriteString("- Line: " + opts.LLMs.Line + "\n")
	}
	if opts.Language != "" {
		b.WriteString("- Default language: " + opts.Language + "\n")
	}
	for _, line := range opts.LLMs.About {
		if strings.HasPrefix(strings.TrimSpace(line), "- ") {
			b.WriteString(line)
			if !strings.HasSuffix(line, "\n") {
				b.WriteString("\n")
			}
			continue
		}
		b.WriteString("- " + line + "\n")
	}
	if len(opts.LLMs.Optional) > 0 {
		b.WriteString("\n## Optional\n\n")
		for _, page := range opts.LLMs.Optional {
			b.WriteString(formatLink(base, page))
		}
	}
	return b.String()
}

// LLMsFullTxt renders /llms-full.txt. Uses docs navigation when attached.
func (s *Site) LLMsFullTxt() string {
	s.mu.Lock()
	opts := s.opts
	repo := s.docs
	s.mu.Unlock()
	base := strings.TrimRight(opts.BaseURL, "/")
	var b strings.Builder
	b.WriteString("# " + opts.Name + " — full documentation inventory\n\n")
	b.WriteString("> Complete list of public documentation URLs for crawlers and AI agents.\n\n")
	if base != "" {
		b.WriteString(fmt.Sprintf("- [Home](%s/)\n", base))
		b.WriteString(fmt.Sprintf("- [Docs index](%s%s)\n", base, opts.DocsPrefix))
	}

	if repo != nil {
		nav, err := repo.Navigation()
		if err == nil && len(nav) > 0 {
			for _, section := range nav {
				b.WriteString("\n## " + section.Title + "\n\n")
				for _, page := range section.Pages {
					href := docsHref(base, opts.DocsPrefix, page.Slug)
					b.WriteString(fmt.Sprintf("- [%s](%s)\n", page.Title, href))
				}
			}
		} else {
			pages, listErr := repo.List()
			if listErr == nil {
				b.WriteString("\n## Documentation\n\n")
				for _, page := range pages {
					title := page.Title
					if full, getErr := repo.Get(page.Slug); getErr == nil {
						title = full.Title
					}
					b.WriteString(fmt.Sprintf("- [%s](%s)\n", title, docsHref(base, opts.DocsPrefix, page.Slug)))
				}
			}
		}
	}

	b.WriteString("\n## Discovery\n\n")
	if base != "" {
		b.WriteString(fmt.Sprintf("- [llms.txt](%s/llms.txt)\n", base))
		b.WriteString(fmt.Sprintf("- [sitemap.xml](%s/sitemap.xml)\n", base))
	}
	return b.String()
}

// PluginJSON renders the AI plugin discovery document.
func (s *Site) PluginJSON() map[string]any {
	s.mu.Lock()
	opts := s.opts
	s.mu.Unlock()
	base := strings.TrimRight(opts.BaseURL, "/")
	p := opts.Plugin
	if p.SchemaVersion == "" {
		p.SchemaVersion = "v1"
	}
	if p.NameForHuman == "" {
		p.NameForHuman = opts.Name
	}
	if p.NameForModel == "" {
		p.NameForModel = slugName(opts.Name)
	}
	if p.DescriptionForHuman == "" {
		p.DescriptionForHuman = opts.Description
	}
	if p.DescriptionForModel == "" {
		p.DescriptionForModel = opts.Description
	}
	if p.AuthType == "" {
		p.AuthType = "none"
	}
	if p.APIType == "" {
		p.APIType = "openapi"
	}
	if p.APIURL == "" && base != "" {
		p.APIURL = base + opts.DocsPrefix + "/search-index.json"
	}
	if p.LogoURL == "" && base != "" {
		logo := opts.LogoPath
		if logo == "" {
			logo = opts.ImagePath
		}
		if logo != "" && !strings.HasPrefix(logo, "http") {
			logo = base + logo
		}
		p.LogoURL = logo
	}
	return map[string]any{
		"schema_version":        p.SchemaVersion,
		"name_for_human":        p.NameForHuman,
		"name_for_model":        p.NameForModel,
		"description_for_human": p.DescriptionForHuman,
		"description_for_model": p.DescriptionForModel,
		"auth":                  map[string]any{"type": p.AuthType},
		"api":                   map[string]any{"type": p.APIType, "url": p.APIURL},
		"logo_url":              p.LogoURL,
		"contact_email":         p.ContactEmail,
		"legal_info_url":        p.LegalInfoURL,
	}
}

// SecurityTxt renders RFC 9116 security.txt.
func (s *Site) SecurityTxt() string {
	s.mu.Lock()
	cfg := s.opts.Security
	s.mu.Unlock()
	if cfg.Expires.IsZero() {
		cfg.Expires = time.Now().UTC().AddDate(1, 0, 0)
	}
	var b strings.Builder
	if cfg.ContactEmail != "" {
		b.WriteString("Contact: mailto:" + cfg.ContactEmail + "\n")
	}
	if cfg.ContactURL != "" {
		b.WriteString("Contact: " + cfg.ContactURL + "\n")
	}
	b.WriteString("Expires: " + cfg.Expires.UTC().Format(time.RFC3339) + "\n")
	if cfg.PreferredLang != "" {
		b.WriteString("Preferred-Languages: " + cfg.PreferredLang + "\n")
	}
	if cfg.Canonical != "" {
		b.WriteString("Canonical: " + cfg.Canonical + "\n")
	}
	if cfg.PolicyURL != "" {
		b.WriteString("Policy: " + cfg.PolicyURL + "\n")
	}
	return b.String()
}

// HumansTxt returns the optional humans.txt body.
func (s *Site) HumansTxt() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.opts.Humans
}

// LLMsHandler serves /llms.txt.
func (s *Site) LLMsHandler() routing.HandlerFunc {
	return func(req *http.Request) *http.Response {
		return http.Text(s.LLMsTxt()).
			Header("Content-Type", "text/plain; charset=utf-8").
			CacheFor(time.Hour)
	}
}

// LLMsFullHandler serves /llms-full.txt.
func (s *Site) LLMsFullHandler() routing.HandlerFunc {
	return func(req *http.Request) *http.Response {
		return http.Text(s.LLMsFullTxt()).
			Header("Content-Type", "text/plain; charset=utf-8").
			CacheFor(time.Hour)
	}
}

// PluginHandler serves /.well-known/ai-plugin.json.
func (s *Site) PluginHandler() routing.HandlerFunc {
	return func(req *http.Request) *http.Response {
		return http.JSON(s.PluginJSON()).CacheFor(time.Hour)
	}
}

// SecurityTxtHandler serves /.well-known/security.txt.
func (s *Site) SecurityTxtHandler() routing.HandlerFunc {
	return func(req *http.Request) *http.Response {
		return http.Text(s.SecurityTxt()).
			Header("Content-Type", "text/plain; charset=utf-8").
			CacheFor(24 * time.Hour)
	}
}

// HumansHandler serves /humans.txt.
func (s *Site) HumansHandler() routing.HandlerFunc {
	return func(req *http.Request) *http.Response {
		return http.Text(s.HumansTxt()).
			Header("Content-Type", "text/plain; charset=utf-8").
			CacheFor(24 * time.Hour)
	}
}

// ChangePasswordHandler redirects /.well-known/change-password.
func (s *Site) ChangePasswordHandler() routing.HandlerFunc {
	s.mu.Lock()
	loginPath := s.opts.ChangePasswordPath
	s.mu.Unlock()
	if loginPath == "" {
		loginPath = "/auth/login"
	}
	return func(req *http.Request) *http.Response {
		return http.Redirect(loginPath, 302)
	}
}

func formatLink(base string, page Link) string {
	href := page.Path
	if href != "" && !strings.HasPrefix(href, "http") {
		if !strings.HasPrefix(href, "/") {
			href = "/" + href
		}
		href = base + href
	}
	line := fmt.Sprintf("- [%s](%s)", page.Title, href)
	if page.Note != "" {
		line += ": " + page.Note
	}
	return line + "\n"
}

func docsHref(base, prefix, slug string) string {
	if slug == "" || slug == "index" {
		return base + prefix
	}
	return base + prefix + "/" + slug
}

func slugName(name string) string {
	out := strings.ToLower(strings.TrimSpace(name))
	out = strings.ReplaceAll(out, " ", "_")
	return out
}
