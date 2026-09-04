package session

import (
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/env"
	"github.com/zatrano/framework/kernel"
)

func boot(app contracts.App) error {
	sess := NewManager(
		app.BasePath("storage", "framework", "sessions"),
		env.GetInt("SESSION_LIFETIME", 120),
	)
	app.Container().Instance("session", sess)
	if k, ok := app.(*kernel.Application); ok {
		installHTTPBridge(k)
	}
	return nil
}
