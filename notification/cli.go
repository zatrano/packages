package notification

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/bootutil"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&MakeNotificationCommand{app: app},
	)
}

type MakeNotificationCommand struct {
	app contracts.App
}

func (c *MakeNotificationCommand) Name() string        { return "make:notification" }
func (c *MakeNotificationCommand) Description() string { return "Create a new notification class" }
func (c *MakeNotificationCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("notification name required")
	}
	name := args[0]
	dir := c.app.BasePath("app", "notifications")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(name)+".go")
	content := fmt.Sprintf(`package notifications

import (
	"github.com/zatrano/packages/notification"
)

type %s struct {
	notification.Base
	Message string
}

func (n *%s) Via() []string {
	return []string{"mail", "database"}
}

func (n *%s) ToMail(notifiable notification.Notifiable) *notification.MailMessage {
	return &notification.MailMessage{
		Subject: "%s",
		HTML:    "<p>" + n.Message + "</p>",
		Text:    n.Message,
	}
}

func (n *%s) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{
		"message": n.Message,
	}
}

func (n *%s) ToPush(notifiable notification.Notifiable) map[string]any {
	return map[string]any{
		"title":   "%s",
		"message": n.Message,
	}
}
`, name, name, name, name, name, name, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Notification created: %s\n", path)
	return nil
}
