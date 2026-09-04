package authorization

import (
	"fmt"
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/bootutil"
	"os"
	"path/filepath"

	"github.com/zatrano/framework/kernel"
)

func Commands(app contracts.App) []addons.CLICommand {
	k := bootutil.KernelApp(app)
	if k == nil {
		return nil
	}
	return bootutil.CLI(
		&MakePolicyCommand{app: k},
	)
}

type MakePolicyCommand struct {
	app *kernel.Application
}

func (c *MakePolicyCommand) Name() string        { return "make:policy" }
func (c *MakePolicyCommand) Description() string { return "Create a new policy class" }
func (c *MakePolicyCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("policy name required")
	}
	name := args[0]
	if len(name) < 6 || name[len(name)-6:] != "Policy" {
		name += "Policy"
	}
	dir := c.app.BasePath("app", "policies")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(name)+".go")
	content := fmt.Sprintf(`package policies

import "github.com/zatrano/packages/authorization"

func New%s() *authorization.Policy {
	return authorization.NewPolicy().
		Define("view", func(user authorization.Authenticatable, arguments ...any) bool {
			return user != nil
		}).
		Define("update", func(user authorization.Authenticatable, arguments ...any) bool {
			return user != nil
		}).
		Define("delete", func(user authorization.Authenticatable, arguments ...any) bool {
			return user != nil
		})
}
`, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Policy created: %s\n", path)
	fmt.Println("Register it with authorization.From(app).Policy(\"name\", policies.New" + name + "())")
	return nil
}
