package view

import "testing"

func TestCompileIfInnerLogical(t *testing.T) {
	got, err := compileIfInner("$a && $b")
	if err != nil {
		t.Fatal(err)
	}
	want := "(and (dataGet . `a`) (dataGet . `b`))"
	if got != want {
		t.Fatalf("and: got %q want %q", got, want)
	}

	got, err = compileIfInner("$a || $b")
	if err != nil {
		t.Fatal(err)
	}
	want = "(or (dataGet . `a`) (dataGet . `b`))"
	if got != want {
		t.Fatalf("or: got %q want %q", got, want)
	}

	got, err = compileIfInner("$a || $b && $c")
	if err != nil {
		t.Fatal(err)
	}
	want = "(or (dataGet . `a`) (and (dataGet . `b`) (dataGet . `c`)))"
	if got != want {
		t.Fatalf("precedence: got %q want %q", got, want)
	}

	got, err = compileIfInner("($a || $b) && $c")
	if err != nil {
		t.Fatal(err)
	}
	want = "(and (or (dataGet . `a`) (dataGet . `b`)) (dataGet . `c`))"
	if got != want {
		t.Fatalf("parens: got %q want %q", got, want)
	}

	got, err = compileIfInner("$n > 1 && $ok")
	if err != nil {
		t.Fatal(err)
	}
	want = "(and (cmpGt (dataGet . `n`) 1) (dataGet . `ok`))"
	if got != want {
		t.Fatalf("cmp and: got %q want %q", got, want)
	}

	got, err = compileIfInner("$a + $b")
	if err != nil {
		t.Fatal(err)
	}
	want = "(numAdd (dataGet . `a`) (dataGet . `b`))"
	if got != want {
		t.Fatalf("add: got %q want %q", got, want)
	}

	got, err = compileIfInner("!$a")
	if err != nil {
		t.Fatal(err)
	}
	want = "(not (dataGet . `a`))"
	if got != want {
		t.Fatalf("not: got %q want %q", got, want)
	}

	got, err = compileIfInner("!empty($x)")
	if err != nil {
		t.Fatal(err)
	}
	want = "(not (empty (dataGet . `x`)))"
	if got != want {
		t.Fatalf("!empty: got %q want %q", got, want)
	}

	got, err = compileIfInner("$matrix[$role.ID][$perm.ID]")
	if err != nil {
		t.Fatal(err)
	}
	want = "(dataIndex (dataIndex (dataGet . `matrix`) (dataGet . `role.ID`)) (dataGet . `perm.ID`))"
	if got != want {
		t.Fatalf("index: got %q want %q", got, want)
	}

	got, err = compileIfInner("count($items) > 0")
	if err != nil {
		t.Fatal(err)
	}
	want = "(cmpGt (ifCount (dataGet . `items`)) 0)"
	if got != want {
		t.Fatalf("count: got %q want %q", got, want)
	}

	got, err = compileIfInner("$i % 2 == 0")
	if err != nil {
		t.Fatal(err)
	}
	want = "(eq (printf `%v` (numMod (dataGet . `i`) 2)) (printf `%v` 0))"
	if got != want {
		t.Fatalf("mod: got %q want %q", got, want)
	}

	got, err = compileIfInner("$a ? $b : $c")
	if err != nil {
		t.Fatal(err)
	}
	want = "(ifTernary (dataGet . `a`) (dataGet . `b`) (dataGet . `c`))"
	if got != want {
		t.Fatalf("ternary: got %q want %q", got, want)
	}

	got, err = compileIfInner("$name ?? 'Guest'")
	if err != nil {
		t.Fatal(err)
	}
	want = "(ifCoalesce (dataGet . `name`) `Guest`)"
	if got != want {
		t.Fatalf("coalesce: got %q want %q", got, want)
	}

	got, err = compileIfInner("$a ?: $b")
	if err != nil {
		t.Fatal(err)
	}
	want = "(ifElvis (dataGet . `a`) (dataGet . `b`))"
	if got != want {
		t.Fatalf("elvis: got %q want %q", got, want)
	}

	got, err = compileIfInner("in_array($role, $roles)")
	if err != nil {
		t.Fatal(err)
	}
	want = "(inArray (dataGet . `role`) (dataGet . `roles`))"
	if got != want {
		t.Fatalf("in_array: got %q want %q", got, want)
	}

	if _, err := compileIfInner("$a = $b"); err == nil {
		t.Fatal("expected error for assignment")
	}
}
