package notification

import (
	"strings"

	"github.com/zatrano/framework/contracts"
	pkgconfig "github.com/zatrano/framework/kernel/config"
	"github.com/zatrano/framework/kernel/env"
	"github.com/zatrano/packages/broadcasting"
	"github.com/zatrano/packages/database"
	"github.com/zatrano/packages/localization"
)

func boot(app contracts.App) error {
	pkgconfig.LoadIfAbsent(app.Config(), "notifications", pkgconfig.Notifications())
	mgr := NewManager()
	if app.Logger() != nil {
		mgr.SetErrorHandler(func(err error) {
			app.Logger().Errorf("notification: %v", err)
		})
	}

	mgr.SetMail(NewMailManager(
		env.Get("MAIL_MAILER", "log"),
		env.Get("MAIL_FROM_ADDRESS", "hello@example.com"),
		env.Get("MAIL_FROM_NAME", app.Config().GetString("app.name", "ZATRANO")),
		map[string]Mailer{
			"log": NewLogMailer(app.Logger()),
			"smtp": NewSMTPMailer(SMTPConfig{
				Host:       env.Get("MAIL_HOST", "127.0.0.1"),
				Port:       env.Get("MAIL_PORT", "2525"),
				Username:   env.Get("MAIL_USERNAME"),
				Password:   env.Get("MAIL_PASSWORD"),
				Encryption: env.Get("MAIL_ENCRYPTION"),
			}),
		},
	))

	if broadcasting.From(app) != nil {
		mgr.Extend("broadcast", NewBroadcastChannel(broadcasting.From(app)))
	}
	var pushSender PushSender
	switch strings.ToLower(strings.TrimSpace(env.Get("PUSH_DRIVER", "memory"))) {
	case "http":
		pushSender = &HTTPPushSender{
			Endpoint: env.Get("PUSH_URL", ""),
			Token:    env.Get("PUSH_TOKEN", ""),
		}
	default:
		pushSender = &MemoryPushSender{}
	}
	mgr.Extend("push", NewPushChannel(pushSender))

	smsFrom := env.Get("SMS_FROM", env.Get("APP_NAME", "ZATRANO"))
	smsMgr := NewSmsManager(smsFrom)
	smsMgr.Extend("memory", &MemorySmsSender{})
	smsMgr.Extend("log", &LogSmsSender{
		Log: func(format string, args ...any) {
			if app.Logger() != nil {
				app.Logger().Infof(format, args...)
			}
		},
	})
	smsMgr.Extend("http", &HTTPSmsSender{
		Endpoint: env.Get("SMS_URL", ""),
		Token:    env.Get("SMS_TOKEN", ""),
		Method:   env.Get("SMS_METHOD", "POST"),
	})
	smsMgr.Extend("twilio", &TwilioSmsSender{
		AccountSID: env.Get("TWILIO_ACCOUNT_SID", ""),
		AuthToken:  env.Get("TWILIO_AUTH_TOKEN", ""),
		From:       env.Get("TWILIO_FROM", smsFrom),
	})
	defaultSMS := strings.ToLower(strings.TrimSpace(env.Get("SMS_DRIVER", "memory")))
	if smsMgr.Sender(defaultSMS) == nil {
		defaultSMS = "memory"
	}
	smsMgr.Use(defaultSMS)
	mgr.SetSms(smsMgr)
	if tr := localization.From(app); tr != nil {
		mgr.SetTranslator(tr)
	}
	mgr.SetMailDefaults(
		app.Config().GetString("app.locale", env.Get("APP_LOCALE", "en")),
		app.Config().GetString("app.name", env.Get("APP_NAME", "ZATRANO")),
	)
	if dbMgr := database.From(app); dbMgr != nil {
		if db, err := dbMgr.DB(); err == nil {
			driver, _ := dbMgr.DriverName()
			mgr.Extend("database", NewDatabaseChannel(db, "notifications", driver))
			mgr.SetStore(NewStore(db, "notifications", driver))
		}
	}
	app.Container().Instance("notifications", mgr)
	return nil
}
