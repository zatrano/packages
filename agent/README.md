# agent

Multi-step AI agents: chat ↔ tool loop, conversation memory, optional RAG retrieval.

Import-only library (no `package:enable`). Uses `packages/ai` for chat/tools and optionally `packages/rag`.

```go
import (
    "context"
    "encoding/json"

    "github.com/zatrano/packages/agent"
    "github.com/zatrano/packages/ai"
)

mgr := ai.New()
reg := agent.NewRegistry()
_ = reg.Register(ai.FunctionTool("lookup", "Search docs", json.RawMessage(`{
    "type":"object","properties":{"query":{"type":"string"}},"required":["query"]
}`)), func(ctx context.Context, call ai.ToolCall) (string, error) {
    var args struct{ Query string `json:"query"` }
    _ = call.UnmarshalArguments(&args)
    return `{"hits":[]}`, nil
})

a := &agent.Agent{
    Chat:   mgr.Profile("support"), // *ai.Client implements Chatter; or agent.FromManager(mgr)
    Tools:  reg,
    Memory: &agent.BufferMemory{},
    System: "You help with ZATRANO docs.",
}
res, err := a.Run(ctx, "How do profiles work?")
_ = res.Response.Message.Content
```

## Built-in tools

```go
_ = agent.RegisterWebFetch(reg, agent.WebFetchOptions{})
_ = agent.RegisterFileSearch(reg, agent.FileSearchOptions{Root: "./docs", Extensions: []string{".md"}})
```

## Queue

```go
cat := agent.NewCatalog()
_ = cat.Register("support", a)
runner := &agent.Runner{Catalog: cat, Results: agent.NewMemoryResultStore()}
_ = runner.RegisterQueue(queue.From(app))
_ = runner.PushRun(queue.From(app), agent.RunJob{Agent: "support", Message: "hi", ID: "1"})
```

## RAG

```go
a.Retrieve = agent.RAGRetrieve{Pipeline: ragPipeline, TopK: 5}
```

## Pieces

| Type | Role |
|------|------|
| `Agent` | `Run` loop (`MaxSteps`, last-step `ToolChoiceNone` by default) |
| `Registry` | Tool defs + handlers |
| `BufferMemory` | Conversation buffer (`MaxKeep`) |
| `Retriever` / `RAGRetrieve` | Optional context injection |
| `RegisterWebFetch` / `RegisterFileSearch` | HTTPS fetch + sandboxed file search |
| `Catalog` / `Runner` / `PushRun` | Queue-backed `agent.run` jobs |
| `ResultStore` / `MemoryResultStore` / `JSONFileResultStore` | Persist outcomes by job ID |
| `Chain` / `CatalogChain` | Sequential multi-agent runs (agent-only sugar; generic processes use `workflow`) |
| `Graph` / `RouteIf` / `RouteContains` | Branching multi-agent graph (agent-only; generic graphs use `workflow`) |
| `AsExecutor` | Expose an Agent as `workflow.Executor` (envelope in/out, transcript stays inside) |
| `Authorizer` / `FuncAuthorizer` | Tool allow-list before handler execution. **Nil allows every registered tool** (compatibility default). Set an Authorizer in production. |
| `Timeout` | Agent-level deadline (distinct from `ai` request timeout) |
| `Result` | Final response + step count + transcript + typed `ToolResults` |

`RAGRetrieve` is an **agent-side** adapter (`agent` → `rag`). Workflow does not import RAG.

Human approval and generic process graphs belong in [`workflow`](../workflow) (top-level in-process pause/resume, **not** durable execution). MCP/A2A are not implemented.
