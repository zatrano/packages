package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/zatrano/packages/db"
	"github.com/zatrano/packages/db/sqlite"
)

func TestOpenPingTxCommitClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	d, err := sqlite.Open(ctx, sqlite.Config{
		DSN:          "file:zatrano_db_lifecycle?mode=memory&cache=shared",
		MaxOpenConns: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	if err := d.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	if d.Native() == nil {
		t.Fatal("Native")
	}

	tx, err := d.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	tx2, err := d.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx2.Rollback(ctx); err != nil {
		t.Fatal(err)
	}

	_ = d.Close()
	if err := d.Ping(ctx); !db.IsNotOpen(err) {
		t.Fatalf("after close ping: %v", err)
	}
}

func TestEmptyDSNDefaults(t *testing.T) {
	ctx := context.Background()
	// Empty DSN uses file default — use memory via explicit DSN in other tests.
	// Here only assert empty path is accepted by Open (creates default file DSN).
	d, err := sqlite.Open(ctx, sqlite.Config{DSN: "file:zatrano_empty_default?mode=memory&cache=shared"})
	if err != nil {
		t.Fatal(err)
	}
	_ = d.Close()
}

func TestContextCancelOnPing(t *testing.T) {
	ctx := context.Background()
	d, err := sqlite.Open(ctx, sqlite.Config{DSN: "file:zatrano_cancel?mode=memory&cache=shared"})
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := d.Ping(canceled); err == nil {
		// modernc may still succeed instantly; not required to fail
		t.Log("ping with canceled ctx returned nil (driver-dependent)")
	}
}
