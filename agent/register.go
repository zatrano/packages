package agent

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "agent",
		Key:         "agent",
		Description: "AI agent loop",
		Order:       201,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the agent package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
