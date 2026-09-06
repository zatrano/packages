package maintenance

import (
	"fmt"
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/packages/bootutil"
	"strconv"
	"strings"

	"github.com/zatrano/framework/v2/contracts"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&DownCommand{app: app},
		&UpCommand{app: app},
	)
}

type DownCommand struct {
	app contracts.App
}

func (c *DownCommand) Name() string        { return "down" }
func (c *DownCommand) Description() string { return "Put the application into maintenance mode" }
func (c *DownCommand) Handle(args []string) error {
	if err := c.app.Bootstrap(); err != nil {
		return err
	}
	payload := contracts.MaintenancePayload{
		Message:    "Application is in maintenance mode.",
		RetryAfter: 60,
	}
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--message" && i+1 < len(args):
			payload.Message = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--message="):
			payload.Message = strings.TrimPrefix(args[i], "--message=")
		case args[i] == "--retry" && i+1 < len(args):
			if n, err := strconv.Atoi(args[i+1]); err == nil {
				payload.RetryAfter = n
			}
			i++
		case strings.HasPrefix(args[i], "--retry="):
			if n, err := strconv.Atoi(strings.TrimPrefix(args[i], "--retry=")); err == nil {
				payload.RetryAfter = n
			}
		case args[i] == "--secret" && i+1 < len(args):
			payload.Secret = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--secret="):
			payload.Secret = strings.TrimPrefix(args[i], "--secret=")
		case args[i] == "--allow" && i+1 < len(args):
			payload.AllowedIPs = append(payload.AllowedIPs, strings.Split(args[i+1], ",")...)
			i++
		}
	}
	m := From(c.app)
	if m == nil {
		return fmt.Errorf("maintenance unavailable")
	}
	if err := m.Enable(payload); err != nil {
		return err
	}
	fmt.Println("Application is now in maintenance mode.")
	return nil
}

type UpCommand struct {
	app contracts.App
}

func (c *UpCommand) Name() string        { return "up" }
func (c *UpCommand) Description() string { return "Bring the application out of maintenance mode" }
func (c *UpCommand) Handle(args []string) error {
	if err := c.app.Bootstrap(); err != nil {
		return err
	}
	m := From(c.app)
	if m == nil {
		return fmt.Errorf("maintenance unavailable")
	}
	if err := m.Disable(); err != nil {
		return err
	}
	fmt.Println("Application is now live.")
	return nil
}
