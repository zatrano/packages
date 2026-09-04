package cache

import (
	"fmt"

	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/kernel"
	"github.com/zatrano/packages/bootutil"
)

func Commands(app contracts.App) []addons.CLICommand {
	k := bootutil.KernelApp(app)
	if k == nil {
		return nil
	}
	return bootutil.CLI(
		&CacheClearCommand{app: k},
	)
}

type CacheClearCommand struct {
	app *kernel.Application
}

func (c *CacheClearCommand) Name() string        { return "cache:clear" }
func (c *CacheClearCommand) Description() string { return "Flush the application cache" }
func (c *CacheClearCommand) Handle(args []string) error {
	if err := c.app.Bootstrap(); err != nil {
		return err
	}
	mgr := From(c.app)
	if mgr == nil {
		return fmt.Errorf("cache is not bound; import github.com/zatrano/packages/cache")
	}
	if err := mgr.Flush(); err != nil {
		return err
	}
	fmt.Println("Application cache cleared successfully.")
	return nil
}
