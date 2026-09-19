package facts

import (
	"time"

	"github.com/zatrano/framework/v2/contracts"
	pkgconfig "github.com/zatrano/framework/v2/kernel/config"
)

func boot(app contracts.App) (*Bus, error) {
	pkgconfig.LoadIfAbsent(app.Config(), "facts", DefaultConfig())
	workers := app.Config().GetInt("facts.async.workers", 4)
	size := app.Config().GetInt("facts.async.queue_size", 256)
	attempts := app.Config().GetInt("facts.async.max_attempts", 3)
	backoff := parseDuration(app.Config().GetString("facts.async.backoff", "1s"), time.Second)
	stop := parseDuration(app.Config().GetString("facts.async.stop_timeout", "15s"), 15*time.Second)
	bus := New(
		WithWorkers(workers),
		WithQueueSize(size),
		WithRetry(attempts, backoff),
		WithStopTimeout(stop),
	)
	app.Container().Instance("facts", bus)
	return bus, nil
}

// DefaultConfig is merged into application config when facts is enabled.
func DefaultConfig() map[string]any {
	return map[string]any{
		"async": map[string]any{
			"workers":      4,
			"queue_size":   256,
			"max_attempts": 3,
			"backoff":      "1s",
			"stop_timeout": "15s",
		},
	}
}

func parseDuration(raw string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(raw)
	if err != nil || d < 0 {
		return fallback
	}
	return d
}
