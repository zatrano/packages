package facts

import (
	"context"
	"errors"
	"sync"
)

// Queue is the async execution transport. The in-process implementation is
// the default. Other transports (Redis, database, brokers, Outbox relay)
// satisfy this without changing Reaction[T] or On/Publish.
type Queue interface {
	Enqueue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context) (*Job, error)
	Close() error
}

// Reopenable is optional. MemoryQueue implements it so kernel Start retry
// after Stop can accept Async work again. Durable adapters reopen their own way.
type Reopenable interface {
	Reopen() error
}

// Job is executable async work. It is not a Fact and is not part of the
// application-facing API. Attempt metadata is here so an Outbox/durable worker
// can persist it without changing On/Publish.
type Job struct {
	ID          uint64
	FactType    string
	Fact        any
	Attempt     int
	MaxAttempts int
	call        func(context.Context, any) error
	LastError   error
}

var (
	errQueueClosed = errors.New("facts: async queue closed")
	errNilJob      = errors.New("facts: nil job")
)

// MemoryQueue is an in-process bounded queue. A full buffer applies backpressure
// by blocking Enqueue until space exists, the context ends, or the queue closes.
type MemoryQueue struct {
	mu     sync.Mutex
	size   int
	ch     chan *Job
	done   chan struct{}
	closed bool
}

// NewMemoryQueue builds a buffered in-process queue.
func NewMemoryQueue(size int) *MemoryQueue {
	if size < 1 {
		size = 1
	}
	return &MemoryQueue{
		size: size,
		ch:   make(chan *Job, size),
		done: make(chan struct{}),
	}
}

func (q *MemoryQueue) snapshot() (ch chan *Job, done chan struct{}, closed bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.ch, q.done, q.closed
}

// Enqueue accepts work. Failure here is returned from Publish.
func (q *MemoryQueue) Enqueue(ctx context.Context, job *Job) error {
	if q == nil {
		return errQueueClosed
	}
	if job == nil {
		return errNilJob
	}
	ch, done, closed := q.snapshot()
	if closed {
		return errQueueClosed
	}
	select {
	case <-done:
		return errQueueClosed
	case ch <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Dequeue waits for the next job or context/queue shutdown.
func (q *MemoryQueue) Dequeue(ctx context.Context) (*Job, error) {
	if q == nil {
		return nil, errQueueClosed
	}
	ch, done, closed := q.snapshot()
	if closed {
		select {
		case job := <-ch:
			if job == nil {
				return nil, errQueueClosed
			}
			return job, nil
		default:
			return nil, errQueueClosed
		}
	}
	select {
	case job := <-ch:
		if job == nil {
			return nil, errQueueClosed
		}
		return job, nil
	case <-done:
		select {
		case job := <-ch:
			if job == nil {
				return nil, errQueueClosed
			}
			return job, nil
		default:
			return nil, errQueueClosed
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close stops accepting new work. In-flight Dequeue can still drain the buffer.
func (q *MemoryQueue) Close() error {
	if q == nil {
		return nil
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil
	}
	q.closed = true
	close(q.done)
	return nil
}

// Reopen replaces the closed in-process buffer so Start after Stop can run.
func (q *MemoryQueue) Reopen() error {
	if q == nil {
		return errQueueClosed
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.closed {
		return nil
	}
	q.ch = make(chan *Job, q.size)
	q.done = make(chan struct{})
	q.closed = false
	return nil
}
