package hashing

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "hashing",
		Key:         "hash",
		Description: "Password hashing",
		Order:       15,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	app.Container().Instance("hash", New())
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
