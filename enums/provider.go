package enums

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "enums",
		Key:         "enums",
		Description: "String enum registry",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         Commands,
	})
}

// ServiceProvider boots the enums addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	reg := NewRegistry()
	reg.Register(NewString("post_status", "draft:Draft", "published:Published", "archived:Archived"))
	reg.Register(NewString("user_role", "admin", "editor", "viewer"))
	app.Container().Instance("enums", reg)
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
