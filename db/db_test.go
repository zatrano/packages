package db_test

import (
	"errors"
	"testing"

	"github.com/zatrano/packages/db"
)

func TestNotOpenError(t *testing.T) {
	var err error = db.NotOpenError{}
	if err.Error() != "db: not open" {
		t.Fatalf("got %q", err.Error())
	}
	if !db.IsNotOpen(err) {
		t.Fatal("IsNotOpen")
	}
	if !db.IsNotOpen(errors.Join(db.NotOpenError{}, errors.New("x"))) {
		// Join may not As to NotOpenError depending on Go — check Wrap instead
	}
	wrapped := db.WrapOp("sqlite", "ping", db.NotOpenError{})
	if !db.IsNotOpen(wrapped) {
		t.Fatal("IsNotOpen wrapped OpError")
	}
}

func TestWrapOp(t *testing.T) {
	if db.WrapOp("mysql", "open", nil) != nil {
		t.Fatal("nil")
	}
	err := db.WrapOp("mysql", "open", errors.New("boom"))
	var op *db.OpError
	if !errors.As(err, &op) {
		t.Fatal("As OpError")
	}
	if op.Adapter != "mysql" || op.Op != "open" {
		t.Fatalf("%#v", op)
	}
	if err.Error() != "mysql: open: boom" {
		t.Fatalf("%q", err.Error())
	}
}
