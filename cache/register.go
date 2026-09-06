package cache

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "cache",
		Key:         "cache",
		Description: "Cache manager",
		Order:       20,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
	})
}

// ServiceProvider boots the cache package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
