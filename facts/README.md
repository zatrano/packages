# facts

Typed application reactions. A **Fact** is something that already happened. A **Reaction** is what should happen because of it. **Sync** and **Async** are execution policies, not properties of the Fact.

There is no Event, Listener, Subscriber, string name, or global bus.

```go
bus := facts.From(app)

facts.On[UserRegistered](bus, facts.Sync(SendWelcomeMail{Mail: mail}))
facts.On[UserRegistered](bus, facts.Async(IndexUser{Search: search}))

err := bus.Publish(ctx, UserRegistered{UserID: user.ID, Email: user.Email})
```

`On[T]` is a package function because Go methods cannot be generic. Pass the bus from `facts.From(app)`. Do not use a process-wide singleton.

## Semantics

- Publish waits for Sync Reactions and for Async **enqueue**, not for worker completion.
- Independent Reactions all run. Sync errors are aggregated with `errors.Join`.
- Async enqueue failure is returned from Publish. Worker failures are not.
- Async is an in-process job queue plus workers (`LifecycleProvider.Start` / `Stop`). Not `go React(...)`. `Stop` then `Start` reopens the in-process queue (kernel Start retry).
- Swap the transport with `WithQueue` (Redis, database, Outbox relay) without changing `On` / `Publish` / `Reaction[T]`. Redis/Kafka are not required.
- At-least-once execution may occur. Reactions should be idempotent where practical.
- Do not depend on Reaction order. Do not publish inside an uncommitted transaction.
- ORM model hooks (`BeforeCreate` / `AfterCreate`, …) are not Facts. Jobs are not Facts.

Async workers start when the application `Start`s. `package:enable facts` is Enabled ∩ Imported.
