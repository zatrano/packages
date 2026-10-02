package db_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Adapter modules must not declare foreign database drivers in their go.mod.
// (Static go.mod check — avoids nested-module replace resolution for framework.)
func TestAdapterDependencyIsolation(t *testing.T) {
	root := dbPackageRoot(t)
	cases := []struct {
		dir     string
		must    []string
		mustNot []string
	}{
		{
			dir:  "postgres",
			must: []string{"github.com/jackc/pgx/v5"},
			mustNot: []string{
				"github.com/go-sql-driver/mysql",
				"modernc.org/sqlite",
				"github.com/microsoft/go-mssqldb",
				"github.com/sijms/go-ora",
			},
		},
		{
			dir:  "mysql",
			must: []string{"github.com/go-sql-driver/mysql"},
			mustNot: []string{
				"github.com/jackc/pgx",
				"modernc.org/sqlite",
				"github.com/microsoft/go-mssqldb",
				"github.com/sijms/go-ora",
			},
		},
		{
			dir:  "mariadb",
			must: []string{"github.com/go-sql-driver/mysql"},
			mustNot: []string{
				"github.com/jackc/pgx",
				"modernc.org/sqlite",
				"github.com/microsoft/go-mssqldb",
				"github.com/sijms/go-ora",
			},
		},
		{
			dir:  "sqlite",
			must: []string{"modernc.org/sqlite"},
			mustNot: []string{
				"github.com/jackc/pgx",
				"github.com/go-sql-driver/mysql",
				"github.com/microsoft/go-mssqldb",
				"github.com/sijms/go-ora",
			},
		},
		{
			dir:  "sqlserver",
			must: []string{"github.com/microsoft/go-mssqldb"},
			mustNot: []string{
				"github.com/jackc/pgx",
				"github.com/go-sql-driver/mysql",
				"modernc.org/sqlite",
				"github.com/sijms/go-ora",
			},
		},
		{
			dir:  "oracle",
			must: []string{"github.com/sijms/go-ora"},
			mustNot: []string{
				"github.com/jackc/pgx",
				"github.com/go-sql-driver/mysql",
				"modernc.org/sqlite",
				"github.com/microsoft/go-mssqldb",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			modPath := filepath.Join(root, tc.dir, "go.mod")
			raw, err := os.ReadFile(modPath)
			if err != nil {
				t.Fatal(err)
			}
			text := string(raw)
			// Also scan go.sum if present (transitive pins).
			if sum, err := os.ReadFile(filepath.Join(root, tc.dir, "go.sum")); err == nil {
				text += "\n" + string(sum)
			}
			for _, want := range tc.must {
				if !strings.Contains(text, want) {
					t.Fatalf("expected %q in %s go.mod/go.sum", want, tc.dir)
				}
			}
			for _, bad := range tc.mustNot {
				if strings.Contains(text, bad) {
					t.Fatalf("foreign driver %q leaked into %s go.mod/go.sum", bad, tc.dir)
				}
			}
		})
	}
}

func TestNoCockroachAdapterYet(t *testing.T) {
	root := dbPackageRoot(t)
	if _, err := os.Stat(filepath.Join(root, "cockroachdb")); err == nil {
		t.Fatal("cockroachdb adapter must stay FUTURE — do not ship in v3 initial matrix")
	}
}

func TestMariaDBIsOwnModule(t *testing.T) {
	root := dbPackageRoot(t)
	mysqlMod, err := os.ReadFile(filepath.Join(root, "mysql", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	mariaMod, err := os.ReadFile(filepath.Join(root, "mariadb", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mysqlMod), "module github.com/zatrano/packages/db/mysql") {
		t.Fatal("mysql module path")
	}
	if !strings.Contains(string(mariaMod), "module github.com/zatrano/packages/db/mariadb") {
		t.Fatal("mariadb must be its own module (first-class, not MySQL alias)")
	}
}

func dbPackageRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Dir(file)
}
