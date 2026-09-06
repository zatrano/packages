package queue

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "queue",
		Key:         "queue",
		Description: "Job queues",
		Order:       40,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
	})
}

// ServiceProvider boots the queue package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
