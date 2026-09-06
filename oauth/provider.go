package oauth

import (
	"strings"

	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	pkgconfig "github.com/zatrano/framework/kernel/config"
	"github.com/zatrano/framework/kernel/env"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "oauth",
		Key:         "oauth",
		Description: "OAuth2 authorization server",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the OAuth addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	pkgconfig.LoadIfAbsent(app.Config(), "oauth", DefaultConfig())
	var server *Server
	storePath := strings.TrimSpace(app.Config().GetString("oauth.store_path", env.Get("OAUTH_STORE_PATH", "")))
	if storePath != "" {
		oa, err := NewWithStore(storePath)
		if err != nil {
			if app.Logger() != nil {
				app.Logger().Errorf("oauth store: %v", err)
			}
			server = New()
		} else {
			server = oa
		}
	} else {
		server = New()
	}
	app.Container().Instance("oauth", server)
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
