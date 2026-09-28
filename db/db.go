// Package db is the ZATRANO V3 SQL-first database infrastructure.
//
// It is not an ORM and does not store application SQL. Application SQL lives in
// database/{migrations,queries,sqlc,seeders}. Drivers are selected by importing
// an adapter package (compile-time), never via a runtime registry.
package db

import "context"

// DB is the minimal database-independent lifecycle contract.
type DB interface {
	Ping(ctx context.Context) error
	Close() error
	BeginTx(ctx context.Context) (Tx, error)
}

// Tx is a database transaction.
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Config holds shared connection settings. Adapter packages extend this
// with vendor-specific pool options in their own Config types.
type Config struct {
	DSN string
}

// ErrNotOpen is returned when operations run against a closed or nil handle.
type NotOpenError struct{}

func (NotOpenError) Error() string { return "db: not open" }
