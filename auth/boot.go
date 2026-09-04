package auth

import (
	"fmt"
	"strings"
	"time"

	appconfig "github.com/zatrano/framework/config"
	pkgconfig "github.com/zatrano/framework/config"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/authorization"
	"github.com/zatrano/packages/bootutil"
	"github.com/zatrano/packages/cache"
	"github.com/zatrano/packages/database"
	"github.com/zatrano/packages/events"
	"github.com/zatrano/packages/notification"
	"github.com/zatrano/packages/session"
)

func boot(app contracts.App) error {
	pkgconfig.LoadIfAbsent(app.Config(), "auth", appconfig.Auth())
	app.Container().Instance("gate", authorization.New())
	authManager := NewManager(app.Config().GetString("auth.defaults.guard", "web"))
	authManager.SetSessionManager(session.From(app))
	authManager.SetDispatcher(events.From(app))
	if max := app.Config().GetInt("auth.lockout.max_attempts", 5); max > 0 {
		decayMin := app.Config().GetInt("auth.lockout.decay_minutes", 1)
		if decayMin <= 0 {
			decayMin = 1
		}
		authManager.SetLockout(max, time.Duration(decayMin)*time.Minute)
	}
	if issuer := strings.TrimSpace(app.Config().GetString("auth.two_factor.issuer", "")); issuer != "" {
		authManager.SetTwoFactorIssuer(issuer)
	}
	authManager.SetRememberDeviceDays(app.Config().GetInt("auth.two_factor.remember_device_days", 30))
	if c := cache.From(app); c != nil {
		authManager.SetLockoutCache(c)
	}
	if dbMgr := database.From(app); dbMgr != nil {
		db, err := dbMgr.DB()
		if err == nil {
			driver, _ := dbMgr.DriverName()
			providers := map[string]UserProvider{}
			if rawProviders, ok := app.Config().Get("auth.providers").(map[string]any); ok {
				for name, raw := range rawProviders {
					pcfg, _ := raw.(map[string]any)
					table := "users"
					if pcfg != nil {
						if t := strings.TrimSpace(fmt.Sprint(pcfg["table"])); t != "" && t != "<nil>" {
							table = t
						}
					}
					providers[name] = NewDatabaseUserProvider(db, driver, table)
				}
			}
			if len(providers) == 0 {
				providers["users"] = NewDatabaseUserProvider(db, driver, "users")
			}
			defaultProvider := app.Config().GetString("auth.defaults.provider", "users")
			if rawGuards, ok := app.Config().Get("auth.guards").(map[string]any); ok {
				for name, raw := range rawGuards {
					gcfg, _ := raw.(map[string]any)
					guardDriver := "session"
					providerName := defaultProvider
					if gcfg != nil {
						if d := strings.TrimSpace(fmt.Sprint(gcfg["driver"])); d != "" && d != "<nil>" {
							guardDriver = strings.ToLower(d)
						}
						if p := strings.TrimSpace(fmt.Sprint(gcfg["provider"])); p != "" && p != "<nil>" {
							providerName = p
						}
					}
					// Session guards only; token/PAT auth uses packages/apitoken middleware.
					if guardDriver != "session" {
						continue
					}
					provider := providers[providerName]
					if provider == nil {
						provider = providers["users"]
					}
					if provider == nil {
						continue
					}
					authManager.Extend(name, NewGuard(name, provider))
				}
			}
			if authManager.Guard() == nil {
				provider := providers[defaultProvider]
				if provider == nil {
					provider = providers["users"]
				}
				if provider != nil {
					authManager.Extend(authManager.GetDefaultDriver(), NewGuard(authManager.GetDefaultDriver(), provider))
				}
			}
		}
	}
	app.Container().Instance("auth", authManager)
	if enc := app.Encrypter(); enc != nil {
		authManager.SetEncrypter(enc)
	}

	authManager.SetVerificationURLGenerator(func(user Authenticatable) (string, error) {
		if user == nil || app.URL() == nil {
			return "", fmt.Errorf("verification url unavailable")
		}
		email := EmailForVerification(user)
		return app.URL().Signed("/auth/email/verify/"+fmt.Sprint(user.AuthID()), 60*time.Minute, map[string]string{
			"hash": EmailHash(email),
		})
	})
	authManager.SetEmailVerificationSender(func(user Authenticatable, verifyURL string) error {
		n := notification.From(app)
		if n == nil {
			return fmt.Errorf("notifications not configured")
		}
		return n.Send(notification.Recipient{
			ID:    fmt.Sprint(user.AuthID()),
			Email: EmailForVerification(user),
		}, notification.VerifyEmailNotification{VerifyURL: verifyURL})
	})
	authManager.SetPasswordChangedSender(func(user Authenticatable) error {
		n := notification.From(app)
		if n == nil {
			return nil
		}
		return n.Send(notification.Recipient{
			ID:    fmt.Sprint(user.AuthID()),
			Email: EmailForVerification(user),
		}, notification.PasswordChangedNotification{})
	})

	if dbMgr := database.From(app); dbMgr != nil {
		if db, err := dbMgr.DB(); err == nil {
			driver, _ := dbMgr.DriverName()
			brokerName := app.Config().GetString("auth.defaults.passwords", "users")
			passCfg, _ := app.Config().Get("auth.passwords." + brokerName).(map[string]any)
			table := "password_reset_tokens"
			expireMin := 60
			throttleSec := 60
			providerName := app.Config().GetString("auth.defaults.provider", "users")
			if passCfg != nil {
				if t := strings.TrimSpace(fmt.Sprint(passCfg["table"])); t != "" && t != "<nil>" {
					table = t
				}
				if p := strings.TrimSpace(fmt.Sprint(passCfg["provider"])); p != "" && p != "<nil>" {
					providerName = p
				}
				if v, ok := bootutil.AsInt(passCfg["expire"]); ok && v > 0 {
					expireMin = v
				}
				if v, ok := bootutil.AsInt(passCfg["throttle"]); ok && v >= 0 {
					throttleSec = v
				}
			}
			provTable := "users"
			if rawProviders, ok := app.Config().Get("auth.providers").(map[string]any); ok {
				if raw, ok := rawProviders[providerName].(map[string]any); ok {
					if t := strings.TrimSpace(fmt.Sprint(raw["table"])); t != "" && t != "<nil>" {
						provTable = t
					}
				}
			}
			provider := NewDatabaseUserProvider(db, driver, provTable)
			tokens := NewDatabaseTokenRepositoryTable(db, driver, table, time.Duration(expireMin)*time.Minute)
			passwords := NewPasswordBroker(tokens, provider, time.Duration(expireMin)*time.Minute)
			passwords.SetThrottle(time.Duration(throttleSec) * time.Second)
			passwords.SetDispatcher(events.From(app))
			passwords.SetSessionManager(session.From(app))
			passwords.SetNotifier(func(email, token, resetURL string) error {
				n := notification.From(app)
				if n == nil {
					return fmt.Errorf("notifications not configured")
				}
				return n.Send(notification.Recipient{ID: email, Email: email}, notification.PasswordResetNotification{
					Token:         token,
					ResetURL:      resetURL,
					ExpireMinutes: expireMin,
				})
			})
			app.Container().Instance("passwords", passwords)
		}
	}
	if app.Health() != nil && database.From(app) != nil {
		if db, err := database.From(app).DB(); err == nil {
			app.Health().Database(db)
		}
	}
	return nil
}
