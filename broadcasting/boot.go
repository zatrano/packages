package broadcasting

import (
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/env"
	"github.com/zatrano/framework/v2/kernel/http"
)

func boot(app contracts.App) error {
	fileBroadcast, err := NewFileBroadcaster(app.BasePath("storage", "logs", "broadcast.jsonl"))
	if err != nil {
		return err
	}
	mgr := NewManager(env.Get("BROADCAST_CONNECTION", "log"), map[string]Broadcaster{
		"log":  NewLogBroadcaster(app.Logger()),
		"file": fileBroadcast,
		"null": NullBroadcaster{},
	})
	mgr.Channel("public", func(req *http.Request, channel string) bool {
		return true
	})
	mgr.Channel("private.*", func(req *http.Request, channel string) bool {
		raw, err := app.Make("auth")
		if err != nil || raw == nil {
			return false
		}
		type checker interface{ Check(req *http.Request) bool }
		c, ok := raw.(checker)
		return ok && c.Check(req)
	})
	app.Container().Instance("broadcast", mgr)
	return nil
}
