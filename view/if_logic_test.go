package view_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zatrano/packages/view"
)

func TestIfLogicalAndOr(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "logic.html"), []byte(`
@if($a && $b)
<span class="and">both</span>
@endif
@if($a || $b)
<span class="or">any</span>
@endif
@if($n > 1 && $ok)
<span class="mix">mix</span>
@endif
@if($a || $b && $c)
<span class="prec">prec</span>
@endif
@if(($a || $b) && $c)
<span class="paren">paren</span>
@endif
@if($status == 'draft' || $status == 'pending')
<span class="st">open</span>
@elseif($status == 'ok' && $ok)
<span class="st">ready</span>
@else
<span class="st">other</span>
@endif
`), 0o644)

	engine := view.New(dir)

	out, err := engine.Render("logic", map[string]any{
		"a": true, "b": true, "c": false, "n": 3, "ok": true, "status": "ok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `class="and">both`) || !strings.Contains(out, `class="or">any`) {
		t.Fatalf("true&&true / true||true missing: %s", out)
	}
	if !strings.Contains(out, `class="mix">mix`) {
		t.Fatalf("n>1 && ok missing: %s", out)
	}
	// a || (b && c) with a=true, b=true, c=false → true
	if !strings.Contains(out, `class="prec">prec`) {
		t.Fatalf("|| vs && precedence missing: %s", out)
	}
	// (a || b) && c with c=false → false
	if strings.Contains(out, `class="paren">paren`) {
		t.Fatalf("parenthesized && should be false: %s", out)
	}
	if !strings.Contains(out, `class="st">ready`) {
		t.Fatalf("elseif && branch missing: %s", out)
	}

	out, err = engine.Render("logic", map[string]any{
		"a": false, "b": false, "c": true, "n": 1, "ok": true, "status": "draft",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, `class="and">both`) {
		t.Fatalf("false&&false should hide and: %s", out)
	}
	if strings.Contains(out, `class="or">any`) {
		t.Fatalf("false||false should hide or: %s", out)
	}
	if strings.Contains(out, `class="mix">mix`) {
		t.Fatalf("n>1 && ok should hide mix: %s", out)
	}
	// a || (b && c) with a=false, b=false, c=true → false
	if strings.Contains(out, `class="prec">prec`) {
		t.Fatalf("false || (false && true) should hide prec: %s", out)
	}
	// (false || false) && true → false
	if strings.Contains(out, `class="paren">paren`) {
		t.Fatalf("(false||false)&&true should hide paren: %s", out)
	}
	if !strings.Contains(out, `class="st">open`) {
		t.Fatalf("status draft||pending missing: %s", out)
	}

	out, err = engine.Render("logic", map[string]any{
		"a": false, "b": true, "c": true, "n": 2, "ok": false, "status": "pending",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, `class="and">both`) {
		t.Fatalf("false&&true should hide and: %s", out)
	}
	if !strings.Contains(out, `class="or">any`) {
		t.Fatalf("false||true missing: %s", out)
	}
	if strings.Contains(out, `class="mix">mix`) {
		t.Fatalf("n>1 && !ok should hide mix: %s", out)
	}
	// false || (true && true) → true
	if !strings.Contains(out, `class="prec">prec`) {
		t.Fatalf("false || (true && true) missing: %s", out)
	}
	// (false || true) && true → true
	if !strings.Contains(out, `class="paren">paren`) {
		t.Fatalf("parenthesized or&& missing: %s", out)
	}
	if !strings.Contains(out, `class="st">open`) {
		t.Fatalf("pending should take || branch: %s", out)
	}
}

func TestIfLogicalInsideForeach(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "nav.html"), []byte(`
@foreach($nav as $link)
@if($link.active && $show)
<a class="shown-active">{{ $link.title }}</a>
@elseif($link.active || $force)
<a class="shown">{{ $link.title }}</a>
@endif
@endforeach
`), 0o644)
	engine := view.New(dir)
	out, err := engine.Render("nav", map[string]any{
		"show":  true,
		"force": false,
		"nav": []map[string]any{
			{"title": "A", "active": true},
			{"title": "B", "active": false},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `class="shown-active">A`) {
		t.Fatalf("foreach && missing: %s", out)
	}
	if strings.Contains(out, `>B<`) || strings.Contains(out, `shown">B`) {
		t.Fatalf("inactive link should stay hidden: %s", out)
	}
}

func TestIfLogicalInsideHTMLAttribute(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "btn.html"), []byte(
		`<a class="@if($locked && $warn)danger@endif" href="@if($ok || $draft)/go@else/stop@endif">x</a>`,
	), 0o644)
	e := view.New(dir)
	out, err := e.Render("btn", map[string]any{"locked": true, "warn": true, "ok": false, "draft": true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `class="danger"`) {
		t.Fatalf("class=%q", out)
	}
	if !strings.Contains(out, `href="/go"`) {
		t.Fatalf("href=%q", out)
	}
}

func TestIfNegationArithmeticHelpersAndIndex(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "ops.html"), []byte(`
@if(!$hidden)
<span class="shown">yes</span>
@endif
@if(not $hidden)
<span class="shown-word">yes</span>
@endif
@if($a + $b)
<span class="sum">yes</span>
@endif
@if($a + $b > 2)
<span class="sumgt">yes</span>
@endif
@if($i % 2 == 0)
<span class="even">yes</span>
@endif
@if(!empty($name))
<span class="name">{{ $name }}</span>
@endif
@if(isset($missing))
<span class="missing">no</span>
@endif
@if(count($items) > 0)
<span class="n">ok</span>
@endif
@if($ok === true)
<span class="bool">yes</span>
@endif
@if($a and $b)
<span class="andw">yes</span>
@endif
@if($off or $ok)
<span class="orw">yes</span>
@endif
@if($n > -1)
<span class="neg">yes</span>
@endif
`), 0o644)
	engine := view.New(dir)
	out, err := engine.Render("ops", map[string]any{
		"hidden":  false,
		"a":       2,
		"b":       3,
		"i":       4,
		"name":    "Ada",
		"ok":      true,
		"off":     false,
		"n":       0,
		"items":   []string{"x", "y"},
		"missing": nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`class="shown">yes`,
		`class="shown-word">yes`,
		`class="sum">yes`,
		`class="sumgt">yes`,
		`class="even">yes`,
		`class="name">Ada`,
		`class="bool">yes`,
		`class="andw">yes`,
		`class="orw">yes`,
		`class="neg">yes`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
	if strings.Contains(out, `class="missing"`) {
		t.Fatalf("isset(nil) should be false: %s", out)
	}
}

func TestIfCountDirectiveDoesNotLeak(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "c.html"), []byte(`
@if(count($items) > 0)
<span class="n">ok</span>
@endif
@if(empty($blank))
<span class="blank">empty</span>
@endif
`), 0o644)
	engine := view.New(dir)
	out, err := engine.Render("c", map[string]any{"items": []int{1, 2, 3}, "blank": ""})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `class="n">ok`) || !strings.Contains(out, `class="blank">empty`) {
		t.Fatalf("count/empty failed: %s", out)
	}
}

func TestIfMatrixIndexInsideForeach(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "matrix.html"), []byte(`
@foreach($permissions as $perm)
@foreach($roles as $role)
@if($matrix[$role.ID][$perm.ID])
<span class="grant">{{ $role.ID }}-{{ $perm.ID }}</span>
@endif
@endforeach
@endforeach
`), 0o644)
	engine := view.New(dir)
	out, err := engine.Render("matrix", map[string]any{
		"roles":       []map[string]any{{"ID": 1}, {"ID": 2}},
		"permissions": []map[string]any{{"ID": 10}, {"ID": 11}},
		"matrix": map[int]map[int]bool{
			1: {10: true},
			2: {11: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `class="grant">1-10`) || !strings.Contains(out, `class="grant">2-11`) {
		t.Fatalf("matrix grants missing: %s", out)
	}
	if strings.Contains(out, `1-11`) || strings.Contains(out, `2-10`) {
		t.Fatalf("unexpected grant: %s", out)
	}
}

func TestIfTernaryCoalesceAndEcho(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "t.html"), []byte(`
@if($on ? $ok : $off)
<span class="tern">yes</span>
@endif
@if($missing ?? $ok)
<span class="coal">yes</span>
@endif
@if($off ?: $ok)
<span class="elvis">yes</span>
@endif
@if(in_array('admin', $roles))
<span class="role">yes</span>
@endif
@if(str_contains($email, '@'))
<span class="at">yes</span>
@endif
@if(filled($email) && blank($empty))
<span class="fill">yes</span>
@endif
<p class="echo">{{ $on ? 'Y' : 'N' }}</p>
<p class="guest">{{ $missing ?? 'Guest' }}</p>
<p class="full">{{ $first . ' ' . $last }}</p>
<p class="raw">{!! $on ? '<b>ok</b>' : 'no' !!}</p>
`), 0o644)
	engine := view.New(dir)
	out, err := engine.Render("t", map[string]any{
		"on":      true,
		"ok":      true,
		"off":     false,
		"roles":   []string{"user", "admin"},
		"email":   "a@b.c",
		"empty":   "  ",
		"first":   "Ada",
		"last":    "Lovelace",
		"missing": nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`class="tern">yes`,
		`class="coal">yes`,
		`class="elvis">yes`,
		`class="role">yes`,
		`class="at">yes`,
		`class="fill">yes`,
		`class="echo">Y`,
		`class="guest">Guest`,
		`class="full">Ada Lovelace`,
		`<b>ok</b>`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
}
