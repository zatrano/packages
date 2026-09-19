package facts_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zatrano/packages/facts"
)

func TestAsyncDoesNotWaitForWorker(t *testing.T) {
	bus := facts.New(facts.WithWorkers(1), facts.WithQueueSize(8), facts.WithRetry(1, 0))
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = bus.Stop(context.Background()) }()

	started := make(chan struct{})
	done := make(chan struct{})
	facts.On[UserRegistered](bus, facts.Async(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		close(started)
		time.Sleep(80 * time.Millisecond)
		close(done)
		return nil
	})))
	begin := time.Now()
	if err := bus.Publish(context.Background(), UserRegistered{UserID: 1}); err != nil {
		t.Fatal(err)
	}
	if time.Since(begin) > 50*time.Millisecond {
		t.Fatal("Publish waited for Async Reaction")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("async reaction never started")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("async reaction never finished")
	}
}

func TestMixedSyncAsync(t *testing.T) {
	bus := facts.New(facts.WithWorkers(2), facts.WithQueueSize(8), facts.WithRetry(1, 0))
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = bus.Stop(context.Background()) }()

	var syncN, asyncN atomic.Int32
	asyncDone := make(chan struct{})
	facts.On[UserRegistered](bus,
		facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
			syncN.Add(1)
			return nil
		})),
		facts.Sync(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
			syncN.Add(1)
			return nil
		})),
		facts.Async(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
			asyncN.Add(1)
			return nil
		})),
		facts.Async(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
			if asyncN.Add(1) >= 2 {
				close(asyncDone)
			}
			return nil
		})),
	)
	if err := bus.Publish(context.Background(), UserRegistered{}); err != nil {
		t.Fatal(err)
	}
	if syncN.Load() != 2 {
		t.Fatalf("sync=%d", syncN.Load())
	}
	select {
	case <-asyncDone:
	case <-time.After(time.Second):
		t.Fatalf("async=%d", asyncN.Load())
	}
}

func TestAsyncExecutionFailureNotReturned(t *testing.T) {
	bus := facts.New(facts.WithWorkers(1), facts.WithQueueSize(4), facts.WithRetry(1, 0))
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = bus.Stop(context.Background()) }()

	failed := make(chan struct{})
	facts.On[UserRegistered](bus, facts.Async(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		close(failed)
		return errors.New("worker boom")
	})))
	if err := bus.Publish(context.Background(), UserRegistered{}); err != nil {
		t.Fatalf("Publish must not return worker errors: %v", err)
	}
	select {
	case <-failed:
	case <-time.After(time.Second):
		t.Fatal("async failure never ran")
	}
}

func TestRetryThenSuccess(t *testing.T) {
	bus := facts.New(facts.WithWorkers(1), facts.WithQueueSize(8), facts.WithRetry(3, time.Millisecond))
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = bus.Stop(context.Background()) }()

	var attempts atomic.Int32
	ok := make(chan struct{})
	facts.On[UserRegistered](bus, facts.Async(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		n := attempts.Add(1)
		if n < 3 {
			return errors.New("try again")
		}
		close(ok)
		return nil
	})))
	if err := bus.Publish(context.Background(), UserRegistered{}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ok:
	case <-time.After(2 * time.Second):
		t.Fatalf("attempts=%d", attempts.Load())
	}
	if attempts.Load() != 3 {
		t.Fatalf("attempts=%d", attempts.Load())
	}
}

func TestPermanentFailureNotRetried(t *testing.T) {
	bus := facts.New(facts.WithWorkers(1), facts.WithQueueSize(4), facts.WithRetry(5, time.Millisecond))
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = bus.Stop(context.Background()) }()

	var attempts atomic.Int32
	done := make(chan struct{})
	facts.On[UserRegistered](bus, facts.Async(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		attempts.Add(1)
		close(done)
		return facts.Permanent(errors.New("no"))
	})))
	if err := bus.Publish(context.Background(), UserRegistered{}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("did not run")
	}
	time.Sleep(30 * time.Millisecond)
	if attempts.Load() != 1 {
		t.Fatalf("attempts=%d", attempts.Load())
	}
}

func TestWorkerConcurrencyLimit(t *testing.T) {
	bus := facts.New(facts.WithWorkers(2), facts.WithQueueSize(16), facts.WithRetry(1, 0))
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = bus.Stop(context.Background()) }()

	var current, max atomic.Int32
	var done atomic.Int32
	finished := make(chan struct{})
	facts.On[UserRegistered](bus, facts.Async(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		n := current.Add(1)
		for {
			m := max.Load()
			if n <= m || max.CompareAndSwap(m, n) {
				break
			}
		}
		time.Sleep(40 * time.Millisecond)
		current.Add(-1)
		if done.Add(1) == 6 {
			close(finished)
		}
		return nil
	})))
	for i := 0; i < 6; i++ {
		if err := bus.Publish(context.Background(), UserRegistered{UserID: uint64(i)}); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case <-finished:
	case <-time.After(3 * time.Second):
		t.Fatal("jobs did not finish")
	}
	if max.Load() > 2 {
		t.Fatalf("max concurrent=%d", max.Load())
	}
}

func TestShutdownStopsWorkers(t *testing.T) {
	bus := facts.New(facts.WithWorkers(1), facts.WithQueueSize(4), facts.WithRetry(1, 0), facts.WithStopTimeout(time.Second))
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}
	var n atomic.Int32
	facts.On[UserRegistered](bus, facts.Async(facts.ReactionFunc[UserRegistered](func(ctx context.Context, fact UserRegistered) error {
		n.Add(1)
		return nil
	})))
	if err := bus.Publish(context.Background(), UserRegistered{}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 50 && n.Load() != 1; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if n.Load() != 1 {
		t.Fatalf("n=%d", n.Load())
	}
	if err := bus.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(context.Background(), UserRegistered{}); err == nil {
		t.Fatal("expected enqueue failure after stop")
	}
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = bus.Stop(context.Background()) }()
	if err := bus.Publish(context.Background(), UserRegistered{}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 50 && n.Load() != 2; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if n.Load() != 2 {
		t.Fatalf("n=%d after restart", n.Load())
	}
}
