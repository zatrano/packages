package httpclient

import (
	"github.com/zatrano/framework/contracts"
)

func boot(app contracts.App) error {
	app.Container().Instance("http", New())
	return nil
}
