package bootutil

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/dirs"
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
	if err != nil || strings.TrimSpace(mod) == "" || mod == "github.com/zatrano/framework/v2" {
		return "your/module"
	}
	return mod
}

// ApplyConsumerPlaceholders rewrites generated stub module paths.
func ApplyConsumerPlaceholders(app contracts.App, body string) string {
	mod := ConsumerModule(app)
	body = strings.ReplaceAll(body, "__MODULE__", mod)
	body = strings.ReplaceAll(body, "github.com/zatrano/framework/v2/app/", mod+"/app/")
	return body
}

// ScaffoldDest maps starter path prefixes onto app/views, app/localization, app/database.
func ScaffoldDest(app contracts.App, parts []string) string {
	if app == nil || len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "views":
		return filepath.Join(append([]string{dirs.ViewsDirForCreate(app)}, parts[1:]...)...)
	case "lang":
		return filepath.Join(append([]string{dirs.LocalizationDirForCreate(app)}, parts[1:]...)...)
	case "database":
		return filepath.Join(append([]string{dirs.DatabaseDirForCreate(app)}, parts[1:]...)...)
	default:
		return app.BasePath(parts...)
	}
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

// EnsureBlankImport inserts `_ "importPath"` into the first import block of a Go file.
func EnsureBlankImport(filePath, importPath string) error {
	importPath = strings.TrimSpace(importPath)
	if filePath == "" || importPath == "" {
		return nil
	}
	body, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	text := string(body)
	quoted := `"` + importPath + `"`
	if strings.Contains(text, quoted) {
		return nil
	}
	line := "\t_ " + quoted + "\n"
	const marker = "import ("
	idx := strings.Index(text, marker)
	if idx < 0 {
		return nil
	}
	insert := idx + len(marker)
	if insert < len(text) && text[insert] == '\r' {
		insert++
	}
	if insert < len(text) && text[insert] == '\n' {
		insert++
	}
	out := text[:insert] + line + text[insert:]
	return os.WriteFile(filePath, []byte(out), 0o644)
}

// EnsureEnabledAddon inserts name into bootstrap/enabled.go EnabledAddons when missing.
func EnsureEnabledAddon(app contracts.App, name string) error {
	if app == nil {
		return nil
	}
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return nil
	}
	path := app.BasePath("bootstrap", "enabled.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	text := string(body)
	quoted := `"` + name + `"`
	if strings.Contains(text, quoted) {
		return nil
	}
	const marker = "var EnabledAddons = []string{"
	idx := strings.Index(text, marker)
	if idx < 0 {
		return nil
	}
	insert := idx + len(marker)
	if insert < len(text) && text[insert] == '\r' {
		insert++
	}
	if insert < len(text) && text[insert] == '\n' {
		insert++
	}
	line := "\t" + quoted + ",\n"
	out := text[:insert] + line + text[insert:]
	return os.WriteFile(path, []byte(out), 0o644)
}
