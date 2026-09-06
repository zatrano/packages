package features

import "github.com/zatrano/framework/v2/contracts"

// From resolves the feature manager from the application container.
func From(app contracts.App) *Manager {
	return contracts.Resolve[*Manager](app, "features")
}
