package apitoken

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "apitoken",
		Key:         "apitoken",
		Description: "Personal access tokens",
		Order:       52,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the apitoken package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
