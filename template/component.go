package template

import "strings"

// Component renders a Canvas template under components/{name}.
func (e *Engine) Component(name string, data map[string]any) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "components.")
	name = strings.TrimPrefix(name, "components/")
	return e.Render("components."+name, data)
}
