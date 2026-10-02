package apitoken

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/zatrano/packages/db"
)

var safeIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type sqlTable struct {
	db     *sql.DB
	driver string
	table  string
}

func newSQLTable(sqlDB *sql.DB, driver, table string) *sqlTable {
	return &sqlTable{db: sqlDB, driver: db.NormalizeDriver(driver), table: table}
}

func (t *sqlTable) ph(n int) string {
	if t.driver == "pgsql" {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

func (t *sqlTable) mustIdent(name string) error {
	if !safeIdent.MatchString(name) {
		return fmt.Errorf("apitoken: invalid identifier %q", name)
	}
	return nil
}

func (t *sqlTable) queryFirst(col string, val any) (map[string]any, error) {
	if err := t.mustIdent(col); err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteString("SELECT * FROM ")
	b.WriteString(t.table)
	b.WriteString(" WHERE ")
	b.WriteString(col)
	b.WriteString(" = ")
	b.WriteString(t.ph(1))
	b.WriteString(" LIMIT 1")
	rows, err := t.db.Query(b.String(), val)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return scanRow(rows)
}

func (t *sqlTable) getWhere(col string, val any) ([]map[string]any, error) {
	if err := t.mustIdent(col); err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteString("SELECT * FROM ")
	b.WriteString(t.table)
	b.WriteString(" WHERE ")
	b.WriteString(col)
	b.WriteString(" = ")
	b.WriteString(t.ph(1))
	rows, err := t.db.Query(b.String(), val)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		m, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (t *sqlTable) deleteWhere(col string, val any) error {
	if err := t.mustIdent(col); err != nil {
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

func (t *sqlTable) updateWhere(set map[string]any, whereCol string, whereVal any) error {
	if err := t.mustIdent(whereCol); err != nil {
		return err
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		if err := t.mustIdent(k); err != nil {
			return err
		}
		keys = append(keys, k)
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

func (t *sqlTable) insertGetID(attrs map[string]any) (int64, error) {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		if err := t.mustIdent(k); err != nil {
			return 0, err
		}
		keys = append(keys, k)
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
