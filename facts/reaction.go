package facts

import (
	"context"
	"fmt"
)

// Reaction handles a typed Fact. It must not choose Sync vs Async.
type Reaction[T any] interface {
	React(context.Context, T) error
}

// ReactionFunc adapts a function to Reaction[T].
type ReactionFunc[T any] func(context.Context, T) error

// React calls the adapted function.
func (f ReactionFunc[T]) React(ctx context.Context, fact T) error {
	return f(ctx, fact)
}

// Binding is a Reaction wrapped with an execution policy.
type Binding struct {
	async bool
	call  func(context.Context, any) error
}

// Sync runs the Reaction during Publish, before Publish returns.
func Sync[T any](r Reaction[T]) Binding {
	return bind(r, false)
}

// Async submits the Reaction to the job queue during Publish.
// Execution happens on a worker. This is not a fire-and-forget goroutine.
func Async[T any](r Reaction[T]) Binding {
	return bind(r, true)
}

func bind[T any](r Reaction[T], async bool) Binding {
	if r == nil {
		return Binding{}
	}
	return Binding{
		async: async,
		call: func(ctx context.Context, fact any) error {
			v, ok := fact.(T)
			if !ok {
				return fmt.Errorf("facts: unexpected type %T", fact)
			}
			return r.React(ctx, v)
		},
	}
}
