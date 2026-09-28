package template

import (
	"github.com/zatrano/framework/v3/core/bootstrap/addons"
	"github.com/zatrano/framework/v3/core/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "template",
		Key:         "template",
		Description: "Canvas HTML template engine",
		Order:       129,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
		Scaffold:    WriteStarterTemplates,
	})
}

// ServiceProvider boots the Canvas template package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
