package workflow

import (
	"context"
	"fmt"
	"strings"
)

// Executor is one unit of work in a workflow.
type Executor interface {
	Name() string
	Execute(ctx context.Context, env Envelope) (Envelope, error)
}

// Func is a function executor body.
type Func func(ctx context.Context, env Envelope) (Envelope, error)

type namedFunc struct {
	name string
	fn   Func
}

// Named wraps fn as an Executor. name is required.
func Named(name string, fn Func) Executor {
	return namedFunc{name: strings.TrimSpace(name), fn: fn}
}

func (n namedFunc) Name() string { return n.name }

func (n namedFunc) Execute(ctx context.Context, env Envelope) (Envelope, error) {
	if strings.TrimSpace(n.name) == "" {
		return env, fmt.Errorf("workflow: executor name is required")
	}
	if n.fn == nil {
		return env, fmt.Errorf("workflow: executor %q is nil", n.name)
	}
	if err := ctx.Err(); err != nil {
		return env, err
	}
	return n.fn(ctx, env)
}
