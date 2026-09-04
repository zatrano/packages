package docs

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "docs",
		Key:         "docs",
		Description: "Markdown docs repository",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the docs addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	app.Container().Instance("docs", New(app.BasePath("docs")))
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
