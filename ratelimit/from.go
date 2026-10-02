package ratelimit

import "github.com/zatrano/framework/v3/core/contracts"

// From resolves the rate limiter from the application container.
func From(app contracts.App) *Limiter {
	if app == nil {
		return nil
	}
	if raw, err := app.Make("rateLimiter.inner"); err == nil {
		if l, ok := raw.(*Limiter); ok && l != nil {
			return l
		}
	}
	raw, err := app.Make("rateLimiter")
	if err != nil || raw == nil {
		return nil
	}
	if l, ok := raw.(*Limiter); ok {
		return l
	}
	if c, ok := raw.(*contractLimiter); ok {
		return c.inner
	}
	return nil
}
