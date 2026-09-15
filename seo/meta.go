package seo

import (
	"encoding/json"
	"strings"
)

// Meta holds Open Graph / Twitter / canonical fields for a page.
type Meta struct {
	Title       string
	Description string
	Path        string
	OGType      string
	Robots      string
	ImagePath   string
}

// ViewData builds template variables for SEO partials (canonical, OG, JSON-LD, GTM).
func (s *Site) ViewData(meta Meta) map[string]any {
	s.mu.Lock()
	opts := s.opts
	s.mu.Unlock()

	appURL := strings.TrimRight(opts.BaseURL, "/")
	if meta.Path == "" {
		meta.Path = "/"
	}
	if !strings.HasPrefix(meta.Path, "/") {
		meta.Path = "/" + meta.Path
	}
	if meta.OGType == "" {
		meta.OGType = "website"
	}
	if meta.Robots == "" {
		meta.Robots = "index,follow,max-image-preview:large,max-snippet:-1,max-video-preview:-1"
	}
	if meta.ImagePath == "" {
		meta.ImagePath = opts.ImagePath
	}
	if meta.ImagePath != "" && !strings.HasPrefix(meta.ImagePath, "http") {
		meta.ImagePath = appURL + meta.ImagePath
	}
	if meta.Title == "" {
		meta.Title = opts.Name
	}
	if meta.Description == "" {
		meta.Description = opts.Description
	}

	canonical := appURL + meta.Path
	if meta.Path == "/" {
		canonical = appURL + "/"
	}

	return map[string]any{
		"appUrl":        appURL,
		"gtmId":         opts.GTMID,
		"canonicalUrl":  canonical,
		"ogTitle":       meta.Title,
		"ogDescription": meta.Description,
		"ogType":        meta.OGType,
		"ogImage":       meta.ImagePath,
		"robotsMeta":    meta.Robots,
		"jsonLd":        mustJSON(s.graph(opts, meta, canonical)),
	}
}

func (s *Site) graph(opts Options, meta Meta, canonical string) map[string]any {
	appURL := strings.TrimRight(opts.BaseURL, "/")
	logo := opts.LogoPath
	if logo != "" && !strings.HasPrefix(logo, "http") {
		logo = appURL + logo
	}
	org := map[string]any{
		"@type": "Organization",
		"@id":   appURL + "/#organization",
		"name":  opts.Name,
		"url":   appURL + "/",
	}
	if logo != "" {
		org["logo"] = map[string]any{
			"@type": "ImageObject",
			"url":   logo,
		}
	}
	if len(opts.SameAs) > 0 {
		org["sameAs"] = append([]string{}, opts.SameAs...)
	}

	website := map[string]any{
		"@type":      "WebSite",
		"@id":        appURL + "/#website",
		"url":        appURL + "/",
		"name":       opts.Name,
		"publisher":  map[string]any{"@id": appURL + "/#organization"},
		"inLanguage": opts.Language,
	}
	if opts.Description != "" {
		website["description"] = opts.Description
	}

	nodes := []any{org, website}
	if opts.DownloadURL != "" || opts.InstallURL != "" {
		software := map[string]any{
			"@type":               "SoftwareApplication",
			"@id":                 appURL + "/#software",
			"name":                opts.Name,
			"applicationCategory": opts.ApplicationCategory,
			"operatingSystem":     opts.OperatingSystem,
			"url":                 appURL + "/",
			"author":              map[string]any{"@id": appURL + "/#organization"},
		}
		if opts.DownloadURL != "" {
			software["downloadUrl"] = opts.DownloadURL
		}
		if opts.InstallURL != "" {
			software["installUrl"] = opts.InstallURL
		}
		if opts.Description != "" {
			software["description"] = opts.Description
		}
		nodes = append(nodes, software)
	}

	webpage := map[string]any{
		"@type":       "WebPage",
		"@id":         canonical + "#webpage",
		"url":         canonical,
		"name":        meta.Title,
		"description": meta.Description,
		"isPartOf":    map[string]any{"@id": appURL + "/#website"},
		"inLanguage":  opts.Language,
	}
	if strings.HasPrefix(meta.Path, opts.DocsPrefix) {
		webpage["@type"] = []string{"WebPage", "TechArticle"}
		if opts.DownloadURL != "" || opts.InstallURL != "" {
			webpage["about"] = map[string]any{"@id": appURL + "/#software"}
		}
	}
	nodes = append(nodes, webpage)

	return map[string]any{
		"@context": "https://schema.org",
		"@graph":   nodes,
	}
}

func mustJSON(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
