package session

import "github.com/zatrano/framework/v2/kernel/env"

// DefaultConfig returns session configuration.
func DefaultConfig() map[string]any {
	return map[string]any{
		"driver":   env.Get("SESSION_DRIVER", "file"),
		"lifetime": env.GetInt("SESSION_LIFETIME", 120),
		"path":     "storage/framework/sessions",
		"cookie":   "zatrano_session",
	}
}
