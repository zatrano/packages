package url

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/env"
	"github.com/zatrano/framework/routing"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "url",
		Key:         "url",
		Description: "URL generator",
		Order:       18,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	r := routing.From(app)
	if r == nil {
		return nil
	}
	g := New(r, app.Config().GetString("app.url", env.Get("APP_URL", "http://localhost:8080")))
	g.SetSigningKey(app.Config().GetString("app.key", env.Get("APP_KEY")))
	app.Container().Instance("url", g)
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
