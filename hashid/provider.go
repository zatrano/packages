package hashid

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/env"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "hashid",
		Key:         "hashid",
		Description: "Obfuscated public IDs",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the hashid addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	app.Container().Instance("hashid", New(
		env.Get("HASHID_SALT", app.Config().GetString("app.key", "zatrano")),
		env.GetInt("HASHID_MIN_LENGTH", 8),
	))
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
