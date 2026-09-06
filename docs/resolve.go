package docs

import "github.com/zatrano/framework/contracts"

// From resolves the package service from the application container.
func From(app contracts.App) *Repository {
	return contracts.Resolve[*Repository](app, "docs")
}
