package agent_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/zatrano/packages/agent"
	"github.com/zatrano/packages/ai"
	"github.com/zatrano/packages/rag"
	"github.com/zatrano/packages/workflow"
)

func TestAgentToolLoop(t *testing.T) {
	mgr := ai.New()
	reg := agent.NewRegistry()
	params := json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`)
	err := reg.Register(ai.FunctionTool("lookup", "Lookup", params), func(ctx context.Context, call ai.ToolCall) (string, error) {
		var args struct {
			Query string `json:"query"`
		}
		_ = call.UnmarshalArguments(&args)
		return `{"ok":true,"q":"` + args.Query + `"}`, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	mem := &agent.BufferMemory{}
	a := &agent.Agent{
		Chat:     agent.FromManager(mgr),
		Tools:    reg,
		Memory:   mem,
		System:   "You are a helpful assistant.",
		MaxSteps: 3,
	}
	res, err := a.Run(context.Background(), "find widgets")
	if err != nil {
		t.Fatal(err)
	}
	if res.Steps < 2 {
		t.Fatalf("steps=%d", res.Steps)
	}
	if res.Response.HasToolCalls() {
		t.Fatal("expected final text")
	}
	if !strings.Contains(res.Response.Message.Content, "stub") {
		t.Fatalf("%q", res.Response.Message.Content)
	}
	// transcript includes tool result
	foundTool := false
	for _, m := range res.Messages {
		if m.Role == "tool" {
			foundTool = true
		}
	}
	if !foundTool {
		t.Fatalf("%+v", res.Messages)
	}
	if len(res.ToolResults) == 0 || res.ToolResults[0].Status != agent.ToolOK {
		t.Fatalf("tool results=%+v", res.ToolResults)
	}
}

func TestAgentWithRetriever(t *testing.T) {
	mgr := ai.New()
	a := &agent.Agent{
		Chat: agent.FromManager(mgr),
		Retrieve: agent.FuncRetriever(func(ctx context.Context, q string) (string, error) {
			return "[1] ZATRANO uses profiles for AI routing.", nil
		}),
		MaxSteps: 1,
	}
	res, err := a.Run(context.Background(), "profiles?")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Messages[0].Content, "Context:") {
		t.Fatalf("%+v", res.Messages[0])
	}
}

func TestRAGRetrieve(t *testing.T) {
	mgr := ai.New()
	p := &rag.Pipeline{
		Chunker: rag.TextChunker{Size: 200, MinChars: 1},
		Embed:   rag.FromAI(mgr, ""),
		Store:   rag.NewMemoryStore(),
	}
	_ = p.Index(context.Background(), rag.Document{ID: "d", Text: "Agents call tools in a loop."})
	r := agent.RAGRetrieve{Pipeline: p, TopK: 1}
	s, err := r.Retrieve(context.Background(), "tools")
	if err != nil || !strings.Contains(s, "score=") {
		t.Fatalf("%v %q", err, s)
	}
}

func TestBufferMemoryTrim(t *testing.T) {
	m := &agent.BufferMemory{MaxKeep: 3}
	m.Append(ai.Message{Role: "system", Content: "sys"})
	m.Append(ai.Message{Role: "user", Content: "1"})
	m.Append(ai.Message{Role: "assistant", Content: "2"})
	m.Append(ai.Message{Role: "user", Content: "3"})
	m.Append(ai.Message{Role: "assistant", Content: "4"})
	msgs := m.Messages()
	if len(msgs) != 3 || msgs[0].Role != "system" {
		t.Fatalf("%+v", msgs)
	}
}

func TestAgentAuthorizerDeniesTool(t *testing.T) {
	mgr := ai.New()
	reg := agent.NewRegistry()
	ran := false
	params := json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`)
	if err := reg.Register(ai.FunctionTool("lookup", "Lookup", params), func(ctx context.Context, call ai.ToolCall) (string, error) {
		ran = true
		return `{"ok":true}`, nil
	}); err != nil {
		t.Fatal(err)
	}
	a := &agent.Agent{
		Chat:     agent.FromManager(mgr),
		Tools:    reg,
		MaxSteps: 3,
		Authorizer: agent.FuncAuthorizer(func(ctx context.Context, name string, call ai.ToolCall) error {
			return fmt.Errorf("not allowed")
		}),
	}
	res, err := a.Run(context.Background(), "find widgets")
	if err != nil {
		t.Fatal(err)
	}
	if ran {
		t.Fatal("handler ran despite Authorizer")
	}
	found := false
	for _, tr := range res.ToolResults {
		if tr.Status == agent.ToolDenied {
			found = true
		}
	}
	if !found {
		t.Fatalf("tool results=%+v", res.ToolResults)
	}
}

func TestAgentAuthorizerAllowsTool(t *testing.T) {
	mgr := ai.New()
	reg := agent.NewRegistry()
	ran := false
	params := json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`)
	if err := reg.Register(ai.FunctionTool("lookup", "Lookup", params), func(ctx context.Context, call ai.ToolCall) (string, error) {
		ran = true
		return `{"ok":true}`, nil
	}); err != nil {
		t.Fatal(err)
	}
	a := &agent.Agent{
		Chat:     agent.FromManager(mgr),
		Tools:    reg,
		MaxSteps: 3,
		Authorizer: agent.FuncAuthorizer(func(ctx context.Context, name string, call ai.ToolCall) error {
			return nil
		}),
	}
	res, err := a.Run(context.Background(), "find widgets")
	if err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Fatal("handler did not run")
	}
	if len(res.ToolResults) == 0 || res.ToolResults[0].Status != agent.ToolOK {
		t.Fatalf("%+v", res.ToolResults)
	}
}

func TestAsExecutorUsesAuthorizer(t *testing.T) {
	mgr := ai.New()
	reg := agent.NewRegistry()
	ran := false
	params := json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`)
	_ = reg.Register(ai.FunctionTool("lookup", "Lookup", params), func(ctx context.Context, call ai.ToolCall) (string, error) {
		ran = true
		return `{"ok":true}`, nil
	})
	a := &agent.Agent{
		Chat:  agent.FromManager(mgr),
		Tools: reg,
		Authorizer: agent.FuncAuthorizer(func(ctx context.Context, name string, call ai.ToolCall) error {
			return fmt.Errorf("deny")
		}),
		MaxSteps: 3,
	}
	_, err := workflow.Sequential("w", agent.AsExecutor("bot", a)).Run(context.Background(), workflow.Envelope{Input: "find widgets"})
	if err != nil {
		t.Fatal(err)
	}
	if ran {
		t.Fatal("tool bypassed authorizer via AsExecutor")
	}
}

func TestAgentCancelDuringChat(t *testing.T) {
	started := make(chan struct{})
	chat := chatterFunc(func(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	a := &agent.Agent{Chat: chat, MaxSteps: 2}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := a.Run(ctx, "hello")
		done <- err
	}()
	<-started
	cancel()
	err := <-done
	if err == nil {
		t.Fatal("expected cancel")
	}
}

func TestAsExecutorHandoff(t *testing.T) {
	mgr := ai.New()
	a := &agent.Agent{Chat: agent.FromManager(mgr), MaxSteps: 1}
	g := workflow.Sequential("w", agent.AsExecutor("bot", a))
	res, err := g.Run(context.Background(), workflow.Envelope{Task: "ask", Input: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(res.Envelope.Output) == "" {
		t.Fatal("empty output")
	}
}

func TestRAGRetrieveRerank(t *testing.T) {
	mgr := ai.New()
	p := &rag.Pipeline{
		Chunker: rag.TextChunker{Size: 200, MinChars: 1},
		Embed:   rag.FromAI(mgr, ""),
		Store:   rag.NewMemoryStore(),
	}
	_ = p.Index(context.Background(), rag.Document{ID: "d", Text: "Agents call tools in a loop."})
	r := agent.RAGRetrieve{Pipeline: p, TopK: 1, Rerank: rag.KeywordReranker{}}
	s, err := r.Retrieve(context.Background(), "tools")
	if err != nil || !strings.Contains(s, "score=") {
		t.Fatalf("%v %q", err, s)
	}
}

type chatterFunc func(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error)

func (f chatterFunc) Chat(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	return f(ctx, req)
}
