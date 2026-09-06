package octane

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/env"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "octane",
		Key:         "octane",
		Description: "Concurrent runtime metrics",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
	})
}

// ServiceProvider boots the octane addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	app.Container().Instance("octane", New(env.GetInt("OCTANE_WORKERS", 0)))
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
