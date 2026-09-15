// Package agent is an import-only library (no addon registration).
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zatrano/packages/ai"
)

// Agent runs a chat↔tool loop with optional memory and RAG retrieval.
type Agent struct {
	Chat     Chatter
	Tools    *Registry
	Memory   Memory // optional; defaults to ephemeral BufferMemory per Run
	Retrieve Retriever
	System   string        // optional system prompt prepended when memory is empty
	MaxSteps int           // default 6
	Timeout  time.Duration // 0 = none beyond ctx; distinct from ai.Defaults.Timeout
	// Authorizer, when set, must allow a tool before the handler runs.
	// Nil allows every registered tool (compatibility default; unsafe for untrusted models).
	Authorizer Authorizer
	// AllowToolsOnFinal keeps tools enabled on the last step (default false:
	// last step uses ToolChoiceNone so the model must produce a text answer).
	AllowToolsOnFinal bool
}

// Run appends the user message and loops until a text reply or MaxSteps.
func (a *Agent) Run(ctx context.Context, userMessage string) (*Result, error) {
	if a == nil || a.Chat == nil {
		return nil, fmt.Errorf("agent: Chat is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if a.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.Timeout)
		defer cancel()
	}
	userMessage = strings.TrimSpace(userMessage)
	if userMessage == "" {
		return nil, fmt.Errorf("agent: user message is required")
	}

	mem := a.Memory
	if mem == nil {
		mem = &BufferMemory{}
	}
	maxSteps := a.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 6
	}

	if len(mem.Messages()) == 0 && strings.TrimSpace(a.System) != "" {
		mem.Append(ai.Message{Role: "system", Content: strings.TrimSpace(a.System)})
	}

	content := userMessage
	if a.Retrieve != nil {
		block, err := a.Retrieve.Retrieve(ctx, userMessage)
		if err != nil {
			return nil, err
		}
		block = strings.TrimSpace(block)
		if block != "" {
			content = "Context:\n" + block + "\n\nQuestion: " + userMessage
		}
	}
	mem.Append(ai.Message{Role: "user", Content: content})

	var tools []ai.Tool
	if a.Tools != nil {
		tools = a.Tools.Tools()
	}

	var last *ai.ChatResponse
	var toolResults []ToolResult
	steps := 0
	for step := 1; step <= maxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return &Result{Response: last, Steps: steps, Messages: mem.Messages(), ToolResults: toolResults}, err
		}
		steps = step
		req := ai.ChatRequest{
			Messages: mem.Messages(),
			Tools:    tools,
		}
		if len(tools) > 0 {
			if step == maxSteps && !a.AllowToolsOnFinal {
				req.ToolChoice = ai.ToolChoiceNone()
			} else {
				req.ToolChoice = ai.ToolChoiceAuto()
			}
		}
		resp, err := a.Chat.Chat(ctx, req)
		if err != nil {
			return nil, err
		}
		last = resp
		if !resp.HasToolCalls() {
			mem.Append(resp.Message)
			return &Result{Response: resp, Steps: steps, Messages: mem.Messages(), ToolResults: toolResults}, nil
		}
		mem.Append(ai.AssistantToolCalls(resp.Message.ToolCalls...))
		if a.Tools == nil {
			return nil, fmt.Errorf("agent: model requested tools but Tools registry is nil")
		}
		for _, call := range resp.Message.ToolCalls {
			tr := a.dispatchTool(ctx, call)
			toolResults = append(toolResults, tr)
			mem.Append(ai.ToolResultMessage(call.ID, tr.Content()))
		}
	}
	if last == nil {
		return nil, fmt.Errorf("agent: no response")
	}
	return &Result{Response: last, Steps: steps, Messages: mem.Messages(), ToolResults: toolResults}, fmt.Errorf("agent: max steps (%d) reached", maxSteps)
}

func (a *Agent) dispatchTool(ctx context.Context, call ai.ToolCall) ToolResult {
	name := strings.TrimSpace(call.Function.Name)
	if a.Authorizer != nil {
		if err := a.Authorizer.Allow(ctx, name, call); err != nil {
			msg := err.Error()
			if msg == "" {
				msg = "denied"
			}
			return ToolResult{ID: call.ID, Name: name, Status: ToolDenied, Error: msg}
		}
	}
	return a.Tools.ExecuteResult(ctx, call)
}
