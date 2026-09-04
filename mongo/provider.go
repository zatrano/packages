package mongo

import (
	"fmt"

	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	pkgconfig "github.com/zatrano/framework/packages/config"
	"github.com/zatrano/framework/packages/env"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "mongo",
		Key:         "mongo",
		Description: "MongoDB client (separate module)",
		Heavy:       true,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the MongoDB client addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	pkgconfig.LoadIfAbsent(app.Config(), "mongo", DefaultConfig())
	if app.Bound("mongo") {
		return nil
	}
	uri := app.Config().GetString("mongo.uri", env.Get("MONGO_URI", "memory"))
	client := Connect(uri)
	if err := client.Ping(); err != nil {
		return fmt.Errorf("mongo: %w", err)
	}
	app.Container().Instance("mongo", client)
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
