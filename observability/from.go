package observability

import "github.com/zatrano/framework/v2/contracts"

// From resolves the metrics collector from the application container.
func From(app contracts.App) *Metrics {
	if app == nil {
		return nil
	}
	raw, err := app.Make("metrics")
	if err != nil || raw == nil {
		return nil
	}
	m, _ := raw.(*Metrics)
	return m
}
