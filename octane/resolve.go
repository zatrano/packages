package octane

import "github.com/zatrano/framework/contracts"

// From resolves the Octane runtime from the application container.
func From(app contracts.App) *Runtime {
	return contracts.Resolve[*Runtime](app, "octane")
}
