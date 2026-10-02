package oracle_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/zatrano/packages/db/oracle"
)

func TestOpenEmptyDSN(t *testing.T) {
	_, err := oracle.Open(context.Background(), oracle.Config{})
	if err == nil {
		t.Fatal("expected empty DSN error")
	}
}

func TestOpenIntegration(t *testing.T) {
	dsn := os.Getenv("ZATRANO_TEST_ORACLE_DSN")
	if dsn == "" {
		t.Skip("set ZATRANO_TEST_ORACLE_DSN for Oracle integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	d, err := oracle.Open(ctx, oracle.Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}
	_ = d.Close()
}
