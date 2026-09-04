package backup

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "backup",
		Key:         "backup",
		Description: "Database backup/restore (SQLite + native dump tools)",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
		CLI:         backupCLI,
	})
}

// ServiceProvider boots the backup addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	mgr, err := ManagerFromApp(app, "")
	if err != nil {
		return err
	}
	app.Container().Instance("backup", mgr)
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
