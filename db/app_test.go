package db_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/zatrano/packages/db"
)

type fakeApp struct {
	val any
	err error
}

func (a fakeApp) Make(string) (any, error) {
	if a.err != nil {
		return nil, a.err
	}
	return a.val, nil
}

type nativeWrap struct{ db *sql.DB }

func (n nativeWrap) Native() *sql.DB { return n.db }

type legacyWrap struct {
	db     *sql.DB
	driver string
	err    error
}

func (l legacyWrap) DB(...string) (*sql.DB, error) { return l.db, l.err }
func (l legacyWrap) DriverName(...string) (string, error) {
	return l.driver, nil
}

func TestSQLFromNilApp(t *testing.T) {
	if _, ok := db.SQLFrom(nil); ok {
		t.Fatal("nil app must miss")
	}
}

func TestSQLFromMissing(t *testing.T) {
	if _, ok := db.SQLFrom(fakeApp{err: errors.New("missing")}); ok {
		t.Fatal("missing binding must miss")
	}
}

func TestSQLFromSQLDB(t *testing.T) {
	sqlDB := &sql.DB{}
	h, ok := db.SQLFrom(fakeApp{val: sqlDB})
	if !ok || h.DB != sqlDB {
		t.Fatalf("got %#v ok=%v", h, ok)
	}
}

func TestSQLFromNative(t *testing.T) {
	sqlDB := &sql.DB{}
	h, ok := db.SQLFrom(fakeApp{val: nativeWrap{db: sqlDB}})
	if !ok || h.DB != sqlDB {
		t.Fatalf("got %#v ok=%v", h, ok)
	}
}

func TestSQLFromLegacyManager(t *testing.T) {
	sqlDB := &sql.DB{}
	h, ok := db.SQLFrom(fakeApp{val: legacyWrap{db: sqlDB, driver: "pgsql"}})
	if !ok || h.DB != sqlDB || h.Driver != "pgsql" {
		t.Fatalf("got %#v ok=%v", h, ok)
	}
}

func TestNormalizeDriver(t *testing.T) {
	if got := db.NormalizeDriver("postgres"); got != "pgsql" {
		t.Fatalf("got %q", got)
	}
	if got := db.DefaultPort("", "5432"); got != "5432" {
		t.Fatalf("got %q", got)
	}
}
