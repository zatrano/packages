package webauthn

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	pkgconfig "github.com/zatrano/framework/packages/config"
	"github.com/zatrano/framework/packages/env"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "webauthn",
		Key:         "webauthn",
		Description: "WebAuthn/passkeys (separate module)",
		Heavy:       true,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the WebAuthn addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	pkgconfig.LoadIfAbsent(app.Config(), "webauthn", DefaultConfig())
	cfg := app.Config()
	rpID := cfg.GetString("webauthn.rp_id", env.Get("WEBAUTHN_RP_ID", ""))
	rpOrigin := cfg.GetString("webauthn.rp_origin", env.Get("WEBAUTHN_RP_ORIGIN", ""))
	rpName := cfg.GetString("webauthn.rp_display_name", env.Get("WEBAUTHN_RP_DISPLAY_NAME", env.Get("WEBAUTHN_RP_NAME", env.Get("APP_NAME", "ZATRANO"))))
	app.Container().Instance("webauthn", New(rpID, rpOrigin, rpName))
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
