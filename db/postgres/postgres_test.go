package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/zatrano/packages/db/postgres"
)

func TestOpenEmptyDSN(t *testing.T) {
	_, err := postgres.Open(context.Background(), postgres.Config{})
	if err == nil {
		t.Fatal("expected empty DSN error")
	}
}

func TestOpenIntegration(t *testing.T) {
	dsn := os.Getenv("ZATRANO_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set ZATRANO_TEST_POSTGRES_DSN for PostgreSQL integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	d, err := postgres.Open(ctx, postgres.Config{DSN: dsn, MaxConns: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if d.Native() == nil {
		t.Fatal("Native pool")
	}
	tx, err := d.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
