package db

import (
	"database/sql"
	"strings"
)

// App is satisfied by *kernel.Application without importing the root module.
type App interface {
	Make(abstract string) (any, error)
}

// SQLHandle is a resolved database/sql handle plus optional dialect name
// (sqlite, mysql, pgsql, …). Driver may be empty when only *sql.DB is bound.
type SQLHandle struct {
	DB     *sql.DB
	Driver string
}

// SQLFrom resolves a *sql.DB from the application container key "db".
//
// Accepted bindings (in order):
//   - *sql.DB
//   - any value with Native() *sql.DB (V3 db adapters)
//   - any value with DB()/DriverName() (legacy packages/database.Manager)
//
// Returns ok=false when nothing usable is bound.
func SQLFrom(app App) (SQLHandle, bool) {
	if app == nil {
		return SQLHandle{}, false
	}
	raw, err := app.Make("db")
	if err != nil || raw == nil {
		return SQLHandle{}, false
	}
	if sqlDB, ok := raw.(*sql.DB); ok && sqlDB != nil {
		return SQLHandle{DB: sqlDB}, true
	}
	if n, ok := raw.(interface{ Native() *sql.DB }); ok {
		if sqlDB := n.Native(); sqlDB != nil {
			return SQLHandle{DB: sqlDB}, true
		}
	}
	type legacyMgr interface {
		DB(name ...string) (*sql.DB, error)
		DriverName(name ...string) (string, error)
	}
	if mgr, ok := raw.(legacyMgr); ok {
		sqlDB, err := mgr.DB()
		if err != nil || sqlDB == nil {
			return SQLHandle{}, false
		}
		driver, _ := mgr.DriverName()
		return SQLHandle{DB: sqlDB, Driver: strings.TrimSpace(driver)}, true
	}
	return SQLHandle{}, false
}

// NormalizeDriver maps common aliases to canonical dialect names used by
// queue/notification SQL dialects (sqlite, mysql, pgsql, mssql, oracle).
func NormalizeDriver(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "sqlite", "sqlite3":
		return "sqlite"
	case "mysql":
		return "mysql"
	case "pgsql", "postgres", "postgresql":
		return "pgsql"
	case "mssql", "sqlserver":
		return "mssql"
	case "oracle", "ora":
		return "oracle"
	case "mongo", "mongodb":
		return "mongo"
	default:
		return strings.ToLower(strings.TrimSpace(name))
	}
}

// DefaultPort returns port or fallback when port is blank.
func DefaultPort(port, fallback string) string {
	if strings.TrimSpace(port) == "" {
		return fallback
	}
	return port
}
