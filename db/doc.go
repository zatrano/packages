// Package db is the ZATRANO V3 SQL-first database infrastructure core.
//
// # What this is
//
// Connection / pool / transaction / lifecycle contracts and helpers. Adapters
// live in sibling modules (db/postgres, db/mysql, …) and pull only their own
// native driver into the dependency graph.
//
// # What this is not
//
// ORM, query builder, repository, migration engine, seeder engine, or sqlc.
// Application SQL lives under the app tree:
//
//	database/migrations/
//	database/queries/
//	database/sqlc/
//	database/seeders/
//
// # Binding model
//
//	import "github.com/zatrano/packages/db/postgres"
//	db, err := postgres.Open(ctx, postgres.Config{DSN: os.Getenv("DATABASE_URL")})
//
// There is no database.Open("postgres") registry and no blank-import of all drivers.
//
// # CockroachDB
//
// Future adapter only — not implemented in v3 initial matrix.
package db
