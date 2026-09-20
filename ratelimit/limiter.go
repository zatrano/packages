package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"github.com/zatrano/framework/v2/kernel/http"
	"github.com/zatrano/framework/v2/kernel/routing"
)

// Limiter tracks request attempts.
type Limiter struct {
	mu       sync.Mutex
	attempts map[string]*bucket
	named    map[string]Limit
}

type bucket struct {
	hits       int
	expiresAt  time.Time
	retryAfter time.Duration
}

// Limit describes a named rate limit policy.
type Limit struct {
	MaxAttempts int
	Decay       time.Duration
	Key         func(*http.Request) string
}

// New creates a rate limiter.
func New() *Limiter {
	return &Limiter{
		attempts: make(map[string]*bucket),
		named:    make(map[string]Limit),
	}
}

// For registers a named rate limit.
func (l *Limiter) For(name string, limit Limit) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if limit.MaxAttempts <= 0 {
		limit.MaxAttempts = 60
	}
	if limit.Decay <= 0 {
		limit.Decay = time.Minute
	}
	if limit.Key == nil {
		limit.Key = func(req *http.Request) string {
			return "ip:" + req.IP()
		}
	}
	l.named[name] = limit
}

// Has reports whether a named limit exists.
func (l *Limiter) Has(name string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.named[name]
	return ok
}

// Named returns middleware for a previously registered limit.
func (l *Limiter) Named(name string) routing.MiddlewareFunc {
	l.mu.Lock()
	limit, ok := l.named[name]
	l.mu.Unlock()
	if !ok {
		return func(next routing.HandlerFunc) routing.HandlerFunc {
			return func(req *http.Request) *http.Response {
				return http.Abort(500, fmt.Sprintf("rate limiter [%s] not defined", name))
			}
		}
	}
	return Middleware(l, limit.MaxAttempts, limit.Decay, func(req *http.Request) string {
		return name + ":" + limit.Key(req)
	})
}

// TooManyAttempts reports whether the key has exceeded maxAttempts.
func (l *Limiter) TooManyAttempts(key string, maxAttempts int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.attempts[key]
	if !ok {
		return false
	}
	if time.Now().After(b.expiresAt) {
		delete(l.attempts, key)
		return false
	}
	return b.hits >= maxAttempts
}

// Take records one attempt atomically. allowed is false when the key is already at the limit.
func (l *Limiter) Take(key string, maxAttempts int, decay time.Duration) (allowed bool, remaining int, retryAfter int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.attempts[key]
	if ok && now.After(b.expiresAt) {
		delete(l.attempts, key)
		ok = false
	}
	if ok && b.hits >= maxAttempts {
		sec := int(time.Until(b.expiresAt).Seconds())
		if sec < 0 {
			sec = 0
		}
		return false, 0, sec
	}
	if !ok {
		l.attempts[key] = &bucket{
			hits:       1,
			expiresAt:  now.Add(decay),
			retryAfter: decay,
		}
		rem := maxAttempts - 1
		if rem < 0 {
			rem = 0
		}
		return true, rem, 0
	}
	b.hits++
	rem := maxAttempts - b.hits
	if rem < 0 {
		rem = 0
	}
	return true, rem, 0
}

// Hit increments the attempt counter.
func (l *Limiter) Hit(key string, decay time.Duration) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.attempts[key]
	if !ok || now.After(b.expiresAt) {
		l.attempts[key] = &bucket{
			hits:       1,
			expiresAt:  now.Add(decay),
			retryAfter: decay,
		}
		return 1
	}
	b.hits++
	return b.hits
}

// Attempts returns current attempts for a key.
func (l *Limiter) Attempts(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.attempts[key]
	if !ok || time.Now().After(b.expiresAt) {
		return 0
	}
	return b.hits
}

// AvailableIn returns seconds until the key is available again.
func (l *Limiter) AvailableIn(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.attempts[key]
	if !ok {
		return 0
	}
	remaining := time.Until(b.expiresAt)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Seconds())
}

// Clear clears attempts for a key.
func (l *Limiter) Clear(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}
