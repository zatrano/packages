package queue

import (
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/env"
	"github.com/zatrano/packages/database"
	"github.com/zatrano/packages/redisx"
)

func boot(app contracts.App) error {
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
	mgr := NewManager(env.Get("QUEUE_CONNECTION", "sync"), queues)
	app.Container().Instance("queue", mgr)
	return nil
}
