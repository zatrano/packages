package localization

import (
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/kernel/env"
	"github.com/zatrano/framework/kernel/layout"
)

func boot(app contracts.App) error {
	locale := app.Config().GetString("app.locale", env.Get("APP_LOCALE", "en"))
	fallback := app.Config().GetString("app.fallback", env.Get("APP_FALLBACK_LOCALE", "en"))
	translator := New(layout.LocalizationDir(app), locale, fallback)
	_ = translator.Load(locale)
	if fallback != locale {
		_ = translator.Load(fallback)
	}
	app.Container().Instance("translator", translator)
	return nil
}
