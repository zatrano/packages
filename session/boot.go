package session

import (
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/kernel/config"
	"github.com/zatrano/framework/kernel/env"
)

func boot(app contracts.App) error {
	config.LoadIfAbsent(app.Config(), "session", DefaultConfig())
	sess := NewManager(
		app.BasePath("storage", "framework", "sessions"),
		env.GetInt("SESSION_LIFETIME", 120),
	)
	app.Container().Instance("session", sess)
	installHTTPBridge(app)
	return nil
}
