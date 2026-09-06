package octane

import "github.com/zatrano/framework/v2/contracts"

// From resolves the Octane runtime from the application container.
func From(app contracts.App) *Runtime {
	return contracts.Resolve[*Runtime](app, "octane")
}
