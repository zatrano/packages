package wellknown

import (
	"strings"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/env"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "wellknown",
		Key:         "wellknown",
		Description: "security.txt / well-known",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the wellknown addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	base := strings.TrimRight(app.Config().GetString("app.url", env.Get("APP_URL", "http://localhost:8080")), "/")
	app.Container().Instance("wellknown", New(Config{
		ContactEmail:  env.Get("SECURITY_CONTACT_EMAIL", "security@zatrano.test"),
		ContactURL:    env.Get("SECURITY_CONTACT_URL", base+"/contact"),
		Canonical:     base + "/.well-known/security.txt",
		PolicyURL:     env.Get("SECURITY_POLICY_URL", base+"/documentation"),
		PreferredLang: env.Get("APP_LOCALE", "en"),
	}))
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
