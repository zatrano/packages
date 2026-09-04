package view

import (
	"fmt"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/bootutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zatrano/framework/kernel/layout"
)

type MakeComponentCommand struct {
	app contracts.App
}

func (c *MakeComponentCommand) Name() string        { return "make:component" }
func (c *MakeComponentCommand) Description() string { return "Create a view component scaffold" }
func (c *MakeComponentCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("component name required")
	}
	name := args[0]
	slug := strings.TrimSuffix(bootutil.ToSnake(bootutil.ToExported(name)), "_component")
	slug = strings.ReplaceAll(slug, "_", "-")
	if slug == "" {
		slug = strings.ToLower(name)
	}
	dir := filepath.Join(layout.ViewsDirForCreate(c.app), "components")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, slug+".html")
	content := fmt.Sprintf(`<div class="component-%s">
  {{ index . "slot" }}
</div>
`, slug)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Component created: %s\n", path)
	fmt.Printf("Render with: view.From(app).Component(%q, map[string]any{\"slot\": \"...\"})\n", slug)
	return nil
}
