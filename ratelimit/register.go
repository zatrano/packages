package ratelimit

import (
	"time"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "ratelimit",
		Key:         "rateLimiter",
		Description: "Rate limiter",
		Order:       17,
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

type ServiceProvider struct{}

type contractLimiter struct{ inner *Limiter }

func (c *contractLimiter) For(name string, limit contracts.RateLimit) {
	if c == nil || c.inner == nil {
		return
	}
	c.inner.For(name, Limit{MaxAttempts: limit.MaxAttempts, Decay: limit.Decay})
}

func (p *ServiceProvider) Register(app contracts.App) error {
	lim := New()
	lim.For("api", Limit{MaxAttempts: 60, Decay: time.Minute})
	lim.For("login", Limit{MaxAttempts: 5, Decay: time.Minute})
	app.Container().Instance("rateLimiter", &contractLimiter{inner: lim})
	app.Container().Instance("rateLimiter.inner", lim)
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
