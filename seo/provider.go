package seo

import (
	"strings"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/env"
	"github.com/zatrano/packages/docs"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "seo",
		Key:         "seo",
		Description: "Classic crawler SEO and LLM discovery (sitemap, robots, OG, llms.txt)",
		Optional:    []string{"docs"},
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the seo addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	site := NewWithOptions(optionsFromApp(app))
	site.Add("/", URL{Priority: 1.0, ChangeFreq: "daily"})
	app.Container().Instance("seo", site)
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error {
	site := From(app)
	if site == nil {
		return nil
	}
	if repo := docs.From(app); repo != nil {
		site.AttachDocs(repo)
	}
	return nil
}

func optionsFromApp(app contracts.App) Options {
	base := strings.TrimRight(app.Config().GetString("app.url", env.Get("APP_URL", "http://localhost:8080")), "/")
	name := app.Config().GetString("app.name", env.Get("APP_NAME", "App"))
	locale := env.Get("APP_LOCALE", "en")
	return Options{
		BaseURL:     base,
		Name:        name,
		Description: env.Get("SEO_DESCRIPTION", ""),
		Locale:      env.Get("SEO_LOCALE", "en_US"),
		Language:    locale,
		ImagePath:   env.Get("SEO_IMAGE", "/favicon-512.png"),
		GTMID:       env.Get("GTM_CONTAINER_ID", ""),
		DocsPrefix:  env.Get("SEO_DOCS_PREFIX", "/docs"),
		Security: Security{
			ContactEmail:  env.Get("SECURITY_CONTACT_EMAIL", ""),
			ContactURL:    env.Get("SECURITY_CONTACT_URL", ""),
			PolicyURL:     env.Get("SECURITY_POLICY_URL", ""),
			Canonical:     env.Get("SECURITY_CANONICAL", base+"/.well-known/security.txt"),
			PreferredLang: locale,
		},
	}
}
