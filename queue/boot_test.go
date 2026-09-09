package queue

import (
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/zatrano/framework/v2/kernel"
	"github.com/zatrano/packages/redisx"
)

func TestBootSyncWithoutRedis(t *testing.T) {
	t.Setenv("QUEUE_CONNECTION", "sync")
	app := kernel.NewApplication(t.TempDir())
	if err := boot(app); err != nil {
		t.Fatal(err)
	}
	if From(app) == nil || From(app).Queue() == nil {
		t.Fatal("sync queue must boot")
	}
}

func TestBootRedisWithoutBindingFails(t *testing.T) {
	t.Setenv("QUEUE_CONNECTION", "redis")
	err := boot(kernel.NewApplication(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "redis") {
		t.Fatalf("expected deterministic redis failure, got %v", err)
	}
}

func TestBootRedisWithClientBinding(t *testing.T) {
	t.Setenv("QUEUE_CONNECTION", "redis")
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { _ = client.Close() })
	app := kernel.NewApplication(t.TempDir())
	app.Container().Instance("redis", client)
	if err := boot(app); err != nil {
		t.Fatal(err)
	}
	if From(app) == nil || From(app).Queue() == nil {
		t.Fatal("redis queue must boot when a redis client is bound")
	}
}

func TestBootRedisWhenBindingPublished(t *testing.T) {
	t.Setenv("QUEUE_CONNECTION", "redis")
	client, err := redisx.Connect(redisx.Config{Host: "127.0.0.1", Port: "6379"})
	if err != nil {
		t.Skip("redis not available")
	}
	t.Cleanup(func() { _ = client.Close() })
	app := kernel.NewApplication(t.TempDir())
	app.Container().Instance("redis", client)
	if err := boot(app); err != nil {
		t.Fatal(err)
	}
	if From(app) == nil || From(app).Queue() == nil {
		t.Fatal("redis queue must boot when cache published the binding")
	}
}
