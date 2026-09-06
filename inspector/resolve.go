package inspector

import "github.com/zatrano/framework/v2/contracts"

// From resolves the package service from the application container.
func From(app contracts.App) *Manager {
	return contracts.Resolve[*Manager](app, "inspector")
}
