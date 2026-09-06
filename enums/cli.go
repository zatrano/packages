package enums

import (
	"fmt"
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/bootutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zatrano/framework/v2/bootstrap"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&MakeEnumCommand{app: app},
	)
}

type MakeEnumCommand struct {
	app contracts.App
}

func (c *MakeEnumCommand) Name() string        { return "make:enum" }
func (c *MakeEnumCommand) Description() string { return "Create a string enum scaffold" }
func (c *MakeEnumCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("enum name required")
	}
	enabled := false
	for _, name := range bootstrap.EnabledAddons {
		if strings.EqualFold(strings.TrimSpace(name), "enums") {
			enabled = true
			break
		}
	}
	if !enabled {
		fmt.Println("Warning: enums addon is not in EnabledAddons — run `package:enable enums` to register the registry at boot.")
	}
	name := args[0]
	structName := bootutil.ToExported(name)
	enumKey := bootutil.ToSnake(structName)
	dir := c.app.BasePath("app", "enums")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(structName)+".go")
	cases := []string{`"draft:Draft"`, `"published:Published"`}
	if len(args) > 1 {
		cases = cases[:0]
		for _, raw := range args[1:] {
			cases = append(cases, fmt.Sprintf("%q", raw))
		}
	}
	content := fmt.Sprintf(`package enums

import "github.com/zatrano/packages/enums"

// %s is a backed string enumeration.
var %s = enums.NewString(%q, %s)

// Register%s registers the enum on a registry.
func Register%s(reg *enums.Registry) {
	if reg == nil {
		return
	}
	reg.Register(%s)
}
`, structName, structName, enumKey, strings.Join(cases, ", "), structName, structName, structName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Enum created: %s\n", path)
	fmt.Printf("Call enums.Register%s during boot (resolve enums from the container).\n", structName)
	return nil
}
