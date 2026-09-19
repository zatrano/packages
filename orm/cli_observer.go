package orm

import (
	"fmt"
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/bootutil"
	"os"
	"path/filepath"
	"strings"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&MakeObserverCommand{app: app},
		&MakeRepositoryCommand{app: app},
		&MakeScopeCommand{app: app},
		&MakeCastCommand{app: app},
	)
}

type MakeObserverCommand struct {
	app contracts.App
}

func (c *MakeObserverCommand) Name() string { return "make:observer" }
func (c *MakeObserverCommand) Description() string {
	return "Create a model persistence observer scaffold"
}
func (c *MakeObserverCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("observer name required")
	}
	name := args[0]
	structName := bootutil.ToExported(name)
	if !strings.HasSuffix(strings.ToLower(structName), "observer") {
		structName += "Observer"
	}
	subject := strings.TrimSuffix(strings.ToLower(bootutil.ToSnake(structName)), "_observer")
	dir := c.app.BasePath("app", "observers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(structName)+".go")
	content := fmt.Sprintf(`package observers

import (
	"fmt"

	"github.com/zatrano/packages/orm"
)

// %s observes "%s.*" persistence lifecycle hooks. These are not application Facts.
type %s struct{}

var _ orm.ModelObserver = (*%s)(nil)

func (o *%s) Created(model any) error {
	fmt.Printf("%s created: %%v\n", model)
	return nil
}

func (o *%s) Updated(model any) error {
	fmt.Printf("%s updated: %%v\n", model)
	return nil
}

func (o *%s) Deleted(model any) error {
	fmt.Printf("%s deleted: %%v\n", model)
	return nil
}

// Register%s attaches the observer to ORM persistence hooks.
func Register%s() {
	orm.ObserveModel(%q, &%s{})
}
`, structName, subject, structName, structName, structName, subject, structName, subject, structName, subject, structName, structName, subject, structName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Observer created: %s\n", path)
	fmt.Printf("Call observers.Register%s() during boot.\n", structName)
	return nil
}
