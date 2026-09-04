package database_test

// Link optional SQL drivers so CI live smoke can open mysql/pgsql.
// Production binaries still omit them until db:setup.
import (
	_ "github.com/zatrano/packages/database/driver/mysql"
	_ "github.com/zatrano/packages/database/driver/pgsql"
)
