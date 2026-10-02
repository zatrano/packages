package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zatrano/packages/db"
)

// Config is PostgreSQL-specific pool configuration.
type Config struct {
	DSN               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// DB wraps a pgxpool.Pool and implements db.DB.
type DB struct {
	pool *pgxpool.Pool
}

// Open creates a pool, pings, and returns a ready DB.
func Open(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.DSN == "" {
		return nil, db.WrapOp("postgres", "open", fmt.Errorf("empty DSN"))
	}
	pcfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, db.WrapOp("postgres", "parse", err)
	}
	if cfg.MaxConns > 0 {
		pcfg.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		pcfg.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		pcfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		pcfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.HealthCheckPeriod > 0 {
		pcfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	}
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, db.WrapOp("postgres", "connect", err)
	}
	d := &DB{pool: pool}
	if err := d.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) Ping(ctx context.Context) error {
	if d == nil || d.pool == nil {
		return db.NotOpenError{}
	}
	return db.WrapOp("postgres", "ping", d.pool.Ping(ctx))
}

func (d *DB) Close() error {
	if d == nil || d.pool == nil {
		return nil
	}
	d.pool.Close()
	d.pool = nil
	return nil
}

func (d *DB) BeginTx(ctx context.Context) (db.Tx, error) {
	if d == nil || d.pool == nil {
		return nil, db.NotOpenError{}
	}
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, db.WrapOp("postgres", "begin", err)
	}
	return &Tx{tx: tx}, nil
}

// Native returns the underlying pgx pool (PostgreSQL escape hatch; not on db.DB).
// Use for COPY, LISTEN/NOTIFY, JSONB helpers, or sqlc pgx backends.
func (d *DB) Native() *pgxpool.Pool {
	if d == nil {
		return nil
	}
	return d.pool
}

// Tx wraps pgx.Tx.
type Tx struct {
	tx pgx.Tx
}

func (t *Tx) Commit(ctx context.Context) error {
	if t == nil || t.tx == nil {
		return db.NotOpenError{}
	}
	return t.tx.Commit(ctx)
}

func (t *Tx) Rollback(ctx context.Context) error {
	if t == nil || t.tx == nil {
		return db.NotOpenError{}
	}
	return t.tx.Rollback(ctx)
}

var _ db.DB = (*DB)(nil)
var _ db.Tx = (*Tx)(nil)
