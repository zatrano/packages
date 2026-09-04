package cache

import (
	"context"
	"fmt"

	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/env"
	"github.com/zatrano/packages/redisx"
)

func boot(app contracts.App) error {
	fileStore, err := NewFileStore(app.BasePath("storage", "framework", "cache"))
	if err != nil {
		return err
	}
	stores := map[string]Store{
		"file":   fileStore,
		"memory": NewMemoryStore(),
	}
	redisClient, redisErr := redisx.Connect(redisx.Config{
		Host:     env.Get("REDIS_HOST", "127.0.0.1"),
		Port:     env.Get("REDIS_PORT", "6379"),
		Password: env.Get("REDIS_PASSWORD"),
		DB:       redisx.ParseDB(env.Get("REDIS_DB", "0")),
	})
	if redisErr == nil {
		stores["redis"] = NewRedisStore(redisClient, "zatrano_cache:")
		app.Container().Instance("redis", redisClient)
	} else if app.Logger() != nil {
		app.Logger().Debugf("redis unavailable, skipping redis cache/queue: %v", redisErr)
	}
	mgr := NewManager(env.Get("CACHE_STORE", "file"), stores)
	app.Container().Instance("cache", mgr)
	if app.Health() != nil {
		app.Health().Custom("cache", func(ctx context.Context) error {
			if From(app) == nil {
				return fmt.Errorf("cache unavailable")
			}
			return From(app).Put("health:ping", "ok", 0)
		})
	}
	return nil
}
