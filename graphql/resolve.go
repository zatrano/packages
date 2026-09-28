package graphql

import "github.com/zatrano/framework/v3/core/contracts"

// From resolves the package service from the application container.
func From(app contracts.App) *Schema {
	return contracts.Resolve[*Schema](app, "graphql")
}
