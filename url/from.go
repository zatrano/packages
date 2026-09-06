package url

import "github.com/zatrano/framework/contracts"

// From resolves the URL generator from the application container.
func From(app contracts.App) *Generator {
	if app == nil {
		return nil
	}
	raw, err := app.Make("url")
	if err != nil || raw == nil {
		return nil
	}
	g, _ := raw.(*Generator)
	return g
}
