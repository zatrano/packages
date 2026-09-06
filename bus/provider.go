package bus

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "bus",
		Key:         "bus",
		Description: "Synchronous command bus",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the bus addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	app.Container().Instance("bus", New())
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
