package httpclient

import (
	"github.com/zatrano/framework/v3/core/bootstrap/addons"
	"github.com/zatrano/framework/v3/core/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "httpclient",
		Key:         "httpclient",
		Description: "Outbound HTTP client",
		Order:       80,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the httpclient package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
