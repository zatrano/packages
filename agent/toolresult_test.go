package agent_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zatrano/packages/agent"
	"github.com/zatrano/packages/ai"
)

func TestExecuteResultUnknownTool(t *testing.T) {
	reg := agent.NewRegistry()
	res := reg.ExecuteResult(context.Background(), ai.ToolCall{
		ID:       "c1",
		Function: ai.ToolCallFunction{Name: "missing"},
	})
	if res.Status != agent.ToolInvalid || res.Retryable {
		t.Fatalf("%+v", res)
	}
	if res.ID != "c1" || !strings.Contains(res.Error, "unknown tool") {
		t.Fatalf("%+v", res)
	}
	if !strings.HasPrefix(res.Content(), "error:") {
		t.Fatalf("content=%q", res.Content())
	}
}

func TestExecuteResultTimeout(t *testing.T) {
	reg := agent.NewRegistry()
	err := reg.Register(ai.FunctionTool("slow", "Slow", nil), func(ctx context.Context, call ai.ToolCall) (string, error) {
		return "", context.DeadlineExceeded
	})
	if err != nil {
		t.Fatal(err)
	}
	res := reg.ExecuteResult(context.Background(), ai.ToolCall{Function: ai.ToolCallFunction{Name: "slow"}})
	if res.Status != agent.ToolTimeout || !res.Retryable {
		t.Fatalf("%+v", res)
	}
}

func TestExecuteResultCanceled(t *testing.T) {
	reg := agent.NewRegistry()
	err := reg.Register(ai.FunctionTool("wait", "Wait", nil), func(ctx context.Context, call ai.ToolCall) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(5 * time.Millisecond)
	res := reg.ExecuteResult(ctx, ai.ToolCall{Function: ai.ToolCallFunction{Name: "wait"}})
	if res.Status != agent.ToolTimeout || !res.Retryable {
		t.Fatalf("%+v", res)
	}
}

func TestExecuteResultExecError(t *testing.T) {
	reg := agent.NewRegistry()
	err := reg.Register(ai.FunctionTool("lock", "Lock", nil), func(ctx context.Context, call ai.ToolCall) (string, error) {
		return "", &agent.ExecError{Status: agent.ToolDenied, Message: "nope"}
	})
	if err != nil {
		t.Fatal(err)
	}
	res := reg.ExecuteResult(context.Background(), ai.ToolCall{Function: ai.ToolCallFunction{Name: "lock"}})
	if res.Status != agent.ToolDenied || res.Retryable || res.Error != "nope" {
		t.Fatalf("%+v", res)
	}
	out, execErr := reg.Execute(context.Background(), ai.ToolCall{Function: ai.ToolCallFunction{Name: "lock"}})
	if execErr == nil || out != "error: nope" {
		t.Fatalf("out=%q err=%v", out, execErr)
	}
}

func TestExecuteResultOK(t *testing.T) {
	reg := agent.NewRegistry()
	err := reg.Register(ai.FunctionTool("echo", "Echo", nil), func(ctx context.Context, call ai.ToolCall) (string, error) {
		return "pong", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	res := reg.ExecuteResult(context.Background(), ai.ToolCall{ID: "1", Function: ai.ToolCallFunction{Name: "echo"}})
	if res.Status != agent.ToolOK || res.Output != "pong" || res.Content() != "pong" {
		t.Fatalf("%+v", res)
	}
}

func TestClassifyWrappedDeadline(t *testing.T) {
	reg := agent.NewRegistry()
	err := reg.Register(ai.FunctionTool("wrap", "Wrap", nil), func(ctx context.Context, call ai.ToolCall) (string, error) {
		return "", errors.Join(context.DeadlineExceeded, errors.New("upstream"))
	})
	if err != nil {
		t.Fatal(err)
	}
	res := reg.ExecuteResult(context.Background(), ai.ToolCall{Function: ai.ToolCallFunction{Name: "wrap"}})
	if res.Status != agent.ToolTimeout || !res.Retryable {
		t.Fatalf("%+v", res)
	}
}

func TestWebFetchDeniedHTTP(t *testing.T) {
	reg := agent.NewRegistry()
	if err := agent.RegisterWebFetch(reg, agent.WebFetchOptions{}); err != nil {
		t.Fatal(err)
	}
	res := reg.ExecuteResult(context.Background(), ai.ToolCall{
		Function: ai.ToolCallFunction{Name: "web_fetch", Arguments: `{"url":"http://example.com"}`},
	})
	if res.Status != agent.ToolDenied {
		t.Fatalf("%+v", res)
	}
}

func TestWebFetchInvalidURL(t *testing.T) {
	reg := agent.NewRegistry()
	if err := agent.RegisterWebFetch(reg, agent.WebFetchOptions{AllowHTTP: true}); err != nil {
		t.Fatal(err)
	}
	res := reg.ExecuteResult(context.Background(), ai.ToolCall{
		Function: ai.ToolCallFunction{Name: "web_fetch", Arguments: `{"url":""}`},
	})
	if res.Status != agent.ToolInvalid {
		t.Fatalf("%+v", res)
	}
}
