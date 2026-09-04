package flash

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "flash",
		Key:         "flash",
		Description: "Flash / toast messages",
		Order:       131,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the flash package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
