# ZATRANO AI — Public API Design

**Date:** 2026-09-11  
**Rule:** small, typed, explicit, composable, testable, idiomatic Go.  
**Not implemented in this file.** This is the contract the rebuild must follow.

Existing `ai` / `rag` / `agent` APIs that remain valid are restated only where the target changes them.

---

## 1. `workflow` (new library)

### `Envelope`

```go
type Envelope struct {
    Task        string
    Goal        string
    Input       string
    Previous    string
    Output      string
    Required    string
    Artifacts   map[string]string
    Constraints []string
    Permissions []string
    Metadata    map[string]string // routing scratch only; not a generic bag of domain objects
}
```

| | |
|---|---|
| Purpose | Structured context transfer between steps |
| Responsibilities | Carry task intent and artifacts without a full transcript |
| Non-responsibilities | Conversation history, vector hits, secrets, `any` payloads |
| Dependencies | None |
| Lifecycle | Value copied per hop |
| Concurrency | Immutable-by-convention after Execute returns; callers copy maps if they retain them |
| Cancellation | N/A |
| Error model | N/A |
| Extensibility | Add fields only for shared semantics; app-specific data → Artifacts |
| Example | Handoff from classifier to specialist: `Task="refund"`, `Input=user text`, `Constraints=["no-pii-in-logs"]` |

### `Executor`

```go
type Executor interface {
    Name() string
    Execute(ctx context.Context, env Envelope) (Envelope, error)
}
```

| | |
|---|---|
| Purpose | One unit of work in a workflow |
| Responsibilities | Do work; return updated envelope or error; honor ctx |
| Non-responsibilities | Routing, retries of siblings, persistence |
| Dependencies | None at the interface |
| Lifecycle | Typically stateless; may close over app services |
| Concurrency | Must be safe for the graph’s concurrency model (Parallel may call concurrently) |
| Cancellation | Return `ctx.Err()` promptly |
| Error model | Returned error fails the hop unless a `WaitError` |
| Extensibility | `Func` adapter; `agent.AsExecutor` |
| Example | `workflow.Func("http", callAPI)` |

`WaitError` signals `StatusWaiting` (human input/approval). It is not a failure.

### `Func`

```go
type Func func(ctx context.Context, env Envelope) (Envelope, error)
```

Named via `Named(name string, fn Func) Executor`.

### `Graph`

```go
type Node struct {
    Exec    Executor
    Timeout time.Duration // 0 = inherit / none
    Route   func(ctx context.Context, env Envelope) (next string, error)
}

type Graph struct {
    ID      string
    Start   string
    Nodes   map[string]Node
    MaxHops int
}

func (g *Graph) Run(ctx context.Context, env Envelope, opt ...Option) (*Result, error)
func (g *Graph) Resume(ctx context.Context, executionID string, decision Decision, opt ...Option) (*Result, error)
```

| | |
|---|---|
| Purpose | Deterministic executor graph |
| Responsibilities | Walk nodes, apply timeouts, record traces, checkpoint if configured, stop on wait/fail/cancel |
| Non-responsibilities | LLM tool loops; provider HTTP; being an agent |
| Dependencies | `Executor`; optional `Checkpointer` |
| Lifecycle | Graph is definition (reusable). Each `Run` is an execution |
| Concurrency | One walk per `Run` unless a node is `Parallel` |
| Cancellation | `ctx` on `Run`; propagated to `Execute` |
| Error model | `Result.Status` + error; partial `Trace` retained |
| Extensibility | Any Executor; Route functions |
| Example | Start `classify` → Route to `refund` or `support` → `notify` |

### Constructors

```go
func Sequential(id string, steps ...Executor) *Graph
func Parallel(id string, join Joiner, parts ...Executor) Executor
func Branch(id string, cond func(Envelope) string, paths map[string]*Graph, otherwise *Graph) *Graph
func Supervisor(name string, router Executor, workers map[string]Executor, maxRounds int) Executor
```

`Parallel` and `Supervisor` are **Executors** (one hop) so they sit inside a Graph without a second engine. Supervisor routing uses `Envelope.Route` only (`Output` is the work product).

`Joiner`:

```go
type Joiner func(ctx context.Context, parent Envelope, parts map[string]Envelope) (Envelope, error)
```

Failed parts: `Parallel` always returns `*PartialError` (sibling envelopes retained in `Parts`). There is no silent discard of successful children.

### `Result` / `Status` / `Trace`

```go
type Status string // created|running|waiting|completed|failed|cancelled

type Result struct {
    ExecutionID string
    WorkflowID  string
    Status      Status
    Envelope    Envelope
    Traces      []Trace
    Wait        *Wait
}

type Trace struct {
    StepID   string
    Name     string
    Duration time.Duration
    Err      error
}

type Wait struct {
    Kind   string // "approval" | "input"
    StepID string
    Reason string
}

type Decision struct {
    Approved bool
    Input    string
    Comment  string
}
```

### Checkpoint

```go
type Checkpoint struct {
    ExecutionID string
    WorkflowID  string
    StepID      string
    Status      Status
    Envelope    Envelope
    Path        []string
}

type Checkpointer interface {
    Save(ctx context.Context, cp Checkpoint) error
    Load(ctx context.Context, executionID string) (Checkpoint, error)
}
```

`MemoryCheckpointer` for tests and ephemeral wait/resume in-process.

### Options

```go
func WithTimeout(d time.Duration) Option
func WithCheckpointer(c Checkpointer) Option
func WithExecutionID(id string) Option
func WithObserver(o Observer) Option
```

Workflow timeout wraps `Run`. Step timeout wraps `Execute`. They are not the same value.

### Observer

```go
type Observer interface {
    OnHop(ctx context.Context, t Trace)
}
```

Do not invent OpenTelemetry types here. Apps may forward.

---

## 2. `agent` (existing library, hardened)

### `Agent` (kept, additive)

```go
type Agent struct {
    Chat     Chatter
    Tools    *Registry
    Memory   Memory
    Retrieve Retriever
    System   string
    MaxSteps int
    Timeout  time.Duration // 0 = none (beyond ctx)
    Authorizer Authorizer  // nil = allow all registered tools (documented unsafe default)
    AllowToolsOnFinal bool
}
```

| | |
|---|---|
| Purpose | Single-agent chat↔tool loop |
| Responsibilities | Loop, retrieval injection, tool dispatch, termination |
| Non-responsibilities | Multi-agent graphs, durable workflow, provider HTTP |
| Dependencies | `ai` types; optional `rag` via Retriever; optional `workflow` via `AsExecutor` |
| Lifecycle | `Run` scoped |
| Concurrency | One `Run` at a time per shared Memory unless the Memory impl is safe |
| Cancellation | Check `ctx.Err()` each step; pass ctx to Chat and tools |
| Error model | Retrieve/Chat errors fail; tool errors become tool messages; max steps returns result+error |
| Extensibility | Custom Chatter, Memory, Retriever, Authorizer |
| Example | See current README; add `Authorizer` in production |

### `Authorizer`

```go
type Authorizer interface {
    Allow(ctx context.Context, name string, call ai.ToolCall) error
}
```

Denied → `ToolDenied` without running the handler. Nil authorizer keeps today’s behavior (compatibility) but README must call it out.

### `AsExecutor`

```go
func AsExecutor(name string, a *Agent) workflow.Executor
```

Maps `env.Input` (fallback `env.Previous`) to `Run`. Writes `env.Output` from the assistant text. Does not copy the full transcript into the envelope.

### `Chain` / `Graph` (compatibility)

Remain exported. Internals should call `workflow` where practical. New application code should prefer `workflow` + `AsExecutor`.

---

## 3. `ai` (existing service — no API break required)

Public surface stays: `Manager`, `Client`, `Driver`, `Chat`/`Embed`/`ChatStream`, profiles, `Observer`, `UsageMeter`.

Non-responsibilities stay: executing tools, running agents, storing vectors.

---

## 4. `rag` (existing library — small additive)

Keep `Pipeline`, stores, rerankers.

`agent.RAGRetrieve` should gain optional `Rerank bool` or `Options` so `QueryWith` is reachable without a custom `FuncRetriever`.

---

## 5. Queue

No new queue API. Optional later: `workflow` job payload analogous to `agent.RunJob`. Not required to establish the boundary.

---

## 6. What is not public

- MCP client/server types
- A2A Agent Card / Task protocol
- `map[string]any` workflow state
- Global agent registry in `ai.Manager`
- `app.Workflow()` on the kernel
