package apitoken

import (
	"github.com/zatrano/framework/v3/core/bootstrap/addons"
	"github.com/zatrano/framework/v3/core/contracts"
	"github.com/zatrano/packages/auth"
	"github.com/zatrano/packages/db"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "apitoken",
		Key:         "apitoken",
		Description: "Personal access tokens",
		Order:       52,
		Requires:    []string{"auth"},
		Optional:    []string{"db"},
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the apitoken package.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	return boot(app)
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }

func boot(app contracts.App) error {
	var provider auth.UserProvider
	if authMgr := auth.From(app); authMgr != nil {
		if g := authMgr.Guard(); g != nil {
			provider = g.Provider()
		}
	}
	store := Store(NewMemoryStore())
	if h, ok := db.SQLFrom(app); ok && h.DB != nil {
		driver := db.NormalizeDriver(h.Driver)
		if driver == "" {
			driver = "sqlite"
		}
		store = NewDatabaseStore(h.DB, driver)
	}
	app.Container().Instance("tokens", New(store, provider))
	return nil
}
