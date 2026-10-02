package template

// App is satisfied by *core.Application without importing the root module.
type App interface {
	Make(abstract string) (any, error)
}

// From resolves the template Engine from the application container.
func From(app App) Engine {
	if app == nil {
		return nil
	}
	raw, err := app.Make("template")
	if err != nil {
		return nil
	}
	e, _ := raw.(Engine)
	return e
}
