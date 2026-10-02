# ZATRANO V3 Database (`packages/db`)

SQL-first **infrastructure**. Not an ORM. Not a query builder. Not a repository layer.

**Spec lock:** workspace [`DATABASE-SPEC.md`](../../DATABASE-SPEC.md).

## First-class matrix

| Database | Import | Driver | Notes |
|----------|--------|--------|-------|
| PostgreSQL | `db/postgres` | pgx/v5 (+ pool) | `Native() *pgxpool.Pool` |
| MySQL | `db/mysql` | go-sql-driver/mysql | `Native() *sql.DB` |
| MariaDB | `db/mariadb` | MySQL protocol driver | **Separate** adapter + tests (not a MySQL alias) |
| SQLite | `db/sqlite` | modernc.org/sqlite | CGO-free |
| SQL Server | `db/sqlserver` | go-mssqldb | |
| Oracle | `db/oracle` | go-ora/v2 | |
| CockroachDB | — | — | **FUTURE** only |

First-class ≠ same SQL everywhere. Keep RETURNING / JSONB / vendor features via adapter `Native()` and application sqlc dialects.

## Binding (compile-time)

```go
import "github.com/zatrano/packages/db/postgres"

db, err := postgres.Open(ctx, postgres.Config{
    DSN:      os.Getenv("DATABASE_URL"),
    MaxConns: 16,
})
defer db.Close()

// Generic lifecycle
_ = db.Ping(ctx)
tx, _ := db.BeginTx(ctx)
_ = tx.Commit(ctx)

// Escape hatch (PostgreSQL only — not on core db.DB)
pool := db.Native()
```

**Forbidden:** `database.Open("postgres")`, blank-import of all drivers, global registry.

Each adapter is its **own Go module** so unused drivers stay out of your `go.mod`.

## Core (`github.com/zatrano/packages/db`)

```go
type DB interface {
    Ping(ctx context.Context) error
    Close() error
    BeginTx(ctx context.Context) (Tx, error)
}
```

Vendor types never appear on core. Errors: `NotOpenError`, `OpError` / `WrapOp`, `IsNotOpen`.

Optional container helper for other packages: `db.SQLFrom(app)` → `*sql.DB` when bound as `"db"`.

## Application layout

```
database/
  migrations/   # schema SQL
  queries/      # hand-written SQL for sqlc
  sqlc/         # generated — do not edit
  seeders/      # demo / reference data
```

sqlc is a **codegen tool**, not a runtime dependency of `packages/db`. See [`examples/users/`](examples/users/).

## Testing

```bash
# Core + isolation (from packages/)
set GOWORK=off
go test ./db/ -count=1

# SQLite lifecycle (nested module)
cd db/sqlite && go test . -count=1

# Optional live DBs (see docker-compose.db-matrix.yml)
set ZATRANO_TEST_POSTGRES_DSN=postgres://zatrano:secret@127.0.0.1:5432/zatrano?sslmode=disable
set ZATRANO_TEST_MYSQL_DSN=zatrano:secret@tcp(127.0.0.1:3306)/zatrano
set ZATRANO_TEST_MARIADB_DSN=zatrano:secret@tcp(127.0.0.1:3307)/zatrano
set ZATRANO_TEST_SQLSERVER_DSN=sqlserver://sa:Your_strong_Password123@127.0.0.1:1433?database=master
set ZATRANO_TEST_ORACLE_DSN=...
```

```bash
docker compose -f db/docker-compose.db-matrix.yml up -d
```

Isolation tests fail if an adapter `go list -m all` graph contains a foreign driver.

## Legacy

`packages/database` and `packages/orm` were **removed** from the V3 workspace. Use this package only.
