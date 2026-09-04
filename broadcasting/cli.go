package broadcasting

import (
	"fmt"
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/bootutil"
	"os"
	"path/filepath"
	"strings"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&MakeChannelCommand{app: app},
	)
}

type MakeChannelCommand struct {
	app contracts.App
}

func (c *MakeChannelCommand) Name() string { return "make:channel" }
func (c *MakeChannelCommand) Description() string {
	return "Create a broadcasting channel authorizer scaffold"
}
func (c *MakeChannelCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("channel name required")
	}
	name := args[0]
	structName := bootutil.ToExported(name)
	if !strings.HasSuffix(strings.ToLower(structName), "channel") {
		structName += "Channel"
	}
	pattern := strings.TrimSuffix(strings.ToLower(bootutil.ToSnake(structName)), "_channel")
	if !strings.Contains(pattern, ".") {
		pattern = pattern + ".*"
	}
	dir := c.app.BasePath("app", "broadcasting")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(structName)+".go")
	content := fmt.Sprintf(`package broadcasting

import (
	"github.com/zatrano/packages/auth"
	corebroadcast "github.com/zatrano/packages/broadcasting"
	. "github.com/zatrano/framework/http"
)

// %s authorizes "%s" channels.
type %s struct{}

// Authorize decides whether the request may subscribe.
func (c *%s) Authorize(req *Request, channel string) bool {
	_ = channel
	mgr, _ := req.Get("auth").(*auth.Manager)
	if mgr == nil {
		return false
	}
	return mgr.Check(req)
}

// Register%s registers the channel with the broadcaster.
func Register%s(registry *corebroadcast.Manager) {
	registry.Channel(%q, (&%s{}).Authorize)
}
`, structName, pattern, structName, structName, structName, structName, pattern, structName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Channel created: %s\n", path)
	fmt.Printf("Call broadcasting.Register%s(broadcasting.From(app)) during boot.\n", structName)
	return nil
}
