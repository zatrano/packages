package facts

import (
	"context"
	"errors"
	"time"
)

// Permanent wraps an error so the worker will not retry it.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &permanentError{err}
}

type permanentError struct{ error }

func (e *permanentError) Unwrap() error { return e.error }

func isPermanent(err error) bool {
	var p *permanentError
	return errors.As(err, &p)
}

func (b *Bus) worker(ctx context.Context) {
	defer b.wg.Done()
	for {
		job, err := b.queue.Dequeue(ctx)
		if err != nil {
			return
		}
		b.execute(ctx, job)
	}
}

func (b *Bus) execute(ctx context.Context, job *Job) {
	if job == nil || job.call == nil {
		return
	}
	err := job.call(ctx, job.Fact)
	if err == nil {
		return
	}
	job.LastError = err
	if isPermanent(err) || job.Attempt >= job.MaxAttempts {
		return
	}
	if b.backoff > 0 {
		timer := time.NewTimer(b.backoff * time.Duration(job.Attempt))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
	job.Attempt++
	_ = b.queue.Enqueue(ctx, job)
}
