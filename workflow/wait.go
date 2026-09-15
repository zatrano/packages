package workflow

import (
	"context"
	"fmt"
	"strings"
)

type waitExec struct {
	name   string
	kind   string
	reason string
}

// Approval pauses the workflow until Resume with Decision.Approved.
// This is in-process pause/resume, not crash-safe durable execution.
func Approval(name, reason string) Executor {
	return &waitExec{name: strings.TrimSpace(name), kind: "approval", reason: reason}
}

// InputWait pauses until Resume with Decision.Approved and Input.
func InputWait(name, reason string) Executor {
	return &waitExec{name: strings.TrimSpace(name), kind: "input", reason: reason}
}

func (w *waitExec) Name() string { return w.name }

func (w *waitExec) Execute(ctx context.Context, env Envelope) (Envelope, error) {
	if err := ctx.Err(); err != nil {
		return env, err
	}
	if env.Decision == nil {
		return env, &WaitError{Wait: Wait{Kind: w.kind, Reason: w.reason}}
	}
	if !env.Decision.Approved {
		return env, fmt.Errorf("%w: %s", ErrRejected, w.kind)
	}
	if w.kind == "input" {
		input := strings.TrimSpace(env.Decision.Input)
		if input == "" {
			return env, fmt.Errorf("%w: input wait %q resumed without Input", ErrInvalid, w.name)
		}
		env.Input = input
		return env, nil
	}
	if input := strings.TrimSpace(env.Decision.Input); input != "" {
		env.Input = input
	}
	return env, nil
}
