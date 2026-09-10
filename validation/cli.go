package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/bootutil"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&MakeRequestCommand{app: app},
		&MakeRuleCommand{app: app},
	)
}

type MakeRequestCommand struct {
	app contracts.App
}

func (c *MakeRequestCommand) Name() string        { return "make:request" }
func (c *MakeRequestCommand) Description() string {
	return "Create a form request (--store, --update, --index)"
}
func (c *MakeRequestCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("request name required")
	}
	intent := ""
	var nameArg string
	for _, arg := range args {
		switch arg {
		case "--store":
			if intent != "" {
				return fmt.Errorf("--store, --update, and --index are mutually exclusive")
			}
			intent = "store"
		case "--update":
			if intent != "" {
				return fmt.Errorf("--store, --update, and --index are mutually exclusive")
			}
			intent = "update"
		case "--index":
			if intent != "" {
				return fmt.Errorf("--store, --update, and --index are mutually exclusive")
			}
			intent = "index"
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag %s (want --store|--update|--index)", arg)
			}
			if nameArg == "" {
				nameArg = arg
			}
		}
	}
	if nameArg == "" {
		return fmt.Errorf("request name required")
	}
	name := requestTypeName(nameArg, intent)
	path := filepath.Join(c.app.BasePath("app", "http", "requests"), bootutil.ToSnake(name)+".go")
	content := requestStub(name, intent)
	if err := writeGenerated(path, content); err != nil {
		return err
	}
	fmt.Printf("Request created: %s\n", path)
	return nil
}

func requestTypeName(raw, intent string) string {
	name := bootutil.ToExported(raw)
	name = strings.TrimSuffix(name, "Request")
	switch intent {
	case "store":
		name = strings.TrimSuffix(name, "Store") + "Store"
	case "update":
		name = strings.TrimSuffix(name, "Update") + "Update"
	case "index":
		name = strings.TrimSuffix(name, "Index") + "Index"
	}
	return name + "Request"
}

func requestStub(name, intent string) string {
	rules := `		"name": "required|min:2",`
	if intent == "index" {
		rules = `		"q":    "nullable",
		"page": "nullable|integer",
		"sort": "nullable",`
	}
	return fmt.Sprintf(`package requests

import (
	. "github.com/zatrano/framework/v2/kernel/http"
	"github.com/zatrano/packages/validation"
)

type %s struct {
	validation.Base
}

func (r %s) Rules() map[string]string {
	return map[string]string{
%s
	}
}

func (r %s) Messages() map[string]string {
	return map[string]string{}
}

func (r %s) Authorize(req *Request) bool {
	return true
}
`, name, name, rules, name, name)
}

type MakeRuleCommand struct {
	app contracts.App
}

func (c *MakeRuleCommand) Name() string        { return "make:rule" }
func (c *MakeRuleCommand) Description() string { return "Create a custom validation rule scaffold" }
func (c *MakeRuleCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("rule name required")
	}
	name := args[0]
	ruleName := toRuleName(name)
	structName := bootutil.ToExported(name)
	if !strings.HasSuffix(strings.ToLower(structName), "rule") {
		structName += "Rule"
	}
	path := filepath.Join(c.app.BasePath("app", "rules"), bootutil.ToSnake(structName)+".go")
	content := fmt.Sprintf(`package rules

import "github.com/zatrano/packages/validation"

// Register%s registers the "%s" validation rule.
func Register%s() {
	validation.Extend(%q, func(v *validation.Validator, field, value, param string) bool {
		// TODO: implement "%s" rule.
		_ = param
		return value != ""
	})
}
`, structName, ruleName, structName, ruleName, ruleName)
	if err := writeGenerated(path, content); err != nil {
		return err
	}
	fmt.Printf("Rule created: %s\n", path)
	fmt.Printf("Call rules.Register%s() during boot (e.g. AppServiceProvider).\n", structName)
	return nil
}

func writeGenerated(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func toRuleName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "-", "_")
	var b strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		if r == ' ' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	out := strings.ToLower(b.String())
	out = strings.TrimSuffix(out, "_rule")
	return out
}
