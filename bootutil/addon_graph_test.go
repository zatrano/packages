package bootutil_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type addonNode struct {
	name     string
	requires []string
	optional []string
}

func TestOfficialAddonGraphHasNoCycles(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	nodes := map[string]*addonNode{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == "testdata" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Register" || len(call.Args) != 1 {
				return true
			}
			lit, ok := call.Args[0].(*ast.CompositeLit)
			if !ok {
				return true
			}
			node := parseMetaLit(lit)
			if node == nil || node.name == "" {
				return true
			}
			nodes[node.name] = node
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) < 8 {
		t.Fatalf("expected many addon metas, got %d", len(nodes))
	}
	if cycle := addonCycle(nodes); cycle != "" {
		t.Fatalf("addon Requires/Optional cycle: %s", cycle)
	}
}

func parseMetaLit(lit *ast.CompositeLit) *addonNode {
	node := &addonNode{}
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "Name":
			if s, ok := stringLit(kv.Value); ok {
				node.name = strings.ToLower(s)
			}
		case "Requires":
			node.requires = stringSliceLit(kv.Value)
		case "Optional":
			node.optional = stringSliceLit(kv.Value)
		}
	}
	if node.name == "" {
		return nil
	}
	return node
}

func stringLit(expr ast.Expr) (string, bool) {
	bl, ok := expr.(*ast.BasicLit)
	if !ok || bl.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(bl.Value)
	if err != nil {
		return "", false
	}
	return s, true
}

func stringSliceLit(expr ast.Expr) []string {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var out []string
	for _, elt := range lit.Elts {
		if s, ok := stringLit(elt); ok && s != "" {
			out = append(out, strings.ToLower(s))
		}
	}
	return out
}

func addonCycle(nodes map[string]*addonNode) string {
	incoming := map[string]int{}
	graph := map[string][]string{}
	for name := range nodes {
		incoming[name] = 0
	}
	addEdge := func(from, to string) {
		if from == "" || to == "" || from == to {
			return
		}
		if _, ok := nodes[from]; !ok {
			return
		}
		if _, ok := nodes[to]; !ok {
			return
		}
		graph[from] = append(graph[from], to)
		incoming[to]++
	}
	for _, n := range nodes {
		for _, req := range n.requires {
			addEdge(req, n.name)
		}
		for _, opt := range n.optional {
			addEdge(opt, n.name)
		}
	}
	var ready []string
	for name, n := range incoming {
		if n == 0 {
			ready = append(ready, name)
		}
	}
	sort.Strings(ready)
	seen := 0
	for len(ready) > 0 {
		name := ready[0]
		ready = ready[1:]
		seen++
		for _, next := range graph[name] {
			incoming[next]--
			if incoming[next] == 0 {
				ready = append(ready, next)
			}
		}
		sort.Strings(ready)
	}
	if seen == len(nodes) {
		return ""
	}
	var leftover []string
	for name, n := range incoming {
		if n > 0 {
			leftover = append(leftover, name)
		}
	}
	sort.Strings(leftover)
	return strings.Join(leftover, ",")
}
