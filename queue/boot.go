package queue

import (
	"fmt"

	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/env"
	"github.com/zatrano/packages/database"
	"github.com/zatrano/packages/redisx"
)

func boot(app contracts.App) error {
	want := env.Get("QUEUE_CONNECTION", "sync")
	queues := map[string]Queue{"sync": NewSyncQueue()}
	if dbMgr := database.From(app); dbMgr != nil {
		if db, err := dbMgr.DB(); err == nil {
			driver, _ := dbMgr.DriverName()
			dbQueue := NewDatabaseQueue(db, "jobs", driver)
			_ = dbQueue.EnsureTable()
			queues["database"] = dbQueue
		}
	}
	if raw, err := app.Make("redis"); err == nil {
		if client := redisx.ClientFrom(raw); client != nil {
			queues["redis"] = NewRedisQueue(client, "zatrano:queues:default")
		}
	}
	mgr := NewManager(want, queues)
	if mgr.Queue() == nil {
		return fmt.Errorf("queue: connection %q is not available", want)
	}
	app.Container().Instance("queue", mgr)
	return nil
}
