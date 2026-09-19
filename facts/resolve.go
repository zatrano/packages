package facts

// App is satisfied by *kernel.Application without importing the root module.
type App interface {
	Make(abstract string) (any, error)
}

// From resolves the Fact/Reaction bus. Nil when the facts package is not enabled.
func From(app App) *Bus {
	if app == nil {
		return nil
	}
	raw, err := app.Make("facts")
	if err != nil {
		return nil
	}
	v, _ := raw.(*Bus)
	return v
}
