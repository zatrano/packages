# ZATRANO AI — Architecture Decision

**Date:** 2026-09-11  
**Evidence:** `github.com/zatrano/packages` @ v1.7.2 source, plus principle research (Microsoft Agent Framework, AWS multi-agent guidance, MCP, A2A, durable execution).  
**This report is the implementation gate.** Production source was not modified to produce it.

---

## Executive decision

```text
HYBRID REBUILD
```

**Why not EVOLVE:** `agent.Graph` / `agent.Chain` cannot host non-agent work. Growing them into a workflow engine would make Agent the process runtime — the exact confusion this audit forbids. Cancellation, tool authz, and structured handoff are missing for production use. Those are not README issues.

**Why not REBUILD (greenfield):** OpenAI/Anthropic/Gemini drivers, `rag.Pipeline`, vector stores, and `Agent.Run` are coherent and tested. Rewriting them would spend risk on I/O, not on boundaries.

**What HYBRID means:** introduce `workflow` as a first-class library; keep `ai` as the model service; keep `rag` and `agent` as libraries; move composition semantics out of “everything is an Agent”; harden the agent loop; do not create an orchestration package, MCP, A2A, or a second queue.

---

## Explicit YES/NO

| Question | Answer | Reason |
|---|---|---|
| Is AI a service/runtime? | **YES** | `Manager` is application-scoped provider state; `From(app)` matches other addons. In-process, not a daemon. |
| Is RAG a library? | **YES** | Caller owns `Pipeline`. No boot registry needed. |
| Is Agent a library? | **YES** | Many agents per app. Enablement would imply one global agent. |
| Is Workflow first-class? | **YES** | Processes include non-agent steps. That cannot live under `agent`. |
| Is Orchestration first-class? | **NO** (not a package) | Sequential/parallel are workflow primitives. A third runtime duplicates them. |
| Is Graph retained? | **YES** (engine in `workflow`; `agent.Graph` may wrap) | Semantics are useful; home was wrong. |
| Is Chain retained? | **YES** as convenience; sequential workflow is canonical | Do not keep two engines. |
| Is Queue reused? | **YES** | Already the job plane; agent.Runner is the pattern. |
| Is Memory separate? | **NO** (not a package) | Conversation vs knowledge vs execution state must stay distinct types. |
| Is Multi-Agent first-class? | **YES** as patterns on workflow+agent | Not a fourth runtime. |
| Is Supervisor a strategy? | **YES** | Router executor + workers. Not the definition of orchestration. |
| Is Handoff first-class? | **YES** (`Envelope`) | String transcripts are not a protocol. |
| Is Concurrent orchestration first-class? | **YES** (`workflow.Parallel`) | Missing today; required for fan-out isolation. |
| Is Durable Workflow first-class? | **YES as interface; NO as mandatory persistence** | Checkpoint optional. Default run is ephemeral. |
| Is Human-in-the-loop first-class? | **YES** (`waiting` + Resume) | Not a blocking hack. |
| Is MCP required now? | **NO** | Vertical tool adapter later. Native tools stay Go handlers. |
| Is A2A required now? | **NO** | Remote-agent adapter later. Internal agents are in-process. |
| Is dynamic orchestration required now? | **NO** | Deterministic graphs first. Supervisor is bounded, not a swarm. |
| Is persistence required for all workflows? | **NO** | Only when Checkpointer is set or queue job snapshots are used. |
| Is a separate runtime required? | **NO** (no extra process) | `ai` is in-process service; workflow/agent are libraries. Queue workers already exist. |

---

## Package decision

```
ai/         service    KEEP
rag/        library    KEEP
agent/      library    KEEP + harden + AsExecutor
workflow/   library    INTRODUCE
orchestration/         DO NOT CREATE
memory/                DO NOT CREATE
```

---

## Implementation authorization

Increment 1 of [AI-REBUILD-PLAN.md](AI-REBUILD-PLAN.md) is **authorized**:

- Create `workflow/`
- Harden `agent` (cancel, timeout, authorizer, AsExecutor)
- Additive RAG retrieve option
- Architecture import tests
- Documentation/CHANGELOG

**Forbidden in increment 1:** MCP/A2A packages, deleting provider drivers, modifying acquire/Apply, kernel `contracts.App` growth, automatic enablement, `zatrano.lock`.

---

## Research notes (principles kept)

1. Microsoft: two models (agent loop vs executor graph) that compose — adopted. Magentic group-chat runtime — not adopted.
2. AWS: supervisor is a collaboration mode; Step Functions-like graphs own durable process — adopted as workflow+optional checkpoint, not as AWS clone.
3. MCP vs A2A: vertical tools vs horizontal remote agents — adopted as **later adapters**, not internal buses.
4. Durable execution: identity + checkpoint + wait — adopted as interfaces, not as a Temporal rewrite.

---

## Gate

After this file exists, production code may change **in line with increment 1 only**.
