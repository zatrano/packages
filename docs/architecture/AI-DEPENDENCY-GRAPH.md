# ZATRANO AI — Dependency Graph

**Date:** 2026-09-11  
**Rule:** no cycles. No second implementation of queue/concurrency/observability/health/cache/httpclient unless the existing package is demonstrably the wrong plane.

---

## Allowed dependencies

```
                    ┌─────────────┐
                    │ framework/v2│
                    └──────▲──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
         ┌────┴───┐   ┌────┴────┐  ┌────┴────┐
         │   ai   │   │  queue  │  │ cache…  │  (other services)
         └────▲───┘   └────▲────┘  └─────────┘
              │            │
         ┌────┴───┐        │
         │  rag   │        │
         └────▲───┘        │
              │            │
         ┌────┴────────────┴──┐
         │       agent        │
         └────────▲───────────┘
                  │  (optional AsExecutor)
         ┌────────┴────────┐
         │    workflow     │
         └────────▲────────┘
                  │  Parallel uses
         ┌────────┴────────┐
         │  concurrency    │
         └─────────────────┘
```

Wait — **workflow must not import agent**. The arrow `agent → workflow` is correct; `workflow → agent` is **forbidden**. The diagram above shows workflow below agent, which is wrong visually.

Correct direction:

```
workflow  ──imports──►  concurrency (Parallel)
workflow  ──imports──►  (stdlib only otherwise)
workflow  ──does not──►  ai, rag, agent, queue   [queue optional later, still not agent]

ai        ──imports──►  framework
ai        ──does not──►  rag, agent, workflow, queue, observability, health, circuit, cache, httpclient

rag       ──imports──►  ai          (FromAI, LLM rerank)
rag       ──does not──►  agent, workflow

agent     ──imports──►  ai
agent     ──imports──►  rag         (RAGRetrieve adapter only)
agent     ──imports──►  queue       (Runner)
agent     ──imports──►  workflow    (AsExecutor, optional wrappers)
agent     ──does not──►  mcp/a2a (do not exist)

queue     ──does not──►  ai, agent, rag, workflow
concurrency ──does not──► AI stack
```

Application code (outside this module) may import all of them and compose:

```
app → workflow + agent + rag + ai + queue
```

---

## Forbidden dependencies

| From | To | Why |
|---|---|---|
| `ai` | `rag` | Models must not know about vector stores |
| `ai` | `agent` | Providers must not own the tool loop |
| `ai` | `workflow` | Model I/O is not process orchestration |
| `rag` | `agent` | Retrieval is usable without agents |
| `rag` | `workflow` | Pipeline is a library value; workflow nodes wrap it in **app** code |
| `workflow` | `agent` | Would force every workflow consumer to link agents; non-agent steps must compile without `agent` |
| `workflow` | `ai` | Same |
| `workflow` | `rag` | Same |
| `queue` | `agent`/`workflow` | Infrastructure must not depend on intelligence |
| Any AI package | new retry/health/metrics clones | Use or adapt existing packages; do not fork |

---

## Optional / later (still acyclic)

| Edge | When |
|---|---|
| `workflow` → `queue` | If a `workflow.Runner` is added (mirror of `agent.Runner`) |
| `ai` Observer → `observability` | Application adapter, not an `ai` import |
| `ai.Healthy` → `health.Manager.Register` | Application adapter in Boot, not a package import cycle |
| `agent` MCP adapter package | `agentmcp` → `agent` + protocol; never the reverse |
| `workflow` A2A adapter package | `a2a` → `workflow.Executor`; never core → A2A |

---

## Infrastructure reuse

| Concern | Use | Do not |
|---|---|---|
| Background jobs | `queue` | New AI queue |
| Parallel fan-out | `concurrency.Pool` or equivalent ctx-aware helper | Ad-hoc unbounded goroutines without cancel |
| HTTP metrics | `observability` | Replace `ai.UsageMeter` with HTTP metrics |
| Process health | `health` | Replace provider `Healthy` |
| Provider HTTP | keep `net/http` in drivers | Rewrite all drivers onto `httpclient` as a prerequisite |
| HTTP idempotency | `idempotency` middleware | Use as workflow ExecutionID |
| Workflow identity | `workflow` ExecutionID + Checkpointer | `zatrano.lock` or go.mod tricks |

---

## Compile-time enforcement (recommended tests)

Architecture tests in `packages` should eventually reject:

- `ai` importing `rag`, `agent`, `workflow`
- `workflow` importing `ai`, `rag`, `agent`
- `rag` importing `agent` or `workflow`

This matches how the framework kernel already forbids `github.com/zatrano/packages` imports.
