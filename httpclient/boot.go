package httpclient

import (
	"github.com/zatrano/framework/v3/core/contracts"
)

func boot(app contracts.App) error {
	app.Container().Instance("http", New())
	return nil
}
