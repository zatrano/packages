package bootutil

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/layout"
)

// NamedCmd is a console command that can be registered through addon Meta.CLI.
type NamedCmd interface {
	Name() string
	Description() string
	Handle(args []string) error
}

// CLI wraps named commands for addon registration.
func CLI(cmds ...NamedCmd) []addons.CLICommand {
	out := make([]addons.CLICommand, 0, len(cmds))
	for _, c := range cmds {
		cmd := c
		out = append(out, addons.CLICommand{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Handle:      cmd.Handle,
		})
	}
	return out
}

// ConsumerModule returns the consuming application's Go module path.
func ConsumerModule(app contracts.App) string {
	if app == nil {
		return "your/module"
	}
	mod, err := modulePath(app.BasePath())
	if err != nil || strings.TrimSpace(mod) == "" || mod == "github.com/zatrano/framework" {
		return "your/module"
	}
	return mod
}

// ApplyConsumerPlaceholders rewrites generated stub module paths.
func ApplyConsumerPlaceholders(app contracts.App, body string) string {
	mod := ConsumerModule(app)
	body = strings.ReplaceAll(body, "__MODULE__", mod)
	body = strings.ReplaceAll(body, "github.com/zatrano/framework/app/", mod+"/app/")
	return body
}

// ScaffoldDest maps starter layout prefixes onto app/views, app/localization, app/database.
func ScaffoldDest(app contracts.App, parts []string) string {
	if app == nil || len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "views":
		return filepath.Join(append([]string{layout.ViewsDirForCreate(app)}, parts[1:]...)...)
	case "lang":
		return filepath.Join(append([]string{layout.LocalizationDirForCreate(app)}, parts[1:]...)...)
	case "database":
		return filepath.Join(append([]string{layout.DatabaseDirForCreate(app)}, parts[1:]...)...)
	default:
		return app.BasePath(parts...)
	}
}

// ConsoleStubsDir locates framework console/stubs via go.mod replace or sibling trees.
func ConsoleStubsDir(app contracts.App) string {
	if app == nil {
		return ""
	}
	root := app.BasePath()
	var candidates []string
	if p := goModReplace(root, "github.com/zatrano/framework"); p != "" {
		candidates = append(candidates, filepath.Join(p, "console", "stubs"))
	}
	candidates = append(candidates,
		filepath.Join(filepath.Dir(root), "framework", "console", "stubs"),
		filepath.Join(filepath.Dir(root), "ZATRANO", "console", "stubs"),
		app.BasePath("console", "stubs"),
		app.BasePath("packages", "console", "stubs"),
		filepath.Join(filepath.Dir(root), "packages", "console", "stubs"),
	)
	for _, candidate := range candidates {
		if info, err := os.Stat(filepath.Join(candidate, "layouts", "auth.html")); err == nil && !info.IsDir() {
			return candidate
		}
		if info, err := os.Stat(filepath.Join(candidate, "layouts", "dashboard.html")); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func goModReplace(root, module string) string {
	f, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	defer f.Close()
	prefix := "replace " + module + " => "
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		p := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if i := strings.Index(p, "//"); i >= 0 {
			p = strings.TrimSpace(p[:i])
		}
		p = strings.Trim(p, `"`)
		if p == "" {
			return ""
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, filepath.FromSlash(p))
		}
		return filepath.FromSlash(p)
	}
	return ""
}

func modulePath(root string) (string, error) {
	f, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", sc.Err()
}

// ToSnake converts ExportedName to exported_name.
func ToSnake(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

// ToExported capitalizes the first letter.
func ToExported(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}
	runes := []rune(name)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] = runes[0] - 'a' + 'A'
	}
	return string(runes)
}
