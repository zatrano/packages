package notification

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/view"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "notification",
		Key:         "notification",
		Description: "Notifications",
		Order:       100,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
	})
}

// ServiceProvider boots the notification package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error {
	if n := From(app); n != nil {
		if e := view.From(app); e != nil {
			n.SetMailView(e)
		}
	}
	return nil
}
