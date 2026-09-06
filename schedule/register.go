package schedule

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "schedule",
		Key:         "schedule",
		Description: "Task scheduler",
		Order:       70,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
	})
}

// ServiceProvider boots the schedule package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
