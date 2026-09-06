package webauthn

import "github.com/zatrano/framework/v2/kernel/env"

// DefaultConfig returns WebAuthn/passkey configuration defaults.
func DefaultConfig() map[string]any {
	return map[string]any{
		"rp_id":           env.Get("WEBAUTHN_RP_ID", ""),
		"rp_origin":       env.Get("WEBAUTHN_RP_ORIGIN", ""),
		"rp_display_name": env.Get("WEBAUTHN_RP_DISPLAY_NAME", env.Get("WEBAUTHN_RP_NAME", env.Get("APP_NAME", "ZATRANO"))),
	}
}
