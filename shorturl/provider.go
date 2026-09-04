package shorturl

import (
	"strings"

	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/env"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "shorturl",
		Key:         "shorturl",
		Description: "Short URL manager",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the shorturl addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	base := strings.TrimRight(app.Config().GetString("app.url", env.Get("APP_URL", "http://localhost:8080")), "/")
	app.Container().Instance("shorturl", New(base, env.Get("SHORTURL_PREFIX", "/s")))
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
