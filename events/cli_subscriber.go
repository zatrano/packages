package events

import (
	"fmt"
	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/bootutil"
	"os"
	"path/filepath"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&MakeSubscriberCommand{app: app},
		&MakeEventCommand{app: app},
		&MakeListenerCommand{app: app},
	)
}

type MakeSubscriberCommand struct {
	app contracts.App
}

func (c *MakeSubscriberCommand) Name() string        { return "make:subscriber" }
func (c *MakeSubscriberCommand) Description() string { return "Create an event subscriber scaffold" }
func (c *MakeSubscriberCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("subscriber name required")
	}
	name := bootutil.ToExported(args[0])
	dir := c.app.BasePath("app", "subscribers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(name)+".go")
	content := fmt.Sprintf(`package subscribers

import "github.com/zatrano/packages/events"

// %s registers related event listeners.
type %s struct{}

// Subscribe wires listeners for %s.
func (s *%s) Subscribe(d *events.Dispatcher) {
	d.Listen("example.event", s.HandleExample)
}

// HandleExample handles the example.event event.
func (s *%s) HandleExample(event any) error {
	// TODO: handle event
	_ = event
	return nil
}
`, name, name, name, name, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Subscriber created: %s\n", path)
	fmt.Printf("Register with events.Dispatcher.Register(&subscribers.%s{})\n", name)
	return nil
}
