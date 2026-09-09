package cache

import (
	"strings"
	"testing"

	"github.com/zatrano/framework/v2/kernel"
	"github.com/zatrano/packages/redisx"
)

func TestBootFileStoreWithoutRedis(t *testing.T) {
	t.Setenv("CACHE_STORE", "file")
	t.Setenv("REDIS_HOST", "127.0.0.1")
	t.Setenv("REDIS_PORT", "1")
	app := kernel.NewApplication(t.TempDir())
	if err := boot(app); err != nil {
		t.Fatal(err)
	}
	if From(app) == nil || From(app).Store() == nil {
		t.Fatal("file store must boot")
	}
	if _, err := app.Make("redis"); err == nil {
		t.Fatal("redis must not be published when connect fails")
	}
}

func TestBootRedisStoreWithoutRedisFails(t *testing.T) {
	t.Setenv("CACHE_STORE", "redis")
	t.Setenv("REDIS_HOST", "127.0.0.1")
	t.Setenv("REDIS_PORT", "1")
	err := boot(kernel.NewApplication(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "redis") {
		t.Fatalf("expected deterministic redis failure, got %v", err)
	}
}

func TestBootRedisStoreWhenRedisAvailable(t *testing.T) {
	t.Setenv("CACHE_STORE", "redis")
	t.Setenv("REDIS_HOST", "127.0.0.1")
	t.Setenv("REDIS_PORT", "6379")
	client, err := redisx.Connect(redisx.Config{Host: "127.0.0.1", Port: "6379"})
	if err != nil {
		t.Skip("redis not available")
	}
	_ = client.Close()
	app := kernel.NewApplication(t.TempDir())
	if err := boot(app); err != nil {
		t.Fatal(err)
	}
	if From(app) == nil || From(app).Store() == nil {
		t.Fatal("redis store must boot")
	}
	if _, err := app.Make("redis"); err != nil {
		t.Fatal("cache must publish redis binding")
	}
}
