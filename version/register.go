package version

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "version",
		Key:         "version",
		Description: "Version helper",
		Order:       5,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	_ = LoadFile(app.BasePath("VERSION"))
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
