package apitoken

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/auth"
	"github.com/zatrano/packages/database"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "apitoken",
		Key:         "apitoken",
		Description: "Personal access tokens",
		Order:       52,
		Requires:    []string{"auth"},
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
	if dbMgr := database.From(app); dbMgr != nil {
		if db, err := dbMgr.DB(); err == nil && db != nil {
			driver, _ := dbMgr.DriverName()
			store = NewDatabaseStore(db, driver)
		}
	}
	app.Container().Instance("tokens", New(store, provider))
	return nil
}
