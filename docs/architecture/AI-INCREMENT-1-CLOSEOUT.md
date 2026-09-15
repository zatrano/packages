# ZATRANO AI — Increment 1 Closeout

**Date:** 2026-09-12  
**Architecture:** unchanged (frozen).  
**Increment 2:** remains **BLOCKED** until this closeout is accepted.

```text
ai       → model runtime
rag      → retrieval
agent    → single-agent loop
workflow → generic process execution
```

No MCP, A2A, swarm, planner, orchestration package, memory package, durable engine, or `RetrieveExecutor`.

---

## Architecture

The Increment 1 direction stands. This pass only hardened semantics, tests, and documentation.

---

## Changes

| Area | Change |
|---|---|
| `concurrency.Pool` | Unstarted tasks after `ctx` done are `ctx.Err()`, never nil success |
| `workflow.Parallel` | Unique part keys; `ErrNeverStarted`; nested wait → `ErrNestedWait`; deterministic `PartialError` strings |
| `Envelope.Decision` | Typed HITL payload; `Metadata` is not a control protocol |
| `Resume` | `ClaimWaiting` so double/completed/cancelled resume fails; terminal status saved |
| Nested HITL | Parallel / Branch / Supervisor reject waits (`ErrNestedWait`), not `IsWait` |
| `Branch` cond | `(Envelope) (string, error)` |
| Errors | Sentinel vars + wrapping; `IsWait` excludes nested waits |
| `InputWait` | Requires `Decision.Input` (does not reuse prior `Envelope.Input`) |
| Docs | pause/resume vs durable; top-level HITL only |

Unrelated dirty files (`*publish.go`, `ai/provider.go`, other `provider.go`) were not modified.

---

## Parallel semantics

| Outcome | Meaning |
|---|---|
| Success | Every part **started** and returned nil; join runs |
| Error | `*PartialError`; `Failed` has every failed key; `Parts` keeps sibling envelopes; parent Envelope is not joined |
| Multi-error | All failures recorded; `Error()` lists keys sorted |
| Cancel | Running parts see `ctx`; parts that never started are `ErrNeverStarted` wrapping `context.Canceled`; not reported as success |
| Timeout | `DeadlineExceeded` and/or `ErrNeverStarted`; `StatusFailed` at graph level (deadline is not `StatusCancelled`) |

`err == nil` from Pool means started-and-succeeded.

---

## HITL semantics

- **Supported:** top-level Graph node (`Approval` / `InputWait`).
- **Mechanism:** in-process pause/resume (`WaitError` + `Resume` + `Envelope.Decision`).
- **Not durable:** process crash/restart loses `MemoryCheckpointer` state.
- **Nested HITL:** not supported; API returns `ErrNestedWait` (failure). Continuation position is not claimed as waiting.
- **Paused interval:** does not block a goroutine; `WithTimeout` on `Run` does not cover human wall-clock time.
- **Resume:** must be waiting; `ClaimWaiting` prevents a second Resume; cancelled Resume ctx does not claim; invalid `InputWait` payload fails the execution.

---

## Envelope

- No `any` / `map[string]any`.
- HITL uses `*Decision` (`Approved`, `Input`, `Comment`).
- `Route` is supervisor hop selection.
- `Metadata` is auxiliary only.

---

## State machine (tested)

| Component | Success | Error | Cancel | Timeout | Invalid |
|---|---|---|---|---|---|
| Sequential | ✓ | ✓ | ✓ | ✓ | ✓ empty / missing executor |
| Parallel | ✓ | ✓ multi-fail | ✓ | ✓ | ✓ duplicate names; nested wait |
| Branch | ✓ true/false | ✓ selected fail; cond error | ✓ | ✓ | ✓ unknown route; nil cond |
| Supervisor | ✓ | ✓ worker / router | ✓ | ✓ | ✓ unknown route; max rounds; nested wait |
| HITL | ✓ approve + complete | ✓ reject | N/A on pause (non-blocking); ✓ cancelled Resume does not claim | N/A on pause interval; ✓ Resume continuation timeout | ✓ unknown id; not waiting; empty id; missing checkpointer; InputWait without Input |
| Resume | ✓ | ✓ invalid payload | ✓ cancelled ctx | ✓ continuation timeout | ✓ double resume; completed; cancelled run |

**N/A documented:** HITL timeout/cancel *during the paused interval* — wait returns immediately; there is no parked goroutine to cancel or time out. Resume is the continuation API.

---

## Timeout hierarchy

Tightest context deadline wins.

- `Node.Timeout` ⊂ workflow ctx.
- Step deadline → `StatusFailed` + `DeadlineExceeded`.
- Parent `WithCancel` → `StatusCancelled`.
- Parent `WithTimeout` → `DeadlineExceeded` / `StatusFailed`.
- Completion before deadline → `StatusCompleted`.

---

## Verification

| Command | Result |
|---|---|
| `go test ./...` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test -race ./...` | **ENVIRONMENT BLOCKED** — `cgo: C compiler "gcc" not found` |

Supplementary coverage (not the bar): workflow 83.3%, agent 71.5%, concurrency 54.8%.

`go list -deps github.com/zatrano/packages/workflow` among this module: `concurrency`, `workflow` only.

---

## Known limitations (out of scope, not silently ignored)

- No crash-safe durable workflow.
- Nested HITL not resumable (explicitly rejected).
- `Checkpointer` is in-process; SQL/Redis backends not shipped.
- Agent `Authorizer == nil` still allows all tools (documented default).
- `agent.Chain` / `agent.Graph` remain a second agent-only walker.
- No workflow queue runner.
- No MCP / A2A.
- Hop execution after Resume is at-least-once if a later hop crashes mid-flight (no hop idempotency keys).
- `go test -race` requires cgo/gcc.

---

## Increment 2 gate

> Increment 2 remains blocked until Increment 1 closeout is accepted.
