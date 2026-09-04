package pulse

import "github.com/zatrano/framework/contracts"

// From resolves the package service from the application container.
func From(app contracts.App) *Dashboard {
	return contracts.Resolve[*Dashboard](app, "pulse")
}
