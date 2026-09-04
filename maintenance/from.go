package maintenance

import "github.com/zatrano/framework/contracts"

// From resolves the maintenance manager from the application container.
func From(app contracts.App) contracts.Maintenance {
	if app == nil {
		return nil
	}
	raw, err := app.Make("maintenance")
	if err != nil || raw == nil {
		return nil
	}
	m, _ := raw.(contracts.Maintenance)
	return m
}
