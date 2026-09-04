package tenancy

import "github.com/zatrano/framework/contracts"

// From resolves the package service from the application container.
func From(app contracts.App) *Manager {
	return contracts.Resolve[*Manager](app, "tenancy")
}
