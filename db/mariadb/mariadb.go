// Package mariadb is the MariaDB first-class adapter for ZATRANO V3 db.
//
// MariaDB is not a MySQL alias: it has its own package path, Config, tests, and
// future room for MariaDB-specific features. The wire protocol may use the
// MySQL Go driver today, but applications must import this package explicitly.
package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/zatrano/packages/db"
)

// Config is MariaDB-specific configuration.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DB wraps database/sql for MariaDB.
type DB struct {
	sql *sql.DB
}

// Open opens MariaDB, applies pool settings, pings, and returns a ready DB.
func Open(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.DSN == "" {
		return nil, db.WrapOp("mariadb", "open", fmt.Errorf("empty DSN"))
	}
	// Driver name "mysql" is the Go connector protocol name; this adapter is
	// still MariaDB-first-class (separate module + test matrix).
	sqlDB, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, db.WrapOp("mariadb", "open", err)
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
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
	return db.WrapOp("mariadb", "ping", d.sql.PingContext(ctx))
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
		return nil, db.WrapOp("mariadb", "begin", err)
	}
	return &Tx{tx: tx}, nil
}

// Native returns *sql.DB.
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
