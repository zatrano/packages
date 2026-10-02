package auth

import (
	"github.com/zatrano/framework/v3/core/bootstrap/addons"
	"github.com/zatrano/framework/v3/core/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "auth",
		Key:         "auth",
		Description: "Authentication guards",
		Order:       50,
		Requires:    []string{"hashing", "session"},
		Optional:    []string{"db", "cache", "notification", "authorization", "facts", "url"},
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
	})
}

// ServiceProvider boots the auth package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
