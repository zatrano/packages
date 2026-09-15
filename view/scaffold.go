package view

import (
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/view/starter"
)

// WriteStarterViews writes layout/app.html and web/welcome.html when missing.
func WriteStarterViews(app contracts.App) error {
	return starter.Write(app)
}
