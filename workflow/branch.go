package workflow

import (
	"context"
	"fmt"
	"strings"
)

// Branch picks a sub-graph by cond(env). cond may return an error (invalid decision).
// Nested human waits inside a selected graph are rejected (ErrNestedWait).
func Branch(id string, cond func(Envelope) (string, error), paths map[string]*Graph, otherwise *Graph) *Graph {
	exec := Named(strings.TrimSpace(id)+"-branch", func(ctx context.Context, env Envelope) (Envelope, error) {
		if err := ctx.Err(); err != nil {
			return env, err
		}
		if cond == nil {
			return env, fmt.Errorf("%w: branch %q missing condition", ErrInvalid, id)
		}
		key, err := cond(env)
		if err != nil {
			return env, err
		}
		key = strings.TrimSpace(key)
		g, ok := paths[key]
		if !ok {
			g = otherwise
		}
		if g == nil {
			return env, fmt.Errorf("%w: branch %q has no path for %q", ErrUnknownRoute, id, key)
		}
		res, err := g.Run(ctx, env)
		if res != nil {
			env = res.Envelope
		}
		if err != nil {
			return env, rejectNestedWait(id, err)
		}
		return env, nil
	})
	return Sequential(id, exec)
}
