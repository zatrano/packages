package otp

import (
	"time"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "otp",
		Key:         "otp",
		Description: "One-time passwords",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the OTP addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	app.Container().Instance("otp", New(NewMemoryStore()).WithTTL(5*time.Minute))
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
