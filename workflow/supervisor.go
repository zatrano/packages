package workflow

import (
	"context"
	"fmt"
	"strings"
)

type supervisorExec struct {
	name      string
	router    Executor
	workers   map[string]Executor
	maxRounds int
}

// Supervisor is a strategy: the router executor selects the next worker by
// Envelope.Route. Empty Route ends the loop. Output is the work product, not a hop name.
// Nested human waits in workers are rejected. It is not the definition of orchestration.
func Supervisor(name string, router Executor, workers map[string]Executor, maxRounds int) Executor {
	if maxRounds <= 0 {
		maxRounds = 8
	}
	return &supervisorExec{
		name:      strings.TrimSpace(name),
		router:    router,
		workers:   workers,
		maxRounds: maxRounds,
	}
}

func (s *supervisorExec) Name() string { return s.name }

func (s *supervisorExec) Execute(ctx context.Context, env Envelope) (Envelope, error) {
	if s == nil || s.router == nil {
		return env, fmt.Errorf("%w: supervisor %q missing router", ErrInvalid, s.name)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for round := 0; round < s.maxRounds; round++ {
		if err := ctx.Err(); err != nil {
			return env, err
		}
		routed, err := s.router.Execute(ctx, env.clone())
		if err != nil {
			return routed, rejectNestedWait(s.router.Name(), err)
		}
		next := strings.TrimSpace(routed.Route)
		if next == "" {
			return routed, nil
		}
		w, ok := s.workers[next]
		if !ok || w == nil {
			return routed, fmt.Errorf("%w: supervisor %q worker %q", ErrUnknownRoute, s.name, next)
		}
		routed.Route = ""
		env, err = w.Execute(ctx, routed)
		if err != nil {
			return env, rejectNestedWait(w.Name(), err)
		}
	}
	return env, fmt.Errorf("%w: supervisor %q (%d)", ErrMaxHops, s.name, s.maxRounds)
}
