package authorization

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "authorization",
		Key:         "authorization",
		Description: "Gates and policies",
		Order:       51,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the authorization package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
