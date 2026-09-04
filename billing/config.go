package billing

import "github.com/zatrano/framework/packages/env"

// DefaultConfig returns billing configuration defaults.
func DefaultConfig() map[string]any {
	return map[string]any{
		"default":               env.Get("BILLING_GATEWAY", "memory"),
		"stripe_secret":         env.Get("STRIPE_SECRET_KEY", ""),
		"stripe_webhook_secret": env.Get("STRIPE_WEBHOOK_SECRET", ""),
		"success_url":           env.Get("BILLING_SUCCESS_URL", ""),
		"cancel_url":            env.Get("BILLING_CANCEL_URL", ""),
	}
}
