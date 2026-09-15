package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/bootutil"
)

// MakePanelCommand scaffolds a named HTML HTTP surface (controller, routes, views/layouts).
type MakePanelCommand struct {
	app contracts.App
}

func (c *MakePanelCommand) Name() string { return "make:panel" }
func (c *MakePanelCommand) Description() string {
	return "Scaffold a named HTTP surface: make:panel dashboard"
}

func (c *MakePanelCommand) Handle(args []string) error {
	force := false
	var rawName string
	for _, arg := range args {
		switch {
		case arg == "--force" || arg == "-f":
			force = true
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown flag %s (usage: make:panel {name})", arg)
		default:
			if rawName != "" {
				return fmt.Errorf("make:panel accepts one name")
			}
			rawName = arg
		}
	}
	name, err := panelName(rawName)
	if err != nil {
		return err
	}
	exported := bootutil.ToExported(name)
	mod := bootutil.ConsumerModule(c.app)

	type filePair struct {
		dest    []string
		content string
	}
	pairs := []filePair{
		{[]string{"app", "http", "controllers", name, "home_controller.go"}, panelControllerStub(name, exported)},
		{[]string{"app", "routes", name, name + ".go"}, panelRouteStub(mod, name, exported)},
		{[]string{"views", name, "layouts", name + ".html"}, panelLayoutHTML()},
		{[]string{"views", name, "index.html"}, panelIndexHTML(name)},
		{[]string{"lang", "en", name + ".json"}, fmt.Sprintf("{\n  \"home\": %q\n}\n", exported)},
		{[]string{"lang", "tr", name + ".json"}, fmt.Sprintf("{\n  \"home\": %q\n}\n", exported)},
	}

	created, skipped := 0, 0
	for _, pair := range pairs {
		dst := bootutil.ScaffoldDest(c.app, pair.dest)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if !force {
			if _, err := os.Stat(dst); err == nil {
				fmt.Printf("Skipped (exists): %s\n", dst)
				skipped++
				continue
			}
		}
		body := pair.content
		if strings.HasSuffix(dst, ".go") {
			body = bootutil.ApplyConsumerPlaceholders(c.app, body)
			if conflict, err := goStubConflicts(body, filepath.Dir(dst), dst); err != nil {
				return err
			} else if conflict != "" {
				fmt.Printf("Skipped (declared): %s — %s\n", dst, conflict)
				skipped++
				continue
			}
		}
		if err := os.WriteFile(dst, []byte(body), 0o644); err != nil {
			return err
		}
		fmt.Printf("Created: %s\n", dst)
		created++
	}

	provider := c.app.BasePath("app", "providers", "route_service_provider.go")
	_ = bootutil.EnsureBlankImport(provider, mod+"/app/routes/"+name)

	fmt.Printf("\nPanel %q ready (%d created, %d skipped).\n", name, created, skipped)
	fmt.Println("Use --force to overwrite existing files.")
	return nil
}

func panelName(raw string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if name == "" {
		return "", fmt.Errorf("panel name required (example: make:panel dashboard)")
	}
	switch name {
	case "web", "api", "auth":
		return "", fmt.Errorf("%q is a built-in HTTP surface; choose another panel name", name)
	}
	if len(name) > 64 {
		return "", fmt.Errorf("panel name too long")
	}
	for i, r := range name {
		if i == 0 {
			if r < 'a' || r > 'z' {
				return "", fmt.Errorf("panel name must start with a letter")
			}
			continue
		}
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			if unicode.IsSpace(r) || r == '-' {
				return "", fmt.Errorf("panel name must be a Go package identifier (letters and digits)")
			}
			return "", fmt.Errorf("invalid panel name %q", name)
		}
	}
	return name, nil
}

func panelControllerStub(pkg, exported string) string {
	return `package ` + pkg + `

import "github.com/zatrano/framework/v2/kernel/http"

type HomeController struct{}

func (c *HomeController) Index(req *http.Request) *http.Response {
	return http.View("` + pkg + `/index", map[string]any{})
}
`
}

func panelRouteStub(mod, pkg, exported string) string {
	return `package ` + pkg + `

import (
	approutes "` + `__MODULE__/app/routes` + `"
	` + pkg + `ctrl "` + `__MODULE__/app/http/controllers/` + pkg + `"

	pkgauth "github.com/zatrano/packages/auth"
	"github.com/zatrano/framework/v2/kernel/routing"
)

func init() {
	routing.RegisterWeb(register` + exported + `)
}

func register` + exported + `(router *routing.Router) {
	app := approutes.Current()
	if app == nil || router == nil {
		return
	}
	a := pkgauth.From(app)
	router.Group("", func(r *routing.Router) {
		routing.Controller(r, &` + pkg + `ctrl.HomeController{}, func(rr routing.RouteRegistrar, c *` + pkg + `ctrl.HomeController) {
			rr.Get("/` + pkg + `", c.Index).As("` + pkg + `.home")
		})
	}, pkgauth.Middleware(a), pkgauth.VerifyEmailMiddleware(a))
}
`
}

func panelLayoutHTML() string {
	return `<!DOCTYPE html>
<html lang="{{ $locale }}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>@yield('title')</title>
    @yield('head')
</head>
<body>
    @yield('content')
</body>
</html>
`
}

func panelIndexHTML(pkg string) string {
	return `@extends('` + pkg + `.layouts.` + pkg + `')

@section('title')
@lang('` + pkg + `.home')
@endsection

@section('content')
<h1>@lang('` + pkg + `.home')</h1>
@endsection
`
}
