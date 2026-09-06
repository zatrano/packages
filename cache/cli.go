package cache

import (
	"fmt"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/bootutil"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&CacheClearCommand{app: app},
	)
}

type CacheClearCommand struct {
	app contracts.App
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
