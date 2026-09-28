# ZATRANO V3 Database (`packages/db`)

SQL-first infrastructure. **Not an ORM.** No runtime driver registry.

```
import "github.com/zatrano/packages/db/postgres"

db, err := postgres.Open(ctx, postgres.Config{DSN: os.Getenv("DATABASE_URL")})
```

| Adapter | Import path | Driver |
|---------|-------------|--------|
| PostgreSQL | `db/postgres` | pgx/v5 |
| MySQL | `db/mysql` | go-sql-driver/mysql |
| MariaDB | `db/mariadb` | go-sql-driver/mysql (separate adapter) |
| SQLite | `db/sqlite` | modernc.org/sqlite |
| SQL Server | `db/sqlserver` | go-mssqldb |
| Oracle | `db/oracle` | go-ora/v2 |

Application SQL lives in the app tree:

```
database/
  migrations/
  queries/
  sqlc/
  seeders/
```

Legacy `packages/database` is superseded; do not extend it for V3.
