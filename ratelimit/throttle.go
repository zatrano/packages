package ratelimit

import (
	"time"

	"github.com/zatrano/framework/v2/kernel/http"
	"github.com/zatrano/framework/v2/kernel/middleware"
	"github.com/zatrano/framework/v2/kernel/routing"
)

var _ middleware.AttemptLimiter = (*Limiter)(nil)

// Middleware is the HTTP adapter around kernel/middleware.Throttle.
func Middleware(limiter *Limiter, maxAttempts int, decay time.Duration, key func(*http.Request) string) routing.MiddlewareFunc {
	return middleware.Throttle(limiter, maxAttempts, decay, key)
}

// PerIP limits by client IP.
func PerIP(limiter *Limiter, maxAttempts int, decay time.Duration) routing.MiddlewareFunc {
	return Middleware(limiter, maxAttempts, decay, func(req *http.Request) string {
		return "ip:" + req.IP()
	})
}

// ThrottleRequests applies a named rate limiter policy.
func ThrottleRequests(limiter *Limiter, name string) routing.MiddlewareFunc {
	if limiter == nil {
		return func(next routing.HandlerFunc) routing.HandlerFunc {
			return next
		}
	}
	return limiter.Named(name)
}

// ThrottleRequestsWith limits requests using max attempts, decay, and optional key resolver.
func ThrottleRequestsWith(limiter *Limiter, maxAttempts int, decay time.Duration, keyFn ...func(*http.Request) string) routing.MiddlewareFunc {
	if limiter == nil {
		return func(next routing.HandlerFunc) routing.HandlerFunc {
			return next
		}
	}
	key := func(req *http.Request) string {
		return "ip:" + req.IP()
	}
	if len(keyFn) > 0 && keyFn[0] != nil {
		key = keyFn[0]
	}
	return Middleware(limiter, maxAttempts, decay, key)
}
