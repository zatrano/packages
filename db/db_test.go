package db_test

import (
	"testing"

	"github.com/zatrano/packages/db"
)

func TestNotOpenError(t *testing.T) {
	var err error = db.NotOpenError{}
	if err.Error() != "db: not open" {
		t.Fatalf("got %q", err.Error())
	}
}
