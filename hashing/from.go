package hashing

import "github.com/zatrano/framework/contracts"

// From resolves the hasher from the application container.
func From(app contracts.App) *Manager {
	if app == nil {
		return nil
	}
	raw, err := app.Make("hash")
	if err != nil || raw == nil {
		return nil
	}
	m, _ := raw.(*Manager)
	return m
}
