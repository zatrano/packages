package sqlserver

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"

	"github.com/zatrano/packages/db"
)

// Config is SQL Server configuration.
type Config struct {
	DSN string
}

// DB wraps database/sql for Microsoft SQL Server.
type DB struct {
	sql *sql.DB
}

func Open(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("sqlserver: empty DSN")
	}
	sqlDB, err := sql.Open("sqlserver", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("sqlserver: open: %w", err)
	}
	d := &DB{sql: sqlDB}
	if err := d.Ping(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) Ping(ctx context.Context) error {
	if d == nil || d.sql == nil {
		return db.NotOpenError{}
	}
	return d.sql.PingContext(ctx)
}

func (d *DB) Close() error {
	if d == nil || d.sql == nil {
		return nil
	}
	err := d.sql.Close()
	d.sql = nil
	return err
}

func (d *DB) BeginTx(ctx context.Context) (db.Tx, error) {
	if d == nil || d.sql == nil {
		return nil, db.NotOpenError{}
	}
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

func (d *DB) Native() *sql.DB {
	if d == nil {
		return nil
	}
	return d.sql
}

type Tx struct{ tx *sql.Tx }

func (t *Tx) Commit(ctx context.Context) error {
	if t == nil || t.tx == nil {
		return db.NotOpenError{}
	}
	return t.tx.Commit()
}

func (t *Tx) Rollback(ctx context.Context) error {
	if t == nil || t.tx == nil {
		return db.NotOpenError{}
	}
	return t.tx.Rollback()
}

var _ db.DB = (*DB)(nil)
var _ db.Tx = (*Tx)(nil)
