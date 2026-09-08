package webhooks

import (
	"fmt"
	"strings"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/env"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "webhooks",
		Key:         "webhooks",
		Description: "Outbound signed webhooks",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the webhooks addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	secret := strings.TrimSpace(env.Get("WEBHOOK_SECRET"))
	if err := rejectInsecureWebhookSecret(secret); err != nil {
		return err
	}
	mgr := New()
	if url := strings.TrimSpace(env.Get("WEBHOOK_URL")); url != "" {
		mgr.Register(Endpoint{
			URL:    url,
			Secret: secret,
			Events: []string{"user.created", "demo.ping", "*"},
		})
	}
	app.Container().Instance("webhooks", mgr)
	return nil
}

func rejectInsecureWebhookSecret(secret string) error {
	if secret == "" || secret == "zatrano-webhook-secret" {
		return fmt.Errorf("webhooks: WEBHOOK_SECRET is required")
	}
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
