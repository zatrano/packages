package validation

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "validation",
		Key:         "validation",
		Description: "Input validation",
		Order:       132,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the validation package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
