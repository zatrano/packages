package mariadb_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/zatrano/packages/db/mariadb"
)

func TestOpenEmptyDSN(t *testing.T) {
	_, err := mariadb.Open(context.Background(), mariadb.Config{})
	if err == nil {
		t.Fatal("expected empty DSN error")
	}
}

func TestOpenIntegration(t *testing.T) {
	dsn := os.Getenv("ZATRANO_TEST_MARIADB_DSN")
	if dsn == "" {
		t.Skip("set ZATRANO_TEST_MARIADB_DSN for MariaDB integration (first-class; not MySQL alias)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	d, err := mariadb.Open(ctx, mariadb.Config{DSN: dsn, MaxOpenConns: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if err := d.Ping(ctx); err != nil {
		t.Fatal(err)
	}
}
