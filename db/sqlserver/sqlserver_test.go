package sqlserver_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/zatrano/packages/db/sqlserver"
)

func TestOpenEmptyDSN(t *testing.T) {
	_, err := sqlserver.Open(context.Background(), sqlserver.Config{})
	if err == nil {
		t.Fatal("expected empty DSN error")
	}
}

func TestOpenIntegration(t *testing.T) {
	dsn := os.Getenv("ZATRANO_TEST_SQLSERVER_DSN")
	if dsn == "" {
		t.Skip("set ZATRANO_TEST_SQLSERVER_DSN for SQL Server integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	d, err := sqlserver.Open(ctx, sqlserver.Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}
	_ = d.Close()
}
