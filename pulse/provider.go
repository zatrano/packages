package pulse

import (
	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/inspector"
	"github.com/zatrano/packages/search"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "pulse",
		Key:         "pulse",
		Description: "Metrics pulse dashboard",
		Factory:     func() contracts.Provider { return &ServiceProvider{} },
	})
}

// ServiceProvider boots the pulse addon.
type ServiceProvider struct{}

func (p *ServiceProvider) Register(app contracts.App) error {
	if app.Metrics() == nil {
		return nil
	}
	dash := New(app.Metrics()).WithExtra(func() map[string]any {
		extra := map[string]any{}
		if insp := inspector.From(app); insp != nil {
			extra["inspector_entries"] = insp.Count()
		}
		if s := search.From(app); s != nil {
			extra["search_docs"] = s.Count()
		}
		return extra
	})
	app.Container().Instance("pulse", dash)
	return nil
}

func (p *ServiceProvider) Boot(app contracts.App) error { return nil }
