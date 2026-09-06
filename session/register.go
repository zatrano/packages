package session

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "session",
		Key:         "session",
		Description: "HTTP sessions",
		Order:       130,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the session package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
