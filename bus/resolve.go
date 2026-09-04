package bus

import "github.com/zatrano/framework/contracts"

// From resolves the package service from the application container.
func From(app contracts.App) *Bus {
	return contracts.Resolve[*Bus](app, "bus")
}
