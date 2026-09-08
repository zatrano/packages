package hashid

import (
	"fmt"
	"strings"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/env"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "hashid",
		Key:         "hashid",
		Description: "Obfuscated public IDs",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the hashid addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	salt := strings.TrimSpace(env.Get("HASHID_SALT"))
	if salt == "" {
		salt = strings.TrimSpace(app.Config().GetString("app.key", env.Get("APP_KEY")))
	}
	if err := rejectInsecureSalt(salt); err != nil {
		return err
	}
	app.Container().Instance("hashid", New(
		salt,
		env.GetInt("HASHID_MIN_LENGTH", 8),
	))
	return nil
}

func rejectInsecureSalt(salt string) error {
	if salt == "" || strings.EqualFold(salt, "zatrano") {
		return fmt.Errorf("hashid: HASHID_SALT or APP_KEY is required")
	}
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
