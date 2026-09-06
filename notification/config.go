package notification

import "github.com/zatrano/framework/v2/kernel/env"

// DefaultConfig returns notification configuration.
func DefaultConfig() map[string]any {
	return map[string]any{
		"default_channels": env.Get("NOTIFICATION_CHANNELS", "database,mail"),
		"sms_from":         env.Get("SMS_FROM", env.Get("APP_NAME", "ZATRANO")),
		"sms_driver":       env.Get("SMS_DRIVER", "memory"),
	}
}
