package octane

import (
	"fmt"
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/bootutil"
	"runtime"
	"strconv"
	"strings"

	"github.com/zatrano/framework/kernel"
)

func Commands(app contracts.App) []addons.CLICommand {
	k := bootutil.KernelApp(app)
	if k == nil {
		return nil
	}
	return bootutil.CLI(
		&OctaneStartCommand{app: k},
	)
}

type OctaneStartCommand struct {
	app *kernel.Application
}

func (c *OctaneStartCommand) Name() string { return "octane:start" }
func (c *OctaneStartCommand) Description() string {
	return "Start HTTP serving with Octane metrics and optional GOMAXPROCS worker hint"
}
func (c *OctaneStartCommand) Handle(args []string) error {
	addr := ":8080"
	workers := 0
	for i := 0; i < len(args); i++ {
		if (args[i] == "--port" || args[i] == "-p") && i+1 < len(args) {
			addr = ":" + args[i+1]
			i++
		}
		if strings.HasPrefix(args[i], "--host=") {
			addr = strings.TrimPrefix(args[i], "--host=")
		}
		if (args[i] == "--workers" || args[i] == "-w") && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil {
				workers = n
			}
			i++
		}
	}
	if err := c.app.Bootstrap(); err != nil {
		return err
	}
	if workers > 0 {
		runtime.GOMAXPROCS(workers)
	}
	fmt.Printf("Octane workers hint=%d gomaxprocs=%d\n", workers, runtime.GOMAXPROCS(0))
	return c.app.Run(addr)
}
