package maintenance

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/kernel/routing"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "maintenance",
		Key:         "maintenance",
		Description: "Maintenance mode",
		Order:       21,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
	})
}

type ServiceProvider struct{}

type contractMaintenance struct{ *Manager }

func (c contractMaintenance) Enable(payload contracts.MaintenancePayload) error {
	return c.Manager.Enable(Payload{
		Message:    payload.Message,
		RetryAfter: payload.RetryAfter,
		AllowedIPs: payload.AllowedIPs,
		Secret:     payload.Secret,
		Time:       payload.Time,
	})
}

func (p *ServiceProvider) Register(app contracts.App) error {
	m := New(app.BasePath("storage", "framework"))
	app.Container().Instance("maintenance", contractMaintenance{Manager: m})
	app.Container().Instance("maintenance.inner", m)
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error {
	r := routing.From(app)
	raw, err := app.Make("maintenance.inner")
	if r == nil || err != nil {
		return nil
	}
	m, _ := raw.(*Manager)
	if m != nil {
		r.Use(m.Middleware())
	}
	return nil
}
