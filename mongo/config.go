package mongo

import "github.com/zatrano/framework/v3/core/kernel/env"

// DefaultConfig returns MongoDB configuration defaults.
func DefaultConfig() map[string]any {
	return map[string]any{
		"uri": env.Get("MONGO_URI", "memory"),
	}
}
