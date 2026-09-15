# ZATRANO AI — Rebuild Plan

**Date:** 2026-09-11  
**Prerequisite:** [AI-ARCHITECTURE-DECISION.md](AI-ARCHITECTURE-DECISION.md) (HYBRID REBUILD)

This is a migration plan, not a license to keep a wrong API forever.

---

## Strategy

**HYBRID REBUILD:** keep working `ai` drivers, `rag` pipeline, and `agent.Run`; introduce `workflow` as the missing boundary; harden agent cancellation/authz; wrap or internally route Chain/Graph toward workflow.

A **clean-room rewrite** of OpenAI/Anthropic/Gemini/RAG stores would destroy test equity for no boundary gain. A **pure evolve** (more methods on `Agent`) would cement Graph-as-workflow.

Breaking changes are allowed when they remove a false boundary. Prefer additive + wrappers for `agent.Chain` / `agent.Graph` so existing tests and apps keep compiling.

---

## Component ledger

### KEEP

| Current | Target | Reason | Migration risk | API | Tests | Deps |
|---|---|---|---|---|---|---|
| `ai/*` drivers, Client, Manager, profiles, retry, stream, meter | `ai/` | Correct runtime | Low | Compatible | Keep | Unchanged |
| `rag.Pipeline`, stores, rerank | `rag/` | Correct library | Low | Compatible | Keep | Unchanged |
| `agent.Agent`, Registry, BufferMemory, Catalog, Runner | `agent/` | Correct single-agent loop | Medium (hardening) | Additive | Extend | +workflow (AsExecutor) |
| `queue` | `queue/` | Correct infra | None | None | None | None |

### MOVE

| Current | Target | Reason | Risk | API | Tests | Deps |
|---|---|---|---|---|---|---|
| Graph semantics (executor hops, route, max hops) | `workflow.Graph` | Nodes must not be `*Agent`-only | Medium | New package; `agent.Graph` remains sugar | Port graph tests conceptually | workflow → concurrency |
| Chain semantics (sequential hops) | `workflow.Sequential` | Same | Low | Wrapper | Keep `agent.Chain` tests | agent → workflow optional |

### RENAME

None required. Do not rename `ai.Client` or `rag.Pipeline`.

### SPLIT

| Current | Target | Reason | Risk |
|---|---|---|---|
| `agent` as “loop + fake workflow” | `agent` + `workflow` | Cohesion | Medium (new package) |

Do **not** split `rag` ingest/retrieve.

### MERGE

None. Do not merge rag into ai. Do not merge agent into ai.

### REWRITE

| Current | Target | Reason | Risk | API | Tests |
|---|---|---|---|---|---|
| `Agent.Run` loop control | Same function, add ctx/timeout checks | Cancellation principle | Low | Additive `Timeout` | New cancel/timeout tests |
| Tool dispatch | Authorizer hook | Security | Low | Additive; nil = old behavior | Deny tests |

### DEPRECATE

| Current | Target | Reason | Risk |
|---|---|---|---|
| Using `agent.Graph` / `Chain` as a generic process engine | Prefer `workflow` | Wrong home | Docs + comments; keep types |

### DELETE

| Item | Reason |
|---|---|
| Orchestration package | Duplicates workflow |
| Memory package | Junk drawer |
| MCP/A2A implementations in this increment | Wrong time |
| Dynamic swarm runtime | Unjustified |
| `func Apply` / acquire changes | Frozen; unrelated |

Do **not** delete working provider code, RAG stores, or `agent.Chain`/`Graph` types in the first increment (compatibility wrappers).

---

## Increments

### Increment 1 — Workflow boundary (this rebuild)

1. Add `workflow` library: Envelope, Executor, Graph, Sequential, Parallel, Branch, Supervisor, Wait/Resume, MemoryCheckpointer, tests (cancel, timeout, parallel partial fail, wait/resume).
2. Add `agent.AsExecutor`, `Authorizer`, loop `ctx.Err()` + `Timeout`.
3. Optional: `RAGRetrieve` rerank flag.
4. Docs: package READMEs, CHANGELOG, architecture set.
5. Import-direction tests.

### Increment 2 — (not blocking architecture)

- `workflow.Runner` on `queue`
- Agent streaming loop
- Wire `ai.Healthy` into `health` from application Boot examples
- Durable SQL checkpointer
- MCP adapter module

### Increment 3

- A2A executor adapter
- Metadata filters in RAG
- Compensation helpers

---

## Clean rebuild vs incremental

| Approach | Verdict |
|---|---|
| Delete `ai`/`rag`/`agent` and rewrite | **Inferior** — drivers and pipeline are the valuable part |
| Only add features to `agent.Graph` | **Inferior** — encodes “everything is an agent” |
| New `workflow` + harden agent + keep drivers | **Superior** |

Backwards compatibility is **not** preserved if it keeps Graph as the only composer of HTTP/DB/human steps. Those steps never existed on Graph; adding `workflow` is the compatible *and* correct path.

---

## Test impact

Must add (increment 1):

- workflow unit, cancel, timeout, failure isolation, wait/resume, supervisor bounded rounds
- agent cancel during loop, authorizer deny, AsExecutor envelope mapping
- architecture import tests

Existing `ai`, `rag`, `agent` tests must still pass.
