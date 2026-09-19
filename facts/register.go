package facts

import (
	"context"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
)

var _ contracts.LifecycleProvider = (*ServiceProvider)(nil)

func init() {
	addons.Register(addons.Meta{
		Name:        "facts",
		Key:         "facts",
		Description: "Typed facts and reactions",
		Order:       30,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the facts package and owns async workers.
type ServiceProvider struct {
	bus *Bus
}

func (p *ServiceProvider) Register(app contracts.App) error {
	bus, err := boot(app)
	if err != nil {
		return err
	}
	p.bus = bus
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }

func (p *ServiceProvider) Start(app contracts.App) error {
	if p == nil || p.bus == nil {
		return nil
	}
	return p.bus.Start()
}

func (p *ServiceProvider) Stop(ctx context.Context) error {
	if p == nil || p.bus == nil {
		return nil
	}
	return p.bus.Stop(ctx)
}
