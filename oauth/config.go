package oauth

import "github.com/zatrano/framework/kernel/env"

// DefaultConfig returns OAuth2 server configuration defaults.
func DefaultConfig() map[string]any {
	return map[string]any{
		"store_path": env.Get("OAUTH_STORE_PATH", ""),
	}
}
