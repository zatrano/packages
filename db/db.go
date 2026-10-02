package db

import "context"

// DB is the minimal database-independent lifecycle contract.
// Vendor types (pgxpool, *sql.DB, …) are never part of this interface — use
// each adapter's Native() escape hatch when needed.
type DB interface {
	Ping(ctx context.Context) error
	Close() error
	BeginTx(ctx context.Context) (Tx, error)
}

// Tx is a database transaction. Boundaries are explicit (no magic).
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Config holds shared connection settings. Adapter packages define their own
// Config types with vendor-specific pool options — do not grow this struct
// with PostgreSQL/MySQL-specific fields.
type Config struct {
	DSN string
}
