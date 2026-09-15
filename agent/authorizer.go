package agent

import (
	"context"

	"github.com/zatrano/packages/ai"
)

// Authorizer decides whether a tool may run.
// Nil on Agent allows every registered tool (compatibility default; set an Authorizer in production).
type Authorizer interface {
	Allow(ctx context.Context, name string, call ai.ToolCall) error
}

// FuncAuthorizer adapts a function to Authorizer.
type FuncAuthorizer func(ctx context.Context, name string, call ai.ToolCall) error

// Allow implements Authorizer.
func (f FuncAuthorizer) Allow(ctx context.Context, name string, call ai.ToolCall) error {
	if f == nil {
		return nil
	}
	return f(ctx, name, call)
}
