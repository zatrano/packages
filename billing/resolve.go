package billing

import "github.com/zatrano/framework/contracts"

// From resolves the billing manager from the application container.
func From(app contracts.App) *Manager {
	return contracts.Resolve[*Manager](app, "billing")
}
