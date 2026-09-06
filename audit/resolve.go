package audit

import "github.com/zatrano/framework/contracts"

// From resolves the audit manager from the application container.
func From(app contracts.App) *Manager {
	return contracts.Resolve[*Manager](app, "audit")
}
