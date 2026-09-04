package observability

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/routing"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "observability",
		Key:         "metrics",
		Description: "Metrics collector",
		Order:       19,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	m := New()
	app.Container().Instance("metrics", m)
	if app.Logger() != nil {
		app.Container().Instance("metrics-timing", Timing(m, app.Logger().Infof))
	}
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error {
	r := routing.From(app)
	raw, err := app.Make("metrics")
	if r == nil || err != nil {
		return nil
	}
	m, _ := raw.(*Metrics)
	if m == nil || app.Logger() == nil {
		return nil
	}
	r.Use(Timing(m, app.Logger().Infof))
	return nil
}
