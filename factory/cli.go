package factory

import (
	"fmt"
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/bootutil"
	"os"
	"path/filepath"

	"github.com/zatrano/framework/kernel/dirs"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&MakeFactoryCommand{app: app},
	)
}

type MakeResourceCommand struct {
	app contracts.App
}

func (c *MakeResourceCommand) Name() string        { return "make:resource" }
func (c *MakeResourceCommand) Description() string { return "Create a new API resource" }
func (c *MakeResourceCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("resource name required")
	}
	name := args[0]
	mod := bootutil.ConsumerModule(c.app)
	dir := c.app.BasePath("app", "http", "resources")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(name)+"_resource.go")
	content := fmt.Sprintf(`package resources

import "%s/app/models"

// %s transforms a models.%s into an API resource array.
// Use with github.com/zatrano/packages/resources:
//
//	resources.JSON(model, %s)
func %s(model models.%s) map[string]any {
	return map[string]any{
		"id": model.ID,
	}
}
`, mod, name, name, name, name, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Resource created: %s\n", path)
	return nil
}

type MakeFactoryCommand struct {
	app contracts.App
}

func (c *MakeFactoryCommand) Name() string        { return "make:factory" }
func (c *MakeFactoryCommand) Description() string { return "Create a new model factory" }
func (c *MakeFactoryCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("factory model name required")
	}
	name := args[0]
	mod := bootutil.ConsumerModule(c.app)
	dir := filepath.Join(dirs.DatabaseDirForCreate(c.app), "factories")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(name)+"_factory.go")
	content := fmt.Sprintf(`package factories

import (
	"%s/app/models"
	"github.com/zatrano/packages/factory"
)

func init() {
	factory.For[models.%s](func() map[string]any {
		return map[string]any{
			"name": factory.FakeName(),
		}
	})
}
`, mod, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Factory created: %s\n", path)
	fmt.Printf("Import _ %q from your seeder/tests to register it.\n", mod+"/app/database/factories")
	return nil
}
