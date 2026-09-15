# ZATRANO AI — Increment 1 Deep Architectural Review

**Date:** 2026-09-12  
**Closeout:** see [AI-INCREMENT-1-CLOSEOUT.md](AI-INCREMENT-1-CLOSEOUT.md) (hardening implemented).  
**Original verdict below is the review that authorized closeout, not a live status board.**

---

This is not a rewrite request. It is a boundary-and-correctness audit.

---

## Executive verdict

```text
DIRECTION: KEEP
INCREMENT 1: NOT CLOSED
INCREMENT 2 FEATURES: BLOCKED
NEXT: INCREMENT 1 CLOSEOUT (hardening only)
```

Do **not** add MCP, A2A, swarm, planner, memory package, or more agent features.

Do **not** change this split:

```text
ai       → model runtime
rag      → knowledge / retrieval
agent    → autonomous execution loop
workflow → process execution
```

The five user-raised checks:

| Check | Result |
|---|---|
| 1. `workflow` import purity | **PASS** |
| 2. Supervisor genericity | **PASS** (generic `Executor` only) |
| 3. RAG adapter boundary | **PASS** for workflow; **NOTE** on `agent.RAGRetrieve` |
| 4. Envelope not `map[string]any` | **PASS** with a **WARN** (Metadata string bag) |
| 5. waiting ≠ durable | **FAIL as naming**; capability is pause/resume |

---

## 1. Workflow dependency purity

**Invariant (required):**

```text
workflow → Executor + Envelope + stdlib + concurrency
workflow ↛ agent, rag, ai, queue
```

**Evidence:**

- Production `.go` files under `workflow/` import only `context`, `fmt`, `strings`, `time`, `errors`, `maps`, `sync`, `sort`, `crypto/rand`, `encoding/hex`, and `github.com/zatrano/packages/toolkit/concurrency`.
- `go list -deps github.com/zatrano/packages/workflow` among this module: `concurrency`, `workflow`.
- Compile-time names: no `ai.`, `rag.`, `agent.`, `ChatRequest`, `Pipeline` in `workflow/` except the package comment.
- `agent.AsExecutor` lives in `agent/executor.go` and imports `workflow`. Direction is inverted correctly.

**Test gap:** `TestWorkflowDoesNotImportIntelligence` uses `go list` on `.Imports` only. It does not inspect source ASTs and does not fail if a new file is added but `go list` is skipped. Sufficient for CI; not a substitute for `go vet` + the deps listing.

**Semantic coupling in docs:** README says a workflow may contain “RAG operations”. That is **allowed application composition** (a `Named` func that closes over `rag.Pipeline`). It is not a package dependency. Keep the wording; do not add `RetrieveExecutor` to `workflow`.

**PASS.**

---

## 2. Executor contract

```go
type Executor interface {
    Name() string
    Execute(ctx context.Context, env Envelope) (Envelope, error)
}
```

**What is right:** tiny, typed, ctx-first, no AI types, no `any`.

**Gaps:**

| Gap | Severity |
|---|---|
| No timeout on the interface (only `Node.Timeout` on Graph) | Low — Parallel children have no per-child timeout |
| `WaitError` is an error that is not a failure | Medium — easy to mishandle |
| No execution identity on `Execute` | Medium — nested `Branch.Run` mints a new `ExecutionID` |
| Name uniqueness not enforced | Medium — `Parallel` map keys collide |

The contract itself should **stay small**. Do not grow it into a framework object. Fix the call sites (Parallel keys, nested Run options).

---

## 3. Envelope design

Current fields: `Task`, `Goal`, `Input`, `Previous`, `Output`, `Required`, `Route`, `Artifacts`, `Constraints`, `Permissions`, `Metadata`.

**PASS:** there is no `map[string]any`. Artifacts/Metadata are `map[string]string`. Clone copies maps/slices.

**WARN — junk-drawer seed:** human `Decision` is flattened into `Metadata["decision"]` / `Metadata["comment"]` (`applyDecision` in `graph.go`). That is how Envelope becomes a bag: one magic key at a time.

**WARN — identity lives outside Envelope:** `ExecutionID` / `WorkflowID` / `StepID` are on `Result` and `Checkpoint` only. That is acceptable if we keep Envelope as *handoff payload*, not *execution record*. Do not copy identity into Metadata.

**Do not add** in closeout: extra maps, `any`, conversation history, raw tool transcripts.

**Do in closeout:** stop using Metadata for approval. A typed field or a dedicated wait-decision path (`Decision` already exists) is enough.

---

## 4. Supervisor genericity

`workflow.Supervisor` is `router Executor` + `map[string]Executor` + `Envelope.Route`. No `Agent`, no `Client`, no RAG.

Loop:

```text
router.Execute → Route empty? stop
               → lookup worker by Route
               → worker.Execute
               → repeat until maxRounds
```

**PASS** as a strategy, not as “orchestration.”

**Gaps (correctness, not genericity):**

- No fallback when the selected worker fails (not implemented; tests do not cover it). Do **not** add LLM fallback. If needed later: explicit `Route` after failure, still generic.
- HITL inside a worker: `WaitError` bubbles, but Graph checkpoints the **supervisor node**. `Resume` re-runs the whole supervisor from round 0. Inner wait position is lost.
- Supervisor inner hops are invisible to `workflow.Observer` (one Graph hop).

---

## 5. RAG adapter boundary

**Workflow:** does not import rag. There is no `RetrieveExecutor`. **Correct. Do not add one to workflow.**

**Agent:** `Retriever` is a string port (`Retrieve(ctx, query) (string, error)`). `RAGRetrieve` is an adapter that **embeds** `*rag.Pipeline` and `rag.Reranker`.

```text
agent.Retriever     ← clean port
agent.RAGRetrieve   ← convenience adapter (imports rag)
agent.FuncRetriever ← injection without rag types
```

This is **not** a hidden workflow leak. It is the intended `Agent → Retriever` design.

**NOTE:** the name `RAGRetrieve` exports RAG into the agent public surface. That is acceptable cohesion for a convenience type. Purity option (later, not now): move the struct to `rag` as `func (p *Pipeline) Retriever(...)` returning a func — but `rag → agent` would be the wrong arrow. Keep the adapter in `agent`.

Applications that want retrieval as a **workflow step** should write `workflow.Named("retrieve", ...)` in **app code**, closing over `rag.Pipeline`. That keeps both packages generic.

**PASS** for package boundaries. No change to Increment 1 direction.

---

## 6. Cancellation propagation

| Path | Behavior | Evidence |
|---|---|---|
| Graph hop | `ctx.Err()` before Execute | `graph.go` walk |
| Step | child `WithTimeout` ctx | `Node.Timeout` |
| Executor | must honor ctx; `Named` checks once at entry | `executor.go` — **not** during `fn` |
| Parallel | same ctx to `concurrency.Pool` | see §8 / §13 |
| Supervisor | `ctx.Err()` each round; passed through | `supervisor.go` |
| Agent | `ctx.Err()` each step; Chat/tools get ctx | `agent.go` |
| Wait | non-blocking return; nothing to cancel in-process | `wait.go` |

**WARN:** if an executor ignores `ctx`, step/workflow timeout does not preempt it. Document as contract: `Execute` must return on `ctx.Done()`.

**WARN:** agent checks cancel between steps, not between tool calls in the same step.

**Missing test:** cancel during Parallel; cancel during Supervisor; abandon waiting execution.

---

## 7. Timeout hierarchy

Implemented and distinct:

```text
WithTimeout(Run)     → workflow deadline → StatusFailed on deadline
Node.Timeout         → step deadline     → StatusFailed (parent ctx still live)
Agent.Timeout        → agent.Run wrap
ai.Defaults.Timeout  → model HTTP call
```

`StatusCancelled` is only `context.Canceled`, not deadline. That distinction is correct.

**WARN:** Parallel parts have no per-child timeout except whatever the child executor implements.

**WARN:** `Branch` inner `g.Run(ctx, env)` inherits the step ctx (good) but does **not** forward `WithCheckpointer` / `WithExecutionID` (bad for wait/resume).

---

## 8. Parallel failure semantics

**Intended:** siblings are isolated; `PartialError.Parts` keeps successful envelopes; parent `Envelope` is **not** joined on failure (`return env, &PartialError{...}`).

**PASS** for “failed child does not erase sibling results” — **if** the caller unwraps `*PartialError`. `Result.Envelope` on the Graph is still the **pre-parallel parent**. Easy to misuse.

**FAIL / defect — cancellation + Pool:**

`concurrency.Pool` leaves `errs[i] == nil` for tasks never started after `ctx.Done()`. `Parallel` treats nil error as success with a **zero Envelope**. Cancelled fan-out can look like partial success.

**FAIL / defect — duplicate `Name()`:** last writer wins in `parts`/`failed` maps.

**WAIT inside Parallel:** `WaitError` is classified as a child failure, not Graph `StatusWaiting`. Human wait in a fan-out is not supported. Document as out of scope rather than silently mis-handling.

**Tests present:** success; one child fails (siblings kept).  
**Tests missing:** multiple failures; cancel; timeout; duplicate names.

---

## 9. Branch semantics

`Branch` is a `Sequential` of one `Named` func that calls `paths[key].Run(ctx, env)` without options.

| Case | Behavior | Tested |
|---|---|---|
| Matching key | inner graph runs | yes (`Input` contains `go`) |
| Other key with path | `otherwise` unused | **no** (`no` path untested) |
| Unknown key, `otherwise` set | used | **no** |
| Unknown key, no otherwise | error | **no** |
| `cond == nil` | error | **no** |

**Structural issue:** inner `Run` mints a new `ExecutionID` and has no checkpointer. A `WaitError` from the inner graph returns to the outer hop. Outer checkpoint stores the **branch executor** step. `Resume` re-runs `cond` + inner graph from **Start**, not from the inner wait node.

HITL inside `Branch` is not a supported composition until options are forwarded or Branch inlines nodes.

---

## 10. Human Resume semantics

**What exists:** `Approval` / `InputWait` return `WaitError`; Graph sets `StatusWaiting`; `Resume` requires `WithCheckpointer`; re-executes the wait node with `Metadata["decision"]`.

This is **pause/resume**, not durability.

| Case | Status |
|---|---|
| Wait then approve | Tested |
| Wait then reject | Tested |
| Resume without checkpointer | Errors (untested) |
| Resume unknown id | Errors (untested) |
| Resume when not waiting | Errors (untested) |
| Cancel while waiting | **No API** — wait is not a blocking call |
| Double resume after complete | Untested |
| Resume traces | **New Result.Traces**; prior hops discarded |
| Wait inside Parallel/Branch/Supervisor | Broken or undefined (see above) |

`MemoryCheckpointer` is process-local. Process crash loses waiting executions. README “Optional hop/wait persistence” is accurate; “durable workflow” would not be.

**No `Cancel(executionID)`.** Abandonment is “drop the id.” Fine if documented.

---

## 11. Idempotency

**Not designed.**

- `Resume` re-executes the wait node (ok if wait is pure) then **re-enters the next hop**.
- Checkpoint is written **after** a successful hop, not before. Crash mid-hop + later Resume from previous wait can **run the next executor twice**.
- Same `ExecutionID` + `Save` **overwrites** with no occupancy lock. Two concurrent `Run`s with one id corrupt each other.
- HTTP `packages/idempotency` is the wrong plane (response replay). Do not reuse it as hop identity.

Closeout: document “at-least-once after resume.” Do not claim exactly-once. A hop-level idempotency key is Increment 2+ **if** durable backends appear.

---

## 12. Execution identity

Present: `Result.ExecutionID`, `Checkpoint.ExecutionID`, `WorkflowID` (`Graph.ID`), `StepID` (node name).

Generated with `crypto/rand` when omitted. `Resume` requires the caller to pass the same id **and** the same `Checkpointer`.

**Gaps:** identity not on Envelope; nested `Branch.Run` allocates a second id; no running/waiting occupancy.

---

## 13. Concurrency / race

- `MemoryCheckpointer` is mutex-protected.
- Parallel writes `results[i]` by unique index (safe) then copies to a map (post-join, safe).
- `Envelope.clone()` before child Execute (good).
- `concurrency.Pool` cancel/unstarted-nil-error (see §8).
- `go test -race` **still not run** (no gcc). Open verification item. Make it required in CI, not on this workstation.

Graph `Run` is not advertised as concurrent on the same `*Result`. Agent `Memory` is not concurrent-safe. Fine if documented.

---

## 14. Observability

`Observer.OnHop` fires per **Graph** hop, including wait and failure.

**Blind spots:** Parallel children, Supervisor rounds, Branch inner graph hops, AI token usage.

Do **not** invent a second telemetry system. Closeout: document that nested strategies are opaque unless the application wraps executors.

---

## 15. Error taxonomy

Typed: `WaitError`, `PartialError`, `context.Canceled`, `context.DeadlineExceeded`.

Stringly: unknown node, max hops, unknown worker, approval rejected, empty executor.

**Dual channel:** `StatusWaiting` + `err != nil` (`WaitError`). Callers that do `if err != nil { fail }` treat wait as failure unless they `IsWait`. This must stay in README examples.

**PartialError vs Result.Envelope:** sibling outputs only on the error value.

---

## 16. API ergonomics

**Good:** `Named`, `Sequential`, `Parallel` as Executor, `Supervisor` as Executor, `AsExecutor` on agent, `Route` not `Output` for supervisor (fixed after the `"t done"` worker-name bug).

**Friction:**

- Magic Metadata keys for HITL.
- `Branch` extra wrapper hop in traces (`id-branch`).
- Wait as error.
- Agent `Authorizer == nil` still allows every tool (documented unsafe default).

Do not add a builder DSL in closeout.

---

## 17. Test state-machine coverage

Coverage numbers (workflow 77.9%, agent 71.5%) are **not** the bar.

| Required case | Present? |
|---|---|
| Parallel success | YES |
| Parallel one branch fails | YES |
| Parallel multiple branches fail | NO |
| Parallel cancellation | NO |
| Parallel timeout | NO |
| Parallel partial completion (siblings kept) | YES (one-fail test) |
| Supervisor selection | YES |
| Supervisor selected executor fails | NO |
| Supervisor fallback | NO (no feature) |
| Supervisor cancellation | NO |
| Human waiting | YES |
| Human resume | YES |
| Cancel while waiting | NO |
| Invalid resume | PARTIAL (reject only) |
| Branch true | YES |
| Branch false | NO |
| Branch invalid condition | NO |
| Step timeout | YES |
| Workflow timeout | YES |
| Child timeout (parallel) | NO |
| Import purity | YES (`go list`) |
| Envelope not transcript | YES |

---

## Pause/resume vs durability (naming)

| Term | Increment 1 meaning |
|---|---|
| `StatusWaiting` + `Resume` | Pause/resume in this process (or whatever `Checkpointer` implementation is passed) |
| `MemoryCheckpointer` | **Not durable** across process death |
| Durable workflow | **Not shipped** — interface exists (`Checkpointer`) |

README and FINAL-AUDIT must not say “durable workflow” for MemoryCheckpointer. The architecture decision already said “YES as interface; NO as mandatory persistence.” Keep that sentence; tighten README.

---

## What Increment 2 must **not** be

Blocked until closeout:

- MCP / A2A packages
- Swarm / group chat / planner
- `memory/` package
- `orchestration/` package
- Queue rewrite
- SQL checkpointer as a **feature flag** before Parallel cancel and HITL composition are honest

---

## Increment 1 closeout (authorized hardening only)

Ordered. No new conceptual packages.

1. **Docs honesty:** pause/resume vs durable; HITL only as a **top-level Graph node** (not inside Parallel / Branch inner graphs / Supervisor workers) until those compositions checkpoint inner position.
2. **State-machine tests** for the table in §17 (especially Parallel cancel/timeout/multi-fail, Branch false/invalid, Supervisor worker fail/cancel, invalid Resume).
3. **Parallel:** unique child keys; after Pool, nil `errs[i]` + cancelled ctx → treat as cancelled, not success.
4. **HITL:** stop stuffing `Decision` into `Metadata` strings (typed path). Restore or document empty Resume traces.
5. **CI:** `go test -race` required where gcc/cgo exists.
6. **Do not** add `RetrieveExecutor` to `workflow`.
7. **Do not** make Supervisor know about agents.

After closeout tests are green, Increment 1 can be marked closed. Then Increment 2 may consider: unify `agent.Chain`/`Graph` onto workflow internals, `workflow.Runner` on queue, durable `Checkpointer` backend — still no MCP/A2A.

---

## Decision for the next coding pass

```text
AUTHORIZED: Increment 1 closeout (tests + Parallel cancel correctness + docs/HITL typing)
FORBIDDEN:  Increment 2 product surface
```
