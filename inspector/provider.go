package inspector

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "inspector",
		Key:         "inspector",
		Description: "Request inspector toolbar data",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the inspector addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	app.Container().Instance("inspector", New(200))
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
