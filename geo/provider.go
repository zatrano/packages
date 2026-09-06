package geo

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "geo",
		Key:         "geo",
		Description: "Geolocation resolver",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the geo addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	app.Container().Instance("geo", New())
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
