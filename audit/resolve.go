package audit

import "github.com/zatrano/framework/v3/core/contracts"

// From resolves the audit manager from the application container.
func From(app contracts.App) *Manager {
	return contracts.Resolve[*Manager](app, "audit")
}
