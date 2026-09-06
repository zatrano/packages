package orm

import (
	"fmt"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/bootutil"
	"os"
	"path/filepath"
	"strings"
)

type MakeCastCommand struct {
	app contracts.App
}

func (c *MakeCastCommand) Name() string        { return "make:cast" }
func (c *MakeCastCommand) Description() string { return "Create a custom ORM cast scaffold" }
func (c *MakeCastCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("cast name required")
	}
	name := args[0]
	structName := bootutil.ToExported(name)
	castKey := strings.ToLower(bootutil.ToSnake(structName))
	castKey = strings.TrimSuffix(castKey, "_cast")
	dir := c.app.BasePath("app", "casts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(structName)+".go")
	content := fmt.Sprintf(`package casts

import (
	"fmt"

	"github.com/zatrano/packages/orm"
)

// Register%s registers the "%s" attribute cast.
func Register%s() {
	orm.RegisterCast(%q, Cast%sIn, Cast%sOut)
}

// Cast%sIn converts a stored value into the model attribute.
func Cast%sIn(value any) (any, error) {
	// TODO: implement incoming cast for "%s".
	return fmt.Sprint(value), nil
}

// Cast%sOut converts a model attribute into a storable value.
func Cast%sOut(value any) any {
	// TODO: implement outgoing cast for "%s".
	return fmt.Sprint(value)
}
`, structName, castKey, structName, castKey, structName, structName, structName, structName, castKey, structName, structName, castKey)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Cast created: %s\n", path)
	fmt.Printf("Call casts.Register%s() during boot (e.g. AppServiceProvider).\n", structName)
	return nil
}
