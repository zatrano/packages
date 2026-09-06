package health

import "github.com/zatrano/framework/v2/contracts"

// From resolves the health manager from the application container.
func From(app contracts.App) contracts.Health {
	if app == nil {
		return nil
	}
	raw, err := app.Make("health")
	if err != nil || raw == nil {
		return nil
	}
	h, _ := raw.(contracts.Health)
	return h
}
