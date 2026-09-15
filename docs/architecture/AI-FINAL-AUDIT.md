# ZATRANO AI — Final Audit

**Date:** 2026-09-11  
**Decision executed:** HYBRID REBUILD (increment 1)  
**Module:** `github.com/zatrano/packages` (still v1.7.x line; changes are Unreleased)

This is a production-adoption review of the tree **after** increment 1, not a restatement of the pre-change reconstruction.

---

## 1. What changed

- Architecture investigation and decision set under `docs/architecture/`.
- New library `workflow`: Envelope, Executor, Graph, Sequential, Parallel, Branch, Supervisor, Approval/InputWait, Checkpointer, cancellation/timeouts.
- `agent.AsExecutor` — agent as a workflow hop without dumping transcripts.
- `agent.Authorizer`, `Agent.Timeout`, per-step `ctx.Err()` in `Run`.
- `RAGRetrieve.Rerank` uses `Pipeline.QueryWith`.
- Import-boundary tests: `workflow` ↛ ai/rag/agent; `ai` ↛ rag/agent/workflow.
- README / CHANGELOG updated to describe the actual split.

## 2. What was deleted

- Nothing of the working AI providers, RAG stores, or `agent.Chain`/`Graph` types.
- Rejected as packages (never created): `orchestration/`, `memory/`, `mcp/`, `a2a/`.

## 3. What was moved

- **Semantic home of generic composition** is now `workflow`, not `agent`.
- `agent.Graph` / `Chain` remain as **agent-only sugar** (same files). They were not deleted and were not internally rewritten onto `workflow` in increment 1 (compatibility). Comments and README point new process code at `workflow`.

## 4. What was rewritten

- `Agent.Run` control flow (cancel/timeout/authorizer dispatch) — same function, stricter.
- `RAGRetrieve.Retrieve` — optional rerank path.

Not rewritten: OpenAI/Anthropic/Gemini drivers, `ai.Client` retry/fallback, RAG stores.

## 5. Why each major decision

| Decision | Why |
|---|---|
| HYBRID REBUILD | Drivers/pipeline are sound; Graph-as-workflow is not. |
| `ai` stays a service | Named provider registry is in-process runtime state. |
| `rag` / `agent` stay libraries | Caller-owned values; no global agent. |
| `workflow` is first-class | Non-agent steps cannot live on `*Agent` nodes. |
| No `orchestration` package | Sequential/parallel **are** workflow; supervisor is a strategy. |
| No Memory package | Conversation ≠ knowledge ≠ execution state. |
| MCP/A2A later | Protocol adapters, not internal buses. |
| Queue reused | Already the job plane (`agent.Runner`). |
| Parallel uses `concurrency.Pool` | Do not fork a worker pool. |

## 6. Final package tree (intelligence)

```
ai/          service   provider runtime
rag/         library   retrieve pipeline
agent/       library   single-agent loop + queue adapter + agent-only Chain/Graph
workflow/    library   generic executor graphs
queue/       service   jobs (consumed, not duplicated)
concurrency/ library   used by workflow.Parallel
```

## 7. Final dependency graph

```
workflow → concurrency
ai       → framework (not rag/agent/workflow)
rag      → ai
agent    → ai, rag, queue, workflow
queue    → (not intelligence)
```

Forbidden edges are tested.

## 8. Public API summary

**workflow:** `Envelope`, `Executor`, `Named`, `Graph.Run`/`Resume`, `Sequential`, `Parallel`, `ConcatJoin`, `Branch`, `Supervisor`, `Approval`, `InputWait`, `WaitError`, `PartialError`, `MemoryCheckpointer`, `WithTimeout`/`WithCheckpointer`/`WithExecutionID`/`WithObserver`.

**agent (additive):** `Authorizer`, `FuncAuthorizer`, `Timeout`, `AsExecutor`, `RAGRetrieve.Rerank`.

**ai / rag:** unchanged public surface.

## 9. Workflow capabilities (implemented)

| Capability | Status |
|---|---|
| Step / Executor / Envelope | Yes |
| Sequential | Yes |
| Parallel fan-out/fan-in | Yes (`PartialError` keeps siblings) |
| Branch | Yes |
| Supervisor strategy | Yes (`Envelope.Route`) |
| Cancellation | Yes (`ctx`) |
| Workflow vs step timeout | Yes (`WithTimeout` vs `Node.Timeout`) |
| Human approval pause/resume | Yes |
| In-memory checkpoint | Yes |
| Durable SQL checkpoint | No (interface only) |
| Queue-backed workflow runner | No (agent queue remains) |
| Compensation/sagas | No (explicit next-hop only) |

## 10. Multi-agent capabilities

| Pattern | Status |
|---|---|
| Sequential agents | `workflow.Sequential` + `AsExecutor`, or `agent.Chain` |
| Concurrent agents | `workflow.Parallel` + `AsExecutor` |
| Handoff | `Envelope` (task/goal/artifacts/constraints/permissions) |
| Supervisor | `workflow.Supervisor` |
| Group chat / swarm | Not implemented (by design) |
| Dynamic LLM orchestration | Not implemented (by design) |

## 11. RAG capabilities

Unchanged core pipeline. Agent retrieval can rerank. Still no metadata filter, hybrid search, or Redis store.

## 12. AI runtime capabilities

Unchanged: providers, profiles, retry/fallback, stream, tools schema, vision/image/audio/embed, usage meter. Still not wired to `httpclient`/`health`/`observability` (parallel planes, documented).

## 13. Security model

- **New:** `Authorizer` runs before the tool handler. Denied → `ToolDenied`, handler does not run.
- **Default:** nil authorizer = allow all registered tools (compatibility). Production must set one.
- Retrieved text is still prompt data, not a grant.
- `web_fetch` HTTPS-only and `file_search` root sandbox remain.
- No tenant isolation inside the loop; that stays application-level.
- Secrets still belong in provider config, not envelopes.

## 14. Observability model

- AI: existing `Observer` / `UsageMeter`.
- Workflow: `workflow.Observer` hop traces.
- Not a new telemetry product. No automatic bridge to `packages/observability`.

## 15. Durability model

- Default `workflow.Run` is ephemeral.
- `Checkpointer` + `ExecutionID` support **in-process pause/resume**, not crash-safe durability.
- `MemoryCheckpointer` is process-local; SQL/Redis checkpointer not shipped.
- Queue `ResultStore` still a **final snapshot** of `agent.run`, not a hop log.

## 16. MCP decision

**Not implemented.** Native tools stay `Handler` + `ai.Tool`. MCP is a future adapter into `Registry`.

## 17. A2A decision

**Not implemented.** Internal agents are in-process executors. A2A would be a remote `Executor` later.

## 18. Compatibility impact

| API | Impact |
|---|---|
| `ai.*` | Compatible |
| `rag.*` | Compatible |
| `agent.Run` | Compatible; nil Authorizer preserves old tool policy |
| `agent.Chain` / `Graph` | Compatible (still agent-only) |
| New `workflow` | Additive |
| Breaking | None required for increment 1 |

Migration: new mixed processes should import `workflow` and `agent.AsExecutor` instead of stuffing HTTP/SQL/humans into `agent.Graph`.

## 19. Remaining limitations

- `agent.Chain`/`Graph` are a second, agent-only walker (not deleted).
- No workflow queue runner.
- No durable checkpointer backend.
- No agent streaming loop.
- No metadata filters in RAG.
- `go test -race` **could not run here**: `CGO_ENABLED=1` requires gcc, which is not on PATH. `go test ./...`, `go vet ./...`, and `go build ./...` passed. Cover: workflow 77.9%, agent 71.5%.
- AI still uses `net/http` rather than `httpclient` (intentional this increment).
- Nil `Authorizer` remains an unsafe default.

## 20. Future roadmap

1. Unify `agent.Chain`/`Graph` internals onto `workflow` or deprecate them.
2. `workflow.Runner` on `queue` (mirror of `agent.Runner`).
3. SQL/file `Checkpointer`.
4. Optional Boot adapter: register `ai.Healthy` into `packages/health`.
5. Agent streaming loop.
6. RAG metadata filters.
7. MCP adapter module (tools only).
8. A2A executor adapter (remote agents only).

---

## Production-adoption checklist

| Question | Answer |
|---|---|
| Package boundaries clean? | Yes for increment 1 (`workflow` has no intelligence imports). |
| Dependencies directional? | Yes; tests enforce ai/workflow isolation. |
| Cycles? | No. |
| Duplicated composition engines? | Partially: agent Chain/Graph still exist beside workflow. |
| Giant abstractions? | Avoided (no UniversalOrchestrator). |
| Agent defined? | Single-agent `Run` loop. |
| Tools isolated? | Authorizer + registry; default still permissive. |
| Memory vs knowledge? | `Memory` vs `Retriever`/`rag`. |
| Workflow non-agent steps? | Yes. |
| Cancellation? | Workflow + agent loop + ctx to chat/tools. |
| Durability? | Interface + memory; not production durable yet. |
| Supervisor just a strategy? | Yes. |
| Dynamic orchestration optional? | Absent (not defaulted). |
| RAG independent? | Yes; agent uses Retriever port. |
| MCP/A2A required internally? | No. |
