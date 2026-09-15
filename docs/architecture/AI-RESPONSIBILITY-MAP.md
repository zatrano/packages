# ZATRANO AI — Responsibility Map

**Companion to:** [AI-CURRENT-ARCHITECTURE.md](AI-CURRENT-ARCHITECTURE.md)  
**Date:** 2026-09-11  
**Decisions are target-oriented.** They do not modify code by themselves.

Legend: `KEEP` `MOVE` `SPLIT` `MERGE` `REPLACE` `DELETE` `UNKNOWN`

---

## How to read a row

Each component is classified by **actual** responsibility found in source, not by folder name.

---

## AI package

### `ai.ServiceProvider` / `From` / `Manager`

| Field | Content |
|---|---|
| Current responsibility | Addon registration; in-process driver/profile registry |
| Actual responsibility | Application-scoped AI runtime (named providers + defaults) |
| Target responsibility | Same: the AI **service** boundary |
| Reason | Apps need config, `From(app)`, and a single registry. That is runtime state, not an agent. |
| Dependencies | framework contracts/config/routing |
| Lifecycle | Register/Boot; no Start/Stop |
| Public API impact | Keep `From`, `New`, `BootConfig`, `Using`, `Profile` |
| Decision | **KEEP** |

### `ai.Client`

| Field | Content |
|---|---|
| Current responsibility | Scoped chat/embed/stream with retry + fallback |
| Actual responsibility | Resilient invocation facade over `Driver` |
| Target responsibility | Same |
| Reason | Cohesive. Do not fold into Agent. |
| Dependencies | `Manager` |
| Lifecycle | Stateless handle; Manager is the owner |
| Public API impact | Keep |
| Decision | **KEEP** |

### `ai.Driver` and provider implementations

| Field | Content |
|---|---|
| Current responsibility | Provider HTTP adapters |
| Actual responsibility | Model-provider I/O |
| Target responsibility | Same; remain inside `ai` (no `ai/openai` split required at this size) |
| Reason | Flat package is busy but splitting SDKs now is cosmetic. |
| Dependencies | `net/http` |
| Lifecycle | None |
| Public API impact | Keep constructors |
| Decision | **KEEP** |

### `ai` retry / fallback / profiles / live routing

| Field | Content |
|---|---|
| Current responsibility | Per-provider retry; ordered fallback; health filter |
| Actual responsibility | Invocation resilience for **model calls** |
| Target responsibility | Stay in `ai`. Optionally **adapt** to `circuit` / `health` later; do not replace working fallback with a breaker in v1 of the rebuild. |
| Reason | Model-call resilience is an AI concern. App `/health` is a different plane. |
| Dependencies | None of the sibling infra packages today |
| Lifecycle | `WatchHealth` is caller-owned |
| Public API impact | Keep |
| Decision | **KEEP** (wire, don’t clone, on a later increment) |

### `ai.Observer` / `UsageMeter` / `PriceTable`

| Field | Content |
|---|---|
| Current responsibility | Token/cost hooks |
| Actual responsibility | AI-domain telemetry |
| Target responsibility | Keep types; document how to forward into `observability` if desired |
| Reason | HTTP Prometheus metrics ≠ token economics. Merging them into one type would be a category error. |
| Dependencies | None |
| Lifecycle | Optional on Manager |
| Public API impact | Keep |
| Decision | **KEEP** |

### `ai` tools / vision / structured / image / audio types

| Field | Content |
|---|---|
| Current responsibility | Request/response shapes and some OpenAI-only drivers |
| Actual responsibility | Model I/O vocabulary |
| Target responsibility | Types stay in `ai`. Execution of tools stays in `agent`. |
| Reason | Correct split already. |
| Dependencies | — |
| Lifecycle | — |
| Public API impact | Keep |
| Decision | **KEEP** |

### `ai` demo route

| Field | Content |
|---|---|
| Current responsibility | `POST /demo/ai/chat` on Boot |
| Actual responsibility | Sample HTTP, not runtime |
| Target responsibility | Remain as optional demo (consistent with other addons) |
| Reason | Harmless; not an architecture axis |
| Decision | **KEEP** |

### Concerns **not** in `ai` today that must not be added

Agent loop, RAG pipeline, workflow graph, queue workers, MCP/A2A servers. Adding them would recreate a UniversalAIManager.

---

## RAG package

### `rag.Pipeline` / `Chunker` / `Embedder` / `VectorStore`

| Field | Content |
|---|---|
| Current responsibility | Index + query |
| Actual responsibility | Retrieval library |
| Target responsibility | Same library. Do **not** split ingest vs retrieve packages. |
| Reason | Both sides share Chunk/Embed/Store. Two packages would share almost every type. Split only if ingest becomes a product (connectors, schedulers). |
| Dependencies | `ai` for `FromAI` / LLM rerank |
| Lifecycle | Caller-owned values |
| Public API impact | Keep; additive filters/rerank wiring later |
| Decision | **KEEP** |

### Vector store adapters

| Field | Content |
|---|---|
| Current responsibility | Memory/JSON/SQL/PGVector |
| Actual responsibility | Persistence adapters |
| Target responsibility | Remain adapters behind `VectorStore` |
| Reason | Correct. Do not promote PGVector to “the RAG service.” |
| Decision | **KEEP** |

### `rag` rerankers

| Field | Content |
|---|---|
| Current responsibility | Optional post-retrieval scoring |
| Actual responsibility | Composable rerank stage |
| Target responsibility | Same; agent retrieval should be able to use `QueryWith` |
| Reason | Stage exists but the agent adapter skips it |
| Decision | **KEEP** (agent adapter should consume it) |

### Hypothetical `ingest` package

| Field | Content |
|---|---|
| Current responsibility | N/A (ingestion is `Pipeline.Index`) |
| Actual responsibility | Would own connectors, MIME parsers, schedules |
| Target responsibility | **Not now** |
| Reason | No connectors exist. Creating the package would be a vacant boundary. |
| Decision | **DELETE** as an idea (do not introduce) |

---

## Agent package

### `agent.Agent` / `Run` / `Chatter` / `Result`

| Field | Content |
|---|---|
| Current responsibility | Chat↔tool loop |
| Actual responsibility | **Single-agent execution** |
| Target responsibility | Same, with cancellation, timeouts, and tool authorization |
| Reason | This is the correct Agent definition. |
| Dependencies | `ai` |
| Lifecycle | Per-`Run` (ephemeral unless Memory is shared) |
| Public API impact | Additive fields (`Timeout`, `Authorize`); loop checks `ctx` |
| Decision | **KEEP** (harden) |

### `agent.Registry` / `Handler` / `ToolResult`

| Field | Content |
|---|---|
| Current responsibility | Tool defs + execution |
| Actual responsibility | In-process tool runtime |
| Target responsibility | Same, plus an authorization hook before execute |
| Reason | Tools are agent-owned, not AI-owned |
| Decision | **KEEP** (extend) |

### `agent.Memory` / `BufferMemory`

| Field | Content |
|---|---|
| Current responsibility | Conversation transcript |
| Actual responsibility | **Working/conversation memory only** |
| Target responsibility | Remain agent-scoped interface. Do not become RAG or workflow state. |
| Reason | Collapsing knowledge + execution state + chat history is the defect. The type is fine if we stop pretending it is all memory. |
| Decision | **KEEP** |

### Separate `memory` package

| Field | Content |
|---|---|
| Current responsibility | N/A |
| Actual responsibility | Would own conversation + persistent + working stores |
| Target responsibility | **Not a package.** Persistent stores implement `agent.Memory`. Retrieved knowledge stays `rag`. Execution state stays `workflow`. |
| Reason | A Memory package would become a junk drawer. |
| Decision | **DELETE** as an idea (do not introduce) |

### `agent.Retriever` / `RAGRetrieve`

| Field | Content |
|---|---|
| Current responsibility | String injection before the user turn |
| Actual responsibility | Context provider port |
| Target responsibility | Keep the port. Prefer `QueryWith` when a reranker is set. Agent must not import store internals. |
| Reason | `Agent → Retriever` is the clean dependency. |
| Decision | **KEEP** (small fix) |

### `agent.Catalog` / `Runner` / `PushRun` / `ResultStore`

| Field | Content |
|---|---|
| Current responsibility | Named agents + queue jobs + outcome snapshot |
| Actual responsibility | Async **transport** adapter |
| Target responsibility | Keep. Mirror later for workflow jobs. Do not turn ResultStore into a checkpoint log. |
| Reason | Queue reuse is correct. |
| Decision | **KEEP** |

### `agent.Chain`

| Field | Content |
|---|---|
| Current responsibility | Sequential `*Agent` steps |
| Actual responsibility | Multi-agent sequence, not a workflow |
| Target responsibility | Convenience wrapper over `workflow.Sequential` **or** remain as agent-only sugar that delegates to workflow |
| Reason | Sequential composition of arbitrary steps belongs in workflow. Agent-only sugar may remain if it calls workflow. |
| Dependencies | Today: none of workflow. Target: `workflow` |
| Public API impact | Prefer keep signatures; change internals |
| Decision | **MOVE** (semantics to workflow; API may wrap) |

### `agent.Graph`

| Field | Content |
|---|---|
| Current responsibility | Conditional hops between `*Agent` |
| Actual responsibility | Tiny agent state machine |
| Target responsibility | Generic executor graph lives in `workflow`. Agent graph becomes a wrapper or thin helper. |
| Reason | Nodes cannot be non-agents today. That is the architectural bug. |
| Public API impact | Existing `Graph` can remain as agent-typed sugar; new work uses `workflow.Graph` |
| Decision | **MOVE** (semantics to workflow) |

### `agent` built-in tools (`web_fetch`, `file_search`)

| Field | Content |
|---|---|
| Current responsibility | Example tools with local policy |
| Actual responsibility | Optional tools, not the tool model |
| Target responsibility | Keep |
| Decision | **KEEP** |

---

## Graph / Chain / Workflow / Orchestration

### Workflow (does not exist)

| Field | Content |
|---|---|
| Current responsibility | Missing |
| Actual responsibility | N/A |
| Target responsibility | First-class **library**: executors, edges, envelope, cancellation, optional wait/checkpoint |
| Reason | A process may contain functions, HTTP, SQL, RAG, agents, and humans. Forcing `*Agent` on every node is wrong. |
| Dependencies | Must **not** import `agent` or `ai` |
| Lifecycle | Per execution; optional checkpoint |
| Public API impact | New package |
| Decision | **KEEP** as a new boundary (introduce) |

### Orchestration (does not exist as a package)

| Field | Content |
|---|---|
| Current responsibility | Informal: Chain + Graph |
| Actual responsibility | People use “orchestration” to mean three different things |
| Target responsibility | **Not a package.** Multi-agent patterns are workflow graphs whose executors happen to be agents, plus a supervisor **strategy**. |
| Reason | Sequential/concurrent/fan-in are workflow primitives. Supervisor is one strategy. A third package would duplicate both and invite a UniversalOrchestrator. |
| Decision | **DELETE** as a package idea; **KEEP** as a vocabulary in docs |

### Multi-agent patterns

| Pattern | Current | Target | Decision |
|---|---|---|---|
| Sequential | `agent.Chain` | `workflow.Sequential` | **MOVE** |
| Concurrent / fan-out | Missing | `workflow.Parallel` using `packages/concurrency` | **KEEP** (introduce in workflow) |
| Handoff | Raw string pass | First-class `workflow.Envelope` | **REPLACE** string passing |
| Delegation | Missing | Agent-as-executor + bounded envelope | **KEEP** (introduce as usage, not a protocol) |
| Supervisor | Missing | `workflow.Supervisor` strategy (router executor + workers) | **KEEP** (strategy, not definition) |
| Group chat / swarm | Missing | **Not now** | **DELETE** as a now-item |
| Dynamic LLM swarm | Missing | **Not now** | **DELETE** as a now-item |

---

## Queue

| Field | Content |
|---|---|
| Current responsibility | Job backends + workers |
| Actual responsibility | Async infrastructure |
| Target responsibility | Same. AI/workflow **call** it; they do not fork it. |
| Reason | Already the right boundary. |
| Decision | **KEEP** |

### `queue.Manager.Chain`

| Field | Content |
|---|---|
| Current responsibility | Sequential job handlers |
| Actual responsibility | Queue sugar |
| Target responsibility | Stay in queue. Do not merge with agent/workflow chains. |
| Reason | Different domain (jobs vs executors). |
| Decision | **KEEP** (document the name clash) |

---

## Infrastructure twins

| Component | Decision | Reason |
|---|---|---|
| `ai` vs `httpclient` | **KEEP** both; do not rewrite drivers onto httpclient in this rebuild | Provider APIs are not generic JSON REST |
| `ai.Healthy` vs `packages/health` | **KEEP** both; optional adapter later | Different planes (provider vs process) |
| `ai.Observer` vs `observability` | **KEEP** both | Token economics vs HTTP metrics |
| `ai.RetryPolicy` vs `httpclient.RetryPolicy` | **KEEP** AI policy | Wired to error kinds (`RateLimit`, `Auth`) |
| New AI circuit wrapper | **UNKNOWN** / later | Fallback lists already cover many outages; breaker is additive |
| New AI ratelimit | **UNKNOWN** / later | Useful; not required to correct boundaries |
| New AI cache of completions | **DELETE** as a now-item | Correctness/staleness risk; not a boundary fix |
| `idempotency` HTTP middleware for workflows | **KEEP** package; **do not** reuse as workflow identity | Wrong abstraction (HTTP replay vs execution ID) |
| Workflow execution identity | **KEEP** (introduce on workflow) | Distinct from HTTP idempotency keys |

---

## Attention cluster

| Topic | Decision | One-line |
|---|---|---|
| AI | **KEEP** as service/runtime | Provider manager + resilient client |
| RAG | **KEEP** as library | Do not split ingest/retrieve yet |
| Agent | **KEEP** as library | Single-agent loop only |
| Memory | **KEEP** as agent interface | Do not create `memory/` |
| Graph | **MOVE** to workflow (generic); agent wrapper optional | Today it is mis-homed |
| Chain | **MOVE** / wrap | Sequential workflow |
| Queue | **KEEP** | Infrastructure |
| Workflow | **Introduce** | First-class, non-agent steps allowed |
| Orchestration package | **Do not introduce** | Patterns, not a runtime |
| MCP | **Do not implement now** | Tool adapter later |
| A2A | **Do not implement now** | Remote-agent adapter later |

---

## UNKNOWN (need product evidence, not code)

| Item | Why unknown |
|---|---|
| Durable workflow **backend** (SQL vs queue vs files) | Interface first; storage choice is operational |
| Redis vector store | Mentioned in RAG README; no users in-tree |
| Agent streaming loop | Useful UX; not a boundary defect |
| Circuit around providers | Maybe; fallback may suffice |
| MCP tool gateway | Interop demand not evidenced in this repo |
