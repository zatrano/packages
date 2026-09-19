package facts_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zatrano/packages/facts"
)

type UserRegistered struct {
	UserID uint64
	Email  string
}

type OrderCreated struct {
	OrderID uint64
}

type record struct {
	mu    sync.Mutex
	users []UserRegistered
	n     int
}

func (r *record) add(u UserRegistered) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users = append(r.users, u)
	r.n++
}

func (r *record) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.n
}

type captureUser struct{ rec *record }

func (c captureUser) React(ctx context.Context, fact UserRegistered) error {
	c.rec.add(fact)
	return nil
}

type captureOrder struct{ n *atomic.Int32 }

func (c captureOrder) React(ctx context.Context, fact OrderCreated) error {
	c.n.Add(1)
	return nil
}

func TestTypedRouting(t *testing.T) {
	bus := facts.New()
	users := &record{}
	var orders atomic.Int32
	facts.On[UserRegistered](bus, facts.Sync(captureUser{rec: users}))
	facts.On[OrderCreated](bus, facts.Sync(captureOrder{n: &orders}))

	if err := bus.Publish(context.Background(), UserRegistered{UserID: 1, Email: "ada@zatrano.test"}); err != nil {
		t.Fatal(err)
	}
	if users.count() != 1 || orders.Load() != 0 {
		t.Fatalf("users=%d orders=%d", users.count(), orders.Load())
	}
	if err := bus.Publish(context.Background(), OrderCreated{OrderID: 9}); err != nil {
		t.Fatal(err)
	}
	if users.count() != 1 || orders.Load() != 1 {
		t.Fatalf("users=%d orders=%d", users.count(), orders.Load())
	}
}

func TestMultipleReactions(t *testing.T) {
	bus := facts.New()
	var a, b atomic.Int32
	facts.On[UserRegistered](bus,
		facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
			a.Add(1)
			return nil
		})),
		facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
			b.Add(1)
			return nil
		})),
	)
	if err := bus.Publish(context.Background(), UserRegistered{UserID: 2}); err != nil {
		t.Fatal(err)
	}
	if a.Load() != 1 || b.Load() != 1 {
		t.Fatalf("a=%d b=%d", a.Load(), b.Load())
	}
}

func TestSyncWaitsForCompletion(t *testing.T) {
	bus := facts.New()
	started := make(chan struct{})
	facts.On[UserRegistered](bus, facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		close(started)
		time.Sleep(40 * time.Millisecond)
		return nil
	})))
	begin := time.Now()
	if err := bus.Publish(context.Background(), UserRegistered{}); err != nil {
		t.Fatal(err)
	}
	if time.Since(begin) < 40*time.Millisecond {
		t.Fatal("Publish returned before Sync Reaction finished")
	}
	select {
	case <-started:
	default:
		t.Fatal("sync reaction did not run")
	}
}

func TestSyncErrorIsolationAndJoin(t *testing.T) {
	bus := facts.New()
	var ran atomic.Int32
	errA := errors.New("a")
	errC := errors.New("c")
	facts.On[UserRegistered](bus,
		facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
			ran.Add(1)
			return errA
		})),
		facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
			ran.Add(1)
			return nil
		})),
		facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
			ran.Add(1)
			return errC
		})),
	)
	err := bus.Publish(context.Background(), UserRegistered{})
	if ran.Load() != 3 {
		t.Fatalf("ran=%d", ran.Load())
	}
	if err == nil || !errors.Is(err, errA) || !errors.Is(err, errC) {
		t.Fatalf("aggregate=%v", err)
	}
}

func TestDuplicatePublishRunsTwice(t *testing.T) {
	bus := facts.New()
	var n atomic.Int32
	facts.On[UserRegistered](bus, facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		n.Add(1)
		return nil
	})))
	_ = bus.Publish(context.Background(), UserRegistered{})
	_ = bus.Publish(context.Background(), UserRegistered{})
	if n.Load() != 2 {
		t.Fatalf("n=%d", n.Load())
	}
}

func TestNestedPublish(t *testing.T) {
	bus := facts.New()
	var nested atomic.Int32
	facts.On[UserRegistered](bus, facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		return bus.Publish(ctx, OrderCreated{OrderID: fact.UserID})
	})))
	facts.On[OrderCreated](bus, facts.Sync(facts.ReactionFunc[OrderCreated](func(ctx context.Context, fact OrderCreated) error {
		nested.Add(1)
		return nil
	})))
	if err := bus.Publish(context.Background(), UserRegistered{UserID: 7}); err != nil {
		t.Fatal(err)
	}
	if nested.Load() != 1 {
		t.Fatalf("nested=%d", nested.Load())
	}
}

func TestContextReachesReaction(t *testing.T) {
	bus := facts.New()
	type ctxKey struct{}
	var got any
	facts.On[UserRegistered](bus, facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		got = ctx.Value(ctxKey{})
		return nil
	})))
	ctx := context.WithValue(context.Background(), ctxKey{}, "trace")
	if err := bus.Publish(ctx, UserRegistered{}); err != nil {
		t.Fatal(err)
	}
	if got != "trace" {
		t.Fatalf("got=%v", got)
	}
}

func TestNoImplicitDiscovery(t *testing.T) {
	bus := facts.New()
	if err := bus.Publish(context.Background(), UserRegistered{UserID: 1}); err != nil {
		t.Fatal(err)
	}
}

type failQueue struct{ err error }

func (q failQueue) Enqueue(ctx context.Context, job *facts.Job) error { return q.err }
func (q failQueue) Dequeue(ctx context.Context) (*facts.Job, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}
func (q failQueue) Close() error { return nil }

func TestAsyncSubmissionFailure(t *testing.T) {
	boom := errors.New("queue down")
	bus := facts.New(facts.WithQueue(failQueue{err: boom}))
	facts.On[UserRegistered](bus, facts.Async(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		return nil
	})))
	err := bus.Publish(context.Background(), UserRegistered{})
	if !errors.Is(err, boom) {
		t.Fatalf("err=%v", err)
	}
}
