package view

import (
	"fmt"
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/bootutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zatrano/framework/kernel"
)

func Commands(app contracts.App) []addons.CLICommand {
	k := bootutil.KernelApp(app)
	if k == nil {
		return nil
	}
	return bootutil.CLI(
		&MakeViewCommand{app: k},
		&MakeComponentCommand{app: k},
	)
}

type MakeViewCommand struct {
	app *kernel.Application
}

func (c *MakeViewCommand) Name() string        { return "make:view" }
func (c *MakeViewCommand) Description() string { return "Create a view template scaffold" }
func (c *MakeViewCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("view name required")
	}
	layout := "app"
	layoutSet := false
	var nameArg string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--layout=") {
			layout = strings.TrimSpace(strings.TrimPrefix(arg, "--layout="))
			layoutSet = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if nameArg == "" {
			nameArg = arg
		}
	}
	if nameArg == "" {
		return fmt.Errorf("view name required")
	}
	if layout == "" {
		layout = "app"
	}
	name := strings.ReplaceAll(nameArg, "\\", "/")
	name = strings.Trim(name, "/")
	name = strings.TrimSuffix(name, ".html")
	parts := strings.Split(name, ".")
	for i, part := range parts {
		parts[i] = bootutil.ToSnake(bootutil.ToExported(part))
		parts[i] = strings.ReplaceAll(parts[i], "_", "-")
		if parts[i] == "" {
			parts[i] = strings.ToLower(part)
		}
	}
	rel := strings.Join(parts, string(os.PathSeparator))
	dir := filepath.Join(kernel.ViewsDirForCreate(c.app), filepath.Dir(rel))
	if filepath.Dir(rel) == "." {
		dir = kernel.ViewsDirForCreate(c.app)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	base := filepath.Base(rel)
	path := filepath.Join(dir, base+".html")
	isAuth := strings.HasPrefix(strings.ToLower(strings.ReplaceAll(name, "\\", "/")), "auth/") ||
		strings.HasPrefix(strings.ToLower(name), "auth.") ||
		strings.EqualFold(parts[0], "auth")
	if isAuth && !layoutSet {
		layout = "auth"
	}
	title := bootutil.ToExported(strings.ReplaceAll(base, "-", " "))
	var content string
	if isAuth {
		content = fmt.Sprintf(`@extends('layouts.%s')

@section('title', '%s')

@section('content')
  <h1>%s</h1>
@endsection
`, layout, base, title)
	} else {
		content = fmt.Sprintf(`@extends('layouts.%s')

@section('title', '%s')

@section('content')
  <h1>%s</h1>
  <form method="POST" action="/">
    @csrf
  </form>
@endsection
`, layout, base, title)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("View created: %s\n", path)
	return nil
}
