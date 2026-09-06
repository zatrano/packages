package events

import (
	"github.com/zatrano/framework/v2/contracts"
)

func boot(app contracts.App) error {
	dispatcher := New()
	app.Container().Instance("events", dispatcher)
	return nil
}
