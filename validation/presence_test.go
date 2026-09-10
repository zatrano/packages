package validation_test

import (
	"errors"
	"testing"

	"github.com/zatrano/packages/validation"
)

func presenceKey(table, column, value string) string {
	return table + "." + column + "=" + value
}

func mapPresence(store map[string]bool) validation.PresenceChecker {
	return func(table, column, value string) (bool, error) {
		return store[presenceKey(table, column, value)], nil
	}
}

func errPresence(err error) validation.PresenceChecker {
	return func(table, column, value string) (bool, error) {
		return false, err
	}
}

func TestUniqueAbsentRecordPasses(t *testing.T) {
	v := validation.Make(map[string]string{"email": "new@example.com"}, map[string]string{
		"email": "unique:users,email",
	})
	v.SetPresenceChecker(mapPresence(nil))
	if v.Fails() {
		t.Fatalf("absent unique value must pass, got %#v", v.Errors())
	}
}

func TestUniqueExistingRecordFails(t *testing.T) {
	v := validation.Make(map[string]string{"email": "taken@example.com"}, map[string]string{
		"email": "unique:users,email",
	})
	v.SetPresenceChecker(mapPresence(map[string]bool{
		presenceKey("users", "email", "taken@example.com"): true,
	}))
	if !v.Fails() || !v.Errors().Has("email") {
		t.Fatalf("existing unique value must fail validation, got %#v", v.Errors())
	}
}

func TestUniqueDatabaseErrorMustNotPass(t *testing.T) {
	v := validation.Make(map[string]string{"email": "a@example.com"}, map[string]string{
		"email": "unique:users,email",
	})
	v.SetPresenceChecker(errPresence(errors.New("connection refused")))
	if v.Passes() {
		t.Fatal("unique must not pass when the database check errors")
	}
}

func TestUniqueCheckerUnavailableMustNotPass(t *testing.T) {
	validation.SetDefaultPresenceChecker(nil)
	t.Cleanup(func() { validation.SetDefaultPresenceChecker(nil) })

	v := validation.Make(map[string]string{"email": "a@example.com"}, map[string]string{
		"email": "unique:users,email",
	})
	if v.Passes() {
		t.Fatal("unique must not pass when no PresenceChecker is bound")
	}
}

func TestExistsExistingRecordPasses(t *testing.T) {
	v := validation.Make(map[string]string{"product_id": "9"}, map[string]string{
		"product_id": "exists:products,id",
	})
	v.SetPresenceChecker(mapPresence(map[string]bool{
		presenceKey("products", "id", "9"): true,
	}))
	if v.Fails() {
		t.Fatalf("existing exists value must pass, got %#v", v.Errors())
	}
}

func TestExistsAbsentRecordFails(t *testing.T) {
	v := validation.Make(map[string]string{"product_id": "404"}, map[string]string{
		"product_id": "exists:products,id",
	})
	v.SetPresenceChecker(mapPresence(nil))
	if !v.Fails() || !v.Errors().Has("product_id") {
		t.Fatalf("missing exists value must fail validation, got %#v", v.Errors())
	}
}

func TestExistsDatabaseErrorMustNotPass(t *testing.T) {
	v := validation.Make(map[string]string{"product_id": "1"}, map[string]string{
		"product_id": "exists:products,id",
	})
	v.SetPresenceChecker(errPresence(errors.New("query failed")))
	if v.Passes() {
		t.Fatal("exists must not pass when the database check errors")
	}
}

func TestExistsCheckerUnavailableMustNotPass(t *testing.T) {
	validation.SetDefaultPresenceChecker(nil)
	t.Cleanup(func() { validation.SetDefaultPresenceChecker(nil) })

	v := validation.Make(map[string]string{"product_id": "1"}, map[string]string{
		"product_id": "exists:products,id",
	})
	if v.Passes() {
		t.Fatal("exists must not pass when no PresenceChecker is bound")
	}
}

func TestPresenceMalformedRuleMustNotPass(t *testing.T) {
	v := validation.Make(map[string]string{"email": "a@example.com"}, map[string]string{
		"email": "unique:users",
	})
	v.SetPresenceChecker(mapPresence(nil))
	if v.Passes() {
		t.Fatal("unique without a column must not pass")
	}
}

func TestPresenceEmptyTableOrColumnMustNotPass(t *testing.T) {
	called := false
	v := validation.Make(map[string]string{"email": "a@example.com"}, map[string]string{
		"email": "unique:,email",
	})
	v.SetPresenceChecker(func(table, column, value string) (bool, error) {
		called = true
		return false, nil
	})
	if v.Passes() {
		t.Fatal("unique with empty table must not pass")
	}
	if called {
		t.Fatal("checker must not run for an empty table/column")
	}
}

func TestPresenceEmptyValueStillSkips(t *testing.T) {
	called := false
	v := validation.Make(map[string]string{"email": ""}, map[string]string{
		"email": "unique:users,email",
	})
	v.SetPresenceChecker(func(table, column, value string) (bool, error) {
		called = true
		return true, nil
	})
	if v.Fails() {
		t.Fatalf("empty unique without required must skip, got %#v", v.Errors())
	}
	if called {
		t.Fatal("checker must not run for an empty value")
	}
}

func TestUniqueExtraCSVPartsStillUseTableAndColumn(t *testing.T) {
	var gotTable, gotColumn, gotValue string
	v := validation.Make(map[string]string{"email": "a@example.com"}, map[string]string{
		"email": "unique:users,email,id,5",
	})
	v.SetPresenceChecker(func(table, column, value string) (bool, error) {
		gotTable, gotColumn, gotValue = table, column, value
		return false, nil
	})
	if v.Fails() {
		t.Fatalf("absent unique with extra CSV parts must still pass when the lookup succeeds, got %#v", v.Errors())
	}
	if gotTable != "users" || gotColumn != "email" || gotValue != "a@example.com" {
		t.Fatalf("extra CSV parts must not change table/column/value: %s %s %s", gotTable, gotColumn, gotValue)
	}
}
