package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/zatrano/packages/workflow"
)

type agentExecutor struct {
	name string
	a    *Agent
}

// AsExecutor exposes an Agent as a workflow.Executor.
// Only Envelope.Input (fallback Previous) is sent to Run; the transcript stays inside the agent.
func AsExecutor(name string, a *Agent) workflow.Executor {
	return agentExecutor{name: strings.TrimSpace(name), a: a}
}

func (e agentExecutor) Name() string { return e.name }

func (e agentExecutor) Execute(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
	if e.a == nil {
		return env, fmt.Errorf("agent: AsExecutor %q has nil Agent", e.name)
	}
	msg := strings.TrimSpace(env.Input)
	if msg == "" {
		msg = strings.TrimSpace(env.Previous)
	}
	if msg == "" {
		return env, fmt.Errorf("agent: empty envelope input")
	}
	res, err := e.a.Run(ctx, msg)
	if res != nil && res.Response != nil {
		env.Output = res.Response.Message.Content
		env.Previous = env.Output
	}
	return env, err
}
