package sqlite_test

import (
	"database/sql"
	"testing"

	_ "github.com/zatrano/packages/database/driver/sqlite"
)

func TestOpenMemory(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
}
