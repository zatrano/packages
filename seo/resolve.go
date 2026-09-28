package seo

import "github.com/zatrano/framework/v3/core/contracts"

// From resolves the SEO site from the application container.
func From(app contracts.App) *Site {
	return contracts.Resolve[*Site](app, "seo")
}
