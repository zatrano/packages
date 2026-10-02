package auth

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/zatrano/packages/db"
)

var safeIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// sqlTable is a tiny dialect-aware helper for auth row access (no query builder).
type sqlTable struct {
	db     *sql.DB
	driver string
	table  string
}

func newSQLTable(sqlDB *sql.DB, driver, table string) (*sqlTable, error) {
	table = strings.TrimSpace(table)
	if sqlDB == nil {
		return nil, fmt.Errorf("auth: nil database")
	}
	if err := requireIdent(table); err != nil {
		return nil, err
	}
	return &sqlTable{db: sqlDB, driver: db.NormalizeDriver(driver), table: table}, nil
}

func requireIdent(name string) error {
	if !safeIdent.MatchString(name) {
		return fmt.Errorf("auth: invalid identifier %q", name)
	}
	return nil
}

func (t *sqlTable) ph(n int) string {
	if t.driver == "pgsql" {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

func (t *sqlTable) queryFirst(cols []string, vals []any) (map[string]any, error) {
	if len(cols) == 0 || len(cols) != len(vals) {
		return nil, fmt.Errorf("auth: where columns/values mismatch")
	}
	for _, c := range cols {
		if err := requireIdent(c); err != nil {
			return nil, err
		}
	}
	var b strings.Builder
	b.WriteString("SELECT * FROM ")
	b.WriteString(t.table)
	b.WriteString(" WHERE ")
	args := make([]any, 0, len(vals))
	for i, c := range cols {
		if i > 0 {
			b.WriteString(" AND ")
		}
		b.WriteString(c)
		b.WriteString(" = ")
		b.WriteString(t.ph(i + 1))
		args = append(args, vals[i])
	}
	b.WriteString(" LIMIT 1")
	rows, err := t.db.Query(b.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return scanRow(rows)
}

func (t *sqlTable) updateWhere(set map[string]any, whereCol string, whereVal any) error {
	if err := requireIdent(whereCol); err != nil {
		return err
	}
	if len(set) == 0 {
		return nil
	}
	keys := sortedKeys(set)
	for _, k := range keys {
		if err := requireIdent(k); err != nil {
			return err
		}
	}
	var b strings.Builder
	b.WriteString("UPDATE ")
	b.WriteString(t.table)
	b.WriteString(" SET ")
	args := make([]any, 0, len(keys)+1)
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k)
		b.WriteString(" = ")
		b.WriteString(t.ph(i + 1))
		args = append(args, set[k])
	}
	b.WriteString(" WHERE ")
	b.WriteString(whereCol)
	b.WriteString(" = ")
	b.WriteString(t.ph(len(keys) + 1))
	args = append(args, whereVal)
	_, err := t.db.Exec(b.String(), args...)
	return err
}

func (t *sqlTable) deleteWhere(col string, val any) error {
	if err := requireIdent(col); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("DELETE FROM ")
	b.WriteString(t.table)
	b.WriteString(" WHERE ")
	b.WriteString(col)
	b.WriteString(" = ")
	b.WriteString(t.ph(1))
	_, err := t.db.Exec(b.String(), val)
	return err
}

func (t *sqlTable) insert(attrs map[string]any) error {
	if len(attrs) == 0 {
		return fmt.Errorf("auth: empty insert")
	}
	keys := sortedKeys(attrs)
	for _, k := range keys {
		if err := requireIdent(k); err != nil {
			return err
		}
	}
	var b strings.Builder
	b.WriteString("INSERT INTO ")
	b.WriteString(t.table)
	b.WriteString(" (")
	b.WriteString(strings.Join(keys, ", "))
	b.WriteString(") VALUES (")
	args := make([]any, 0, len(keys))
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(t.ph(i + 1))
		args = append(args, attrs[k])
	}
	b.WriteString(")")
	_, err := t.db.Exec(b.String(), args...)
	return err
}

func (t *sqlTable) insertGetID(attrs map[string]any) (int64, error) {
	if len(attrs) == 0 {
		return 0, fmt.Errorf("auth: empty insert")
	}
	keys := sortedKeys(attrs)
	for _, k := range keys {
		if err := requireIdent(k); err != nil {
			return 0, err
		}
	}
	var b strings.Builder
	b.WriteString("INSERT INTO ")
	b.WriteString(t.table)
	b.WriteString(" (")
	b.WriteString(strings.Join(keys, ", "))
	b.WriteString(") VALUES (")
	args := make([]any, 0, len(keys))
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(t.ph(i + 1))
		args = append(args, attrs[k])
	}
	b.WriteString(")")
	if t.driver == "pgsql" {
		b.WriteString(" RETURNING id")
		var id int64
		err := t.db.QueryRow(b.String(), args...).Scan(&id)
		return id, err
	}
	res, err := t.db.Exec(b.String(), args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Stable enough for SQL; sort for deterministic tests.
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func scanRow(rows *sql.Rows) (map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	raw := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range raw {
		ptrs[i] = &raw[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return nil, err
	}
	out := make(map[string]any, len(cols))
	for i, c := range cols {
		v := raw[i]
		if b, ok := v.([]byte); ok {
			out[c] = string(b)
		} else {
			out[c] = v
		}
	}
	return out, nil
}
