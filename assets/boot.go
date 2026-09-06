package assets

import (
	"strings"

	"github.com/zatrano/framework/contracts"
)

func boot(app contracts.App) error {
	publicURL := strings.TrimRight(app.Config().GetString("app.url", "http://localhost:8080"), "/")
	mgr := LoadDefault(app.BasePath(), publicURL)
	app.Container().Instance("assets", mgr)
	return nil
}
