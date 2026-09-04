package events

import (
	"github.com/zatrano/framework/contracts"
)

func boot(app contracts.App) error {
	dispatcher := New()
	app.Container().Instance("events", dispatcher)
	return nil
}
