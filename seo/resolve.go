package seo

import "github.com/zatrano/framework/v2/contracts"

// From resolves the SEO site from the application container.
func From(app contracts.App) *Site {
	return contracts.Resolve[*Site](app, "seo")
}
