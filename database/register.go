package database

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "database",
		Key:         "database",
		Description: "Database manager",
		Order:       10,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
	})
}

// ServiceProvider boots the database package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
