// Package mysql is the MySQL first-class adapter for ZATRANO V3 db.
//
// Importing this package pulls go-sql-driver/mysql only — not PostgreSQL,
// SQLite, SQL Server, or Oracle drivers.
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/zatrano/packages/db"
)

// Config is MySQL-specific configuration.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DB wraps database/sql for MySQL.
type DB struct {
	sql *sql.DB
}

// Open opens MySQL, applies pool settings, pings, and returns a ready DB.
func Open(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.DSN == "" {
		return nil, db.WrapOp("mysql", "open", fmt.Errorf("empty DSN"))
	}
	sqlDB, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, db.WrapOp("mysql", "open", err)
	}
	applyPool(sqlDB, cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime, cfg.ConnMaxIdleTime)
	d := &DB{sql: sqlDB}
	if err := d.Ping(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return d, nil
}

func applyPool(sqlDB *sql.DB, maxOpen, maxIdle int, maxLife, maxIdleTime time.Duration) {
	if maxOpen > 0 {
		sqlDB.SetMaxOpenConns(maxOpen)
	}
	if maxIdle > 0 {
		sqlDB.SetMaxIdleConns(maxIdle)
	}
	if maxLife > 0 {
		sqlDB.SetConnMaxLifetime(maxLife)
	}
	if maxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(maxIdleTime)
	}
}

func (d *DB) Ping(ctx context.Context) error {
	if d == nil || d.sql == nil {
		return db.NotOpenError{}
	}
	return db.WrapOp("mysql", "ping", d.sql.PingContext(ctx))
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
		return nil, db.WrapOp("mysql", "begin", err)
	}
	return &Tx{tx: tx}, nil
}

// Native returns *sql.DB for database/sql or sqlc mysql backends.
func (d *DB) Native() *sql.DB {
	if d == nil {
		return nil
	}
	return d.sql
}

type Tx struct {
	tx *sql.Tx
}

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
