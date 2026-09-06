package schedule

import (
	"github.com/zatrano/framework/v2/contracts"
)

func boot(app contracts.App) error {
	s := New()
	s.SetMutexPath(app.BasePath("storage", "framework", "schedule"))
	app.Container().Instance("scheduler", s)
	return nil
}
