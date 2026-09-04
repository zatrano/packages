package health

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "health",
		Key:         "health",
		Description: "Health checks",
		Order:       16,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

type ServiceProvider struct{}

type contractHealth struct{ *Manager }

func (c contractHealth) Handler() any { return c.Manager.Handler() }

func (p *ServiceProvider) Register(app contracts.App) error {
	h := New()
	h.Disk(app.BasePath("storage"))
	app.Container().Instance("health", contractHealth{Manager: h})
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
