package orm

import (
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"

	_ "modernc.org/sqlite"
)

func TestConfigureIsMutexGuarded(t *testing.T) {
	db, err := sql.Open("sqlite", "file:orm_configure?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var wg sync.WaitGroup
	var failed atomic.Bool
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Configure(db, "sqlite")
			got, driver := configuredConn()
			if got != db || driver != "sqlite" {
				failed.Store(true)
			}
		}()
	}
	wg.Wait()
	if failed.Load() {
		t.Fatal("configured connection raced")
	}
	Configure(db, "sqlite")
	if configuredDB() != db {
		t.Fatal("database service Configure must be the ORM connection owner")
	}
}
