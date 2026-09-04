package agent

import (
	"context"

	"github.com/zatrano/packages/ai"
)

// Chatter is satisfied by *ai.Client (Using / Profile).
// Use FromManager to adapt *ai.Manager (whose Chat has a variadic provider arg).
type Chatter interface {
	Chat(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error)
}

// FromManager adapts *ai.Manager to Chatter (default / profile-less Chat).
func FromManager(m *ai.Manager) Chatter {
	return chatterFunc(func(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
		return m.Chat(ctx, req)
	})
}

type chatterFunc func(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error)

func (f chatterFunc) Chat(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	return f(ctx, req)
}

// Memory stores conversation turns for multi-step agent runs.
type Memory interface {
	Messages() []ai.Message
	Append(msgs ...ai.Message)
	Clear()
}

// Retriever optionally injects RAG context before the user turn.
type Retriever interface {
	Retrieve(ctx context.Context, query string) (string, error)
}

// Handler executes a single tool call and returns a string result for the model.
type Handler func(ctx context.Context, call ai.ToolCall) (string, error)

// ToolStatus classifies a tool execution outcome.
type ToolStatus string

const (
	ToolOK      ToolStatus = "ok"
	ToolError   ToolStatus = "error"
	ToolTimeout ToolStatus = "timeout"
	ToolDenied  ToolStatus = "denied"
	ToolInvalid ToolStatus = "invalid"
)

// ToolResult is a typed outcome of one tool call (status, retryability, error).
type ToolResult struct {
	ID        string
	Name      string
	Status    ToolStatus
	Output    string
	Error     string
	Retryable bool
	Metadata  map[string]string
}

// Content is the string sent back to the model as the tool message body.
func (r ToolResult) Content() string {
	if r.Status == ToolOK {
		return r.Output
	}
	if r.Output != "" {
		return r.Output
	}
	msg := r.Error
	if msg == "" {
		msg = string(r.Status)
		if msg == "" {
			msg = "tool failed"
		}
	}
	return "error: " + msg
}

// ExecError lets a handler set status and retryability instead of a bare error.
type ExecError struct {
	Status    ToolStatus
	Message   string
	Retryable bool
}

func (e *ExecError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return string(e.Status)
}

// Result is the outcome of Agent.Run.
type Result struct {
	Response    *ai.ChatResponse
	Steps       int
	Messages    []ai.Message // full transcript including tools
	ToolResults []ToolResult
}
