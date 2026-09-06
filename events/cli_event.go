package events

import (
	"fmt"
	"github.com/zatrano/framework/contracts"
	"os"
	"path/filepath"

	"github.com/zatrano/packages/bootutil"
)

type MakeEventCommand struct {
	app contracts.App
}

func (c *MakeEventCommand) Name() string        { return "make:event" }
func (c *MakeEventCommand) Description() string { return "Create a new event class" }
func (c *MakeEventCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("event name required")
	}
	name := args[0]
	dir := c.app.BasePath("app", "events")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(name)+".go")
	content := fmt.Sprintf(`package events

const %sName = "%s"

type %s struct {
	Payload map[string]any
}
`, name, bootutil.ToSnake(name), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Event created: %s\n", path)
	return nil
}

type MakeListenerCommand struct {
	app contracts.App
}

func (c *MakeListenerCommand) Name() string        { return "make:listener" }
func (c *MakeListenerCommand) Description() string { return "Create a new event listener" }
func (c *MakeListenerCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("listener name required")
	}
	name := args[0]
	dir := c.app.BasePath("app", "listeners")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(name)+".go")
	content := fmt.Sprintf(`package listeners

import "fmt"

func %s(event any) error {
	fmt.Printf("%s received: %%v\n", event)
	return nil
}
`, name, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Listener created: %s\n", path)
	return nil
}
