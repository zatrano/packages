package sitemap

import "github.com/zatrano/framework/v2/contracts"

// From resolves the package service from the application container.
func From(app contracts.App) *Builder {
	return contracts.Resolve[*Builder](app, "sitemap")
}
