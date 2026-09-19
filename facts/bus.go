package facts

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Bus routes typed Facts to registered Reactions.
// Go methods cannot be generic, so On[T] is a package function that takes *Bus.
// That keeps registration on the capability instance (From(app)), not a global.
type Bus struct {
	mu          sync.RWMutex
	bindings    map[reflect.Type][]Binding
	queue       Queue
	workers     int
	maxAttempts int
	backoff     time.Duration
	stopWait    time.Duration

	started atomic.Bool
	stop    context.CancelFunc
	wg      sync.WaitGroup
	seq     atomic.Uint64
}

// Option configures a Bus.
type Option func(*Bus)

// WithQueue replaces the default in-process queue.
func WithQueue(q Queue) Option {
	return func(b *Bus) {
		if q != nil {
			b.queue = q
		}
	}
}

// WithWorkers sets async worker concurrency. Values below 1 become 1.
func WithWorkers(n int) Option {
	return func(b *Bus) {
		if n < 1 {
			n = 1
		}
		b.workers = n
	}
}

// WithRetry sets async max attempts (including the first run) and backoff between retries.
func WithRetry(maxAttempts int, backoff time.Duration) Option {
	return func(b *Bus) {
		if maxAttempts < 1 {
			maxAttempts = 1
		}
		b.maxAttempts = maxAttempts
		if backoff < 0 {
			backoff = 0
		}
		b.backoff = backoff
	}
}

// WithStopTimeout bounds how long Stop waits for in-flight Reactions.
func WithStopTimeout(d time.Duration) Option {
	return func(b *Bus) {
		if d > 0 {
			b.stopWait = d
		}
	}
}

// WithQueueSize sets the in-process queue buffer when no custom Queue is given.
func WithQueueSize(n int) Option {
	return func(b *Bus) {
		if n < 1 {
			n = 1
		}
		if _, ok := b.queue.(*MemoryQueue); ok || b.queue == nil {
			b.queue = NewMemoryQueue(n)
		}
	}
}

// New builds a Bus with an in-process async queue.
func New(opts ...Option) *Bus {
	b := &Bus{
		bindings:    make(map[reflect.Type][]Binding),
		queue:       NewMemoryQueue(256),
		workers:     4,
		maxAttempts: 3,
		backoff:     time.Second,
		stopWait:    15 * time.Second,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(b)
		}
	}
	if b.queue == nil {
		b.queue = NewMemoryQueue(256)
	}
	return b
}

// On registers Reactions for Fact type T. Registration is explicit; there is
// no package scanning. reflect.Type is used only as the map key for mixed Fact
// types on one Bus — not for discovery.
func On[T any](bus *Bus, reactions ...Binding) {
	if bus == nil {
		return
	}
	typ := typeOf[T]()
	bus.mu.Lock()
	defer bus.mu.Unlock()
	for _, r := range reactions {
		if r.call == nil {
			continue
		}
		bus.bindings[typ] = append(bus.bindings[typ], r)
	}
}

// Publish runs Sync Reactions and enqueues Async Reactions for type(fact).
// It waits for Sync completion and Async acceptance, not for worker completion.
// Independent Reactions all run; errors are joined.
func (b *Bus) Publish(ctx context.Context, fact any) error {
	if b == nil || fact == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	typ := reflect.TypeOf(fact)
	b.mu.RLock()
	regs := append([]Binding{}, b.bindings[typ]...)
	b.mu.RUnlock()
	if len(regs) == 0 {
		return nil
	}

	var errs []error
	for _, r := range regs {
		if r.async {
			continue
		}
		if err := r.call(ctx, fact); err != nil {
			errs = append(errs, err)
		}
	}
	for _, r := range regs {
		if !r.async {
			continue
		}
		job := &Job{
			ID:          b.seq.Add(1),
			FactType:    typ.String(),
			Fact:        fact,
			Attempt:     1,
			MaxAttempts: b.maxAttempts,
			call:        r.call,
		}
		if err := b.queue.Enqueue(ctx, job); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Start launches async workers. Safe to call again after Stop (kernel Start retry).
// Do not call from Register/Boot; the kernel LifecycleProvider.Start hook owns this.
func (b *Bus) Start() error {
	if b == nil {
		return nil
	}
	if !b.started.CompareAndSwap(false, true) {
		return nil
	}
	if r, ok := b.queue.(Reopenable); ok {
		if err := r.Reopen(); err != nil {
			b.started.Store(false)
			return err
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	b.stop = cancel
	n := b.workers
	if n < 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		b.wg.Add(1)
		go b.worker(ctx)
	}
	return nil
}

// Stop closes the queue, stops workers, and waits up to the configured timeout
// (or ctx) for in-flight Reactions. Start may be called again afterwards.
func (b *Bus) Stop(ctx context.Context) error {
	if b == nil || !b.started.Load() {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	_ = b.queue.Close()
	if b.stop != nil {
		b.stop()
	}
	wait := b.stopWait
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < wait {
			wait = remaining
		}
	}
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-done:
		b.started.Store(false)
		return nil
	case <-timer.C:
		b.started.Store(false)
		return fmt.Errorf("facts: stop timed out after %s", wait)
	case <-ctx.Done():
		b.started.Store(false)
		return ctx.Err()
	}
}

func typeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
