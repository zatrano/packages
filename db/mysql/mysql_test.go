package mysql_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/zatrano/packages/db/mysql"
)

func TestOpenEmptyDSN(t *testing.T) {
	_, err := mysql.Open(context.Background(), mysql.Config{})
	if err == nil {
		t.Fatal("expected empty DSN error")
	}
}

func TestOpenIntegration(t *testing.T) {
	dsn := os.Getenv("ZATRANO_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set ZATRANO_TEST_MYSQL_DSN for MySQL integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	d, err := mysql.Open(ctx, mysql.Config{DSN: dsn, MaxOpenConns: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	tx, err := d.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
}
