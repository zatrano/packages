# workflow

Generic **process execution**: deterministic executor graphs. A workflow may contain functions, HTTP or database work, application-owned RAG calls, agents, and a **top-level** human wait. Nodes are **not** required to be LLM agents.

**Stability:** experimental. Import-only library (`LayerIntelligence`, no `package:enable`). Does **not** import `ai`, `rag`, or `agent`. There is no `RetrieveExecutor`.

```go
g := workflow.Sequential("docs",
    workflow.Named("fetch", fetchFn),
    workflow.Named("write", writeFn),
)
res, err := g.Run(ctx, workflow.Envelope{Task: "summarize", Input: src})
```

Agents enter a workflow with `agent.AsExecutor`. Do not put HTTP/SQL/human steps on `agent.Graph`.

## Human-in-the-loop

`Approval` / `InputWait` are **in-process pause/resume** (`StatusWaiting` + `Resume` + `Envelope.Decision`).

They are **not** durable workflow resume. They do **not** continue after process crash, restart, or machine failure. `MemoryCheckpointer` is process-local.

HITL is supported only as a **top-level Graph node**. Nested waits inside `Parallel`, `Branch` inner graphs, or `Supervisor` workers return `ErrNestedWait` (a failure, not a resumable wait).

`Resume` validates execution identity (`ClaimWaiting`): unknown id, not waiting, completed, cancelled, or a second Resume fail deterministically.

## Timeouts

The tightest `context` deadline wins (standard Go).

| Knob | Effect |
|------|--------|
| `WithTimeout` on `Run`/`Resume` | Workflow deadline → `context.DeadlineExceeded`, `StatusFailed` |
| `Node.Timeout` | Step deadline (child of the workflow ctx) → step failure, not `StatusCancelled` |
| `context.WithCancel` | `StatusCancelled` |
| HITL paused interval | Does **not** hold a goroutine; workflow timeout does not cover wall-clock time spent waiting for a human |

## Parallel

Fan-out/fan-in via `toolkit/concurrency.Pool`. `err == nil` means the part **started and succeeded**. Unstarted parts after cancel/timeout are `ErrNeverStarted` (wrapped with `context` error), never silent success. Multi-error results are `*PartialError` with sorted names; sibling envelopes are kept in `Parts`. Duplicate part names are uniquified.

## Pieces

| Type | Role |
|------|------|
| `Envelope` | Typed handoff (`Decision` for HITL; `Metadata` is auxiliary only) |
| `Executor` / `Named` | One unit of work |
| `Graph` | Named nodes + `Route` edges |
| `Sequential` | Ordered executors |
| `Parallel` | Fan-out / fan-in; `PartialError` on child failure |
| `Branch` | Condition `(Envelope) (string, error)` → sub-graph |
| `Supervisor` | Router picks workers via `Envelope.Route`. Generic; not orchestration |
| `Approval` / `InputWait` | Top-level in-process pause/resume |
| `Checkpointer` / `MemoryCheckpointer` | In-process pause/resume state (not crash-safe) |
| `WithTimeout` | Workflow deadline (distinct from `Node.Timeout`) |

MCP and A2A are **not** implemented here.
