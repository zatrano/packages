# ZATRANO AI — Target Architecture

**Status:** design for implementation (not a description of v1.7.2)  
**Date:** 2026-09-11  
**Inputs:** current-source reconstruction, responsibility map, production agent frameworks (principles only)

This document answers: **what ZATRANO AI should be if designed today in Go**, for production applications, without copying another framework.

---

## 1. Design thesis

ZATRANO AI is **four cohesive pieces**, not one platform object:

1. **AI runtime** — talk to models reliably.
2. **RAG library** — index and retrieve knowledge.
3. **Agent library** — one model + tools + conversation memory, in a bounded loop.
4. **Workflow library** — compose **any** step (function, HTTP, DB, RAG, agent, human wait).

There is **no** UniversalOrchestrator.  
There is **no** orchestration package.  
**Supervisor is a strategy**, not the architecture.

```
           application code
                  │
      ┌───────────┼───────────────┐
      ▼           ▼               ▼
   workflow     agent            rag
      │           │               │
      │           ▼               ▼
      │          ai  ◄────────────┘
      │           │
      ▼           ▼
   queue      (providers)
```

`workflow` does not import `agent` or `ai`.  
`agent` may import `workflow` to expose an agent as an executor.  
`rag` may import `ai` for embeddings.  
`ai` imports neither.

---

## 2. Principles extracted (not copied)

From Microsoft Agent Framework, Semantic Kernel process ideas, AWS multi-agent guidance, MCP, A2A, and durable-execution practice:

| Principle | Implication for ZATRANO |
|---|---|
| Two execution models | Autonomous **agent loop** vs deterministic **workflow graph**. Both exist. They compose. |
| Workflow nodes are executors | A node is not “an agent.” An agent is one executor implementation. |
| Deterministic first | Prefer explicit graphs when the process is known. Do not start with swarms. |
| Orchestration patterns are templates | Sequential / concurrent / handoff / supervisor are graphs, not runtimes. |
| Supervisor ≠ orchestration | Supervisor is centralized delegation. Handoff and sequential are also orchestration. |
| Structured handoff | Pass task, goal, artifacts, constraints, permissions — not a full transcript by default. |
| HITL is a wait state | Approval is `waiting` + resume, not `fmt.Scan` inside an executor. |
| Durability is optional | In-process run first; checkpoint interface for executions that must outlive the request. |
| MCP is vertical | Agent → tools/resources. Internal Go tools stay native. |
| A2A is horizontal | Independent remote agents. Internal agents use in-process calls. |
| Go-native | Explicit interfaces, `context.Context`, no reflection DI, no hidden global kernel. |

ZATRANO should be **smaller and stricter** than those frameworks: fewer types, compiler-enforced import direction, no magentic group-chat runtime until a real product need appears.

---

## 3. Package tree

```
packages/
  ai/              # SERVICE  — model providers, client, routing, usage
  rag/             # LIBRARY  — chunk, embed, store, retrieve, rerank, assemble
  agent/           # LIBRARY  — single-agent loop, tools, conversation memory, catalog/queue
  workflow/        # LIBRARY  — executors, graph, sequential/parallel/branch, wait, checkpoint
```

Rejected as packages (for this architecture):

| Name | Why rejected |
|---|---|
| `orchestration/` | Would duplicate workflow primitives and become a god package |
| `memory/` | Conversation, knowledge, and execution state must not share a type |
| `ingest/` | No connectors yet; vacant boundary |
| `mcp/` / `a2a/` | Protocol adapters later, outside the core loop |

---

## 4. Required questions

### A. AI runtime

**Should `ai` remain a service?** **Yes.**

It is stateful/runtime-oriented because a `Manager` owns named drivers, profiles, defaults, prices, and observers for the application process. That matches other ZATRANO services (`queue`, `cache`): `Register` + `From(app)`.

**Belongs in `ai`:** providers, models-as-strings, capabilities, profiles, routing, provider health, model-call retry/fallback, streaming, structured output, tool **schemas**, vision/image/audio/embed I/O, usage/metering.

**Does not belong in `ai`:** tool **execution**, RAG, agent loop, workflow, queue workers, MCP/A2A servers, application `/health` aggregation (optional adapter only).

It remains an **in-process** runtime. No separate AI daemon. No `contracts.App.AI()`.

### B. RAG

**Should RAG remain a library?** **Yes.** Applications construct a `Pipeline`. There is no registry of “the” vector store at boot that would justify a service.

**Should ingestion become separate?** **No** (not now). `Index` and `Query` share chunk/embed/store.

**Should retrieval become separate?** **No.**

**Should vector stores remain adapters?** **Yes.**

**Should RAG depend on AI?** **Yes, optionally, one way:** `rag → ai` for `FromAI` / LLM rerank. Core interfaces (`Embedder`) must remain implementable without `ai`. Never `ai → rag`.

### C. Agent

**Should Agent remain a library?** **Yes.** An agent is a value you `Run`. Enablement would imply one global agent, which is wrong.

**What is an Agent?** A **bounded autonomous loop**: given a user message (or envelope), it calls a chatter, may retrieve knowledge, may execute **authorized** tools, and stops on a text answer, max steps, timeout, or cancel.

| Term | Meaning (implementation, not poetry) |
|---|---|
| Agent | Configured loop: Chat, Tools, Memory, Retriever, limits, authorizer |
| Agent Runtime | Not a process. Execution is `Run` / queue handler |
| Agent State | Conversation `Memory` for that agent instance |
| Agent Context | The user turn plus optional retrieved text; for composition, a `workflow.Envelope` |
| Agent Memory | Transcript buffer (`Memory`). Not RAG. Not workflow checkpoint |
| Agent Tool | Named handler + schema + authz + timeout/cancel |
| Agent Result | Final model message, step count, transcript, tool results |
| Agent Execution | One `Run` (or one queue job wrapping `Run`) |

### D. Workflow

**First-class boundary?** **Yes.**

A workflow must contain things that are **not** agents: `Func`, HTTP, DB, RAG index/query, Agent, human approval, condition, parallel branch.

**Minimal abstraction:**

- `Executor` — `Execute(ctx, Envelope) (Envelope, error)` plus `Name()`
- `Envelope` — typed handoff (task, goal, input, output, artifacts, constraints, permissions, previous)
- `Graph` — named nodes, `Route` edges, `MaxHops`
- `Sequential` / `Parallel` / `Branch` / `Supervisor` as constructors over Graph
- `Status` — created, running, waiting, completed, failed, cancelled
- `context.Context` cancellation from graph → executor → (if agent) model/tool
- Optional `Checkpointer` and `Wait` (human)

Not required for v1 of the boundary: a visual designer, a YAML DSL, Temporal compatibility, event sourcing.

### E. Orchestration

**First-class package?** **No.**

| Concept | Where it lives |
|---|---|
| Workflow | `workflow` package — deterministic composition of **any** executor |
| Agent | `agent` package — one LLM loop |
| Multi-Agent | Workflow whose executors are agents (plus `Supervisor` strategy) |
| Orchestration | English for “composing executions.” Not a type and not a package |

| Pattern | Treatment |
|---|---|
| Sequential | Workflow primitive |
| Concurrent / fan-out-fan-in | Workflow primitive (`Parallel` + join) |
| Handoff | Envelope + `Route`; not a transcript dump |
| Delegation | Parent executor invokes a child executor with a **narrow** envelope |
| Supervisor | Strategy: router executor picks worker names until empty |
| Group collaboration / dynamic swarm | Out of scope now |

### F. Graph

**Decision:** **Move the semantic engine to `workflow`.**  
`agent.Graph` may remain as agent-typed sugar wrapping workflow, or stay as a compatibility shim. It must not remain the only composition engine.

Graph is **not** deleted; its **home** was wrong.

### G. Chain

**Decision:** Sequential workflow is the generalization. `agent.Chain` becomes a convenience for agent-only sequences (wrapper) rather than a second engine. It is **not** promoted to “the workflow.”

### H. Memory

| Kind | Owner |
|---|---|
| Conversation / working transcript | `agent.Memory` |
| Persistent conversation store | Implements `agent.Memory` (future); not a new package now |
| Retrieved knowledge | `rag` via `Retriever` |
| Execution / checkpoint state | `workflow` |
| Tool results in a turn | `agent.ToolResult` on the transcript |

Do not merge these into one `Memory` service.

---

## 5. Execution model

```
Workflow.Run(ctx, envelope)
  → for each hop
       apply step timeout (child context)
       Executor.Execute
         ├─ Func / HTTP / DB / RAG
         ├─ agent.AsExecutor → Agent.Run
         │     → Retriever
         │     → Chat (ai.Client) → retry/fallback inside ai
         │     → Tools (authorized) 
         └─ Wait → status=waiting, checkpoint, return
  → Route / join / supervisor
  → completed | failed | cancelled | waiting
```

Cancellation: parent `ctx` cancelled → step ctx cancelled → agent loop sees `ctx.Err()` → chat/tool see `ctx`.

Timeouts are **distinct**: workflow timeout, step timeout, agent timeout, `ai.Defaults.Timeout`, tool handler timeout.

Failure isolation: `Parallel` records per-child errors; join policy decides fail-fast vs partial. A failed child does not write sibling envelopes. Sequential stops on error unless a node routes to a compensation executor (explicit; no hidden sagas in v1).

---

## 6. Context model

Do **not** use `map[string]any` as the interchange type.

`workflow.Envelope` is the structured handoff. Conversation history stays inside the agent unless a step **explicitly** copies a summary into `Envelope.Previous` or an artifact.

Retrieved chunks stay in RAG/`Retriever` unless assembled into `Envelope.Input` or an artifact named by the application.

---

## 7. Tool model

Native tools remain `agent.Handler` + `ai.Tool` schema.

Required: name, description, input schema, output string, **authorization**, execution, timeout/cancel via ctx, error/status, observability via existing agent/AI hooks.

Optional now: streaming tool output, approval-required tools (can be a `Wait` step **around** the tool, or an authorizer that returns `ToolDenied`). Idempotent tool retries only when the handler declares `Retryable` **and** the agent has a retry policy (default: still no auto-retry, to avoid duplicate side effects).

MCP: future adapter that **registers** MCP tools into `Registry`. Not the internal tool type.

---

## 8. RAG integration

**Chosen design:** `Agent → Retriever`.

Not `Agent → rag.Pipeline` as a required field (the adapter `RAGRetrieve` is enough).  
Not embedding the vector store in the agent.

Workflow RAG operations are executors the **application** writes (`IndexExecutor`, `QueryExecutor`) using `rag.Pipeline`. Those types do not need to live in `agent`.

---

## 9. Human-in-the-loop

Represented as workflow status `waiting` plus a `Wait` record (`Kind`: approval | input, `StepID`, `Reason`).  
`Resume(ctx, executionID, decision)` loads checkpoint and continues.

Not implemented as a blocking callback that holds a request goroutine hostage unless the application **chooses** an in-process wait (still interruptible by `ctx`).

---

## 10. Durability

| Execution | Persistence |
|---|---|
| `ai.Client.Chat` | None (ephemeral) |
| `Agent.Run` | Optional shared `Memory`; else ephemeral |
| Queue `agent.run` | Job + optional `ResultStore` snapshot |
| `workflow.Run` default | Ephemeral |
| `workflow.Run` with `Checkpointer` | Checkpoint at hop boundaries and on wait |

Identities when durable: `ExecutionID`, `WorkflowID` (graph name), `StepID`.  
Not every workflow is durable. No JSON dump of random structs without a `Checkpoint` type.

Reuse `queue` to **start** long work. Do not treat queue payload as the checkpoint log.

---

## 11. Observability, cost, security

- Observe workflow execution, hops, waits; agent runs, tool calls; AI already has `Observer`/`UsageMeter`.
- Forwarding into `packages/observability` is an application adapter, not a new telemetry product.
- Cost: keep AI metering; workflow traces should include model/provider when the executor is an agent (via existing result/usage).
- Security: **authorizer on the agent** (`Allow(ctx, toolName) bool` or error). Retrieved text is data, never implicit tool grants. Remote agents (future A2A) are untrusted until authenticated. Secrets stay in provider config/env, not envelopes.

---

## 12. Interoperability (now vs later)

| Protocol | Role | Now? |
|---|---|---|
| MCP | External tools/resources/prompts adapter into `Registry` | **Later** (extension point only) |
| A2A | Remote independent agents | **Later** (executor adapter, never internal bus) |

Internal multi-agent = in-process executors. HTTP between them would be accidental architecture.

---

## 13. What we refuse to build

- UniversalAIManager / UniversalAgentManager / UniversalOrchestrator
- Dynamic swarm as default
- YAML agent networks
- Reflection autowiring of tools
- Second queue
- Second HTTP retry stack “for AI” beyond the existing `ai.RetryPolicy`
- `contracts.App` methods for AI/Agent/Workflow
