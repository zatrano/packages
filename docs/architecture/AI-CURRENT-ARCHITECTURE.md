# ZATRANO AI — Current Architecture

**Repository:** `github.com/zatrano/packages`  
**Evidence tag:** `v1.7.2` (working tree on `main`)  
**Module:** `github.com/zatrano/packages` (Go 1.25, framework pin `v2.0.28`)  
**Date:** 2026-09-11  
**Method:** Source tracing. README claims are secondary.

This document reconstructs what the code **does**, not what the names suggest.

---

## Summary

| Area | Kind today | What it actually is |
|---|---|---|
| `ai` | Service addon (`From(app)`, `addons.Register`) | In-process **provider runtime**: named drivers, profiles, retry/fallback, metering. Not an agent engine. |
| `rag` | Import-only library | **Retrieval pipeline**: chunk → embed → store → search → optional rerank → prompt text. |
| `agent` | Import-only library | **Single-agent ReAct loop** plus two multi-agent conveniences (`Chain`, `Graph`) and a queue job adapter. |
| Graph | Types inside `agent` | Conditional **agent hop** state machine. Nodes are `*Agent` only. |
| Chain | Types inside `agent` | Sequential **agent** pipeline. Not a generic workflow. Distinct from `queue.Manager.Chain`. |
| Queue | Service addon | **Background job infrastructure**. Agent uses it for `agent.run`. Queue is not a workflow engine. |

There is **no** `workflow` package and **no** `orchestration` package.

```
ai.Manager ──Driver──► OpenAI / Anthropic / Gemini / Fake
     ▲
     │ Embed / ChatJSON
rag.Pipeline ──VectorStore──► Memory / JSON / SQL / PGVector
     ▲
     │ Retriever (optional)
agent.Agent.Run ──tools──► Registry
     │
     ├── Chain / Graph  (multi-agent, agent-only nodes)
     └── Runner ──► queue.Manager  (async jobs)
```

**Import graph (AI stack):**

```
ai        → framework/v2, net/http          (no rag, agent, queue, observability, health, circuit, cache, httpclient)
rag       → ai
agent     → ai, rag, queue
queue     → database, redisx
```

No cycles.

---

## 1. AI

Package path: `ai/` (flat; 49 Go files; no subpackages).  
No package README. Root README describes it as “Chat / completion providers” — true and incomplete.

### 1.1 Lifecycle and “service vs library”

`ai` **is a service addon**.

- `init()` → `addons.Register`
- `ServiceProvider.Register` loads `DefaultConfig`, constructs `Manager`, `BootConfig`, binds `app.Container().Instance("ai", mgr)`
- `Boot` registers demo route `POST /demo/ai/chat`
- `From(app) *Manager`

It is also usable as a library: `ai.New()` + `BootConfig` / `Extend` (how `rag`/`agent` examples work).

It is **stateful in-process**: a `Manager` holds drivers, profiles, defaults, observer, prices, model metadata. It is **not** a separate OS process, worker, or `contracts.LifecycleProvider` (`Start`/`Stop` are absent). `WatchHealth` blocks on the caller’s context.

### 1.2 Provider abstraction

**Provider = named `Driver`.**

```go
type Driver interface {
    Name() string
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
```

Optional capability interfaces (type assertions, not a plugin bus):

| Interface | File | Role |
|---|---|---|
| `EmbeddingDriver` | `driver.go` | `Embed` |
| `StreamDriver` | `driver.go` | `ChatStream` |
| `ImageDriver` / `ImageEditor` | `image.go` | generate / edit / vary |
| `SpeechDriver` | `audio.go` | TTS / STT |
| `Healthy` | `health.go` | probe |
| `Capabler` | `capability.go` | explicit caps |

Constructors: `OpenAI`, `OpenAICompatible`, `Anthropic`, `Gemini`, `BuildDriver(ProviderOptions)`, plus `FakeDriver` and `LogDriver` (chat/stream decorator only).

HTTP is **raw `net/http`**. `packages/httpclient` is not used.

### 1.3 Model abstraction

There is **no `Model` type**. A model is a string on `ChatRequest.Model` / `EmbedRequest.Model`, filled from `Defaults` or `Profile`. Optional `ModelInfo` metadata via `SetModels` for ranking (`PreferCheapest` / `PreferSmartest`). The runtime does not load model cards or enforce context windows.

### 1.4 Capability system

`CapChat | CapEmbed | CapStream | CapTools | CapJSON | CapVision | CapImage | CapSpeech`.

`InferCapabilities` uses `Capabler` if present, otherwise interface probes and type switches. The matrix is **asymmetric**: tools/JSON/vision/image/speech are strongest on OpenAI + Fake; Anthropic has tools/vision/stream; Gemini has embed/stream/vision; Gemini has no `CapTools`.

### 1.5 Profiles

`Profile` is an ordered **fallback provider list** plus optional model / temperature / max_tokens overrides (`profile.go`). It is routing config, not an “agent persona” and not a prompt template.

### 1.6 Routing

| Entry | Behavior |
|---|---|
| `Using(name)` | Single provider |
| `Profile(name)` | Profile provider order |
| `ProfileLive` / `UsingLive` | Filter by capability then health; empty filter **fail-open** to original chain unless `FailIfNone` |
| `PreferCheapest` / `PreferSmartest` | Sort name lists; do not themselves produce a `Client` |
| `WatchHealth` | Periodically **rewrites** the profile provider list |

Routing is **not** request-time load balancing. It is static order + optional live filter.

### 1.7 Health

Package-local, not `packages/health`.

- Optional `Healthy.Health(ctx)`
- OpenAI `GET /models`, Anthropic `GET /v1/models`, Gemini `GET /v1beta/models`
- Missing probe → `Skipped=true, OK=true`
- `CheckHealth` / `CheckHealthProfile` / `WatchHealth`

AI provider health is **not** registered on the application `/health` surface.

### 1.8 Retry

`RetryPolicy` on `Defaults` (`retry.go`). Per-provider `callWithRetry` before fallback. Retries only `Retryable` kinds: `KindRateLimit`, `KindUnavailable`. Exponential backoff + jitter; `Retry-After` honored.

This is a **second** retry type next to `httpclient.RetryPolicy`.

### 1.9 Fallback

Ordered provider chain. Stops (no further fallback) when `Fallbackable` is false: `KindAuth`, `KindInvalid`, cancel. `DeadlineExceeded` falls back only if `Defaults.FallbackOnTimeout`.

Streaming: fallback only at **setup**. Mid-stream errors stay on the channel; provider is not switched.

### 1.10 Streaming

`StreamDriver.ChatStream` → `<-chan StreamChunk`. OpenAI SSE, Anthropic, Gemini, Fake. `CollectStream` assembles text and tool-call deltas. Agents **do not** use streaming.

### 1.11 Structured output

`ResponseFormat` on `ChatRequest`; `ChatJSON` + `DecodeJSON` (markdown fence strip). Capability `CapJSON` is OpenAI-oriented. Anthropic/Gemini are not first-class JSON-mode peers.

### 1.12 Tools

`ai` owns **schema and wire format** (`Tool`, `ToolCall`, `ToolChoice`). It does **not** execute tools. Execution is `agent.Registry`. This split is correct.

### 1.13 Vision / image / audio / embedding

| Modality | Types | Who implements |
|---|---|---|
| Vision | `Message.Parts`, `UserVision` | OpenAI, Anthropic, Gemini, Fake |
| Image gen/edit/vary | `ImageDriver` | OpenAI, Fake |
| TTS/STT | `SpeechDriver` | OpenAI, Fake |
| Embedding | `EmbeddingDriver` | OpenAI, Gemini, Fake (not Anthropic) |

### 1.14 Usage, metering, pricing, observation

- `Usage` token counts on responses/streams
- `UsageMeter` implements `Observer`; aggregates by provider/model/op; `EstimatedUSD` via `PriceTable`
- `Manager.SetPrices` — static table, not live billing
- `Observer` hooks (`OnRequest` / `OnResult`) — **not** `packages/observability`

### 1.15 Concurrency and context

- `Manager.mu` `RWMutex` for registry
- Stream readers in goroutines; `bindStreamCancel`
- Nil context → `Background`; optional `Defaults.Timeout` via `WithTimeout`
- No worker pool; `packages/concurrency` unused

### 1.16 What belongs together vs what does not

**Cohesive in `ai`:** driver interfaces, request/response types, provider HTTP adapters, profile fallback, retry, usage.

**Stuffed into the same flat package but different jobs:** multimodal OpenAI-only APIs, health watcher that mutates profiles, demo HTTP route, cost ranking helpers, log decorator.

**Does not belong in `ai` (and is correctly elsewhere):** tool execution, RAG, agent loop, graphs, queues.

**Reimplemented instead of reused:** HTTP client+retry, health checks, metrics observers.

---

## 2. RAG

Package path: `rag/`. Import-only (`no addon registration`). README matches the pipeline.

### 2.1 Actual pipeline

```
Document
  → Chunker.Split          (TextChunker default: 800/100/40 runes)
  → Embedder.Embed         (batch all chunk texts)
  → VectorStore.Upsert
```

```
question
  → Embedder.Embed([question])
  → VectorStore.Search(vec, topK)     → []Hit
  → optional Reranker.Rerank          (QueryWith only)
  → FormatContext                     → prompt string
```

**Filtering:** no metadata filter, no score threshold, no hybrid/MMR. “Filter” is top-K / `FinalTop` truncation.

### 2.2 Interfaces

| Interface | Responsibility |
|---|---|
| `Chunker` | `Split(Document) []Chunk` |
| `Embedder` | `Embed(ctx, texts) ([][]float64, error)` |
| `VectorStore` | `Upsert`, `Search`, `DeleteDocument`, `Len` |
| `Reranker` | `Rerank(ctx, query, hits)` |
| `PairScorer` / `ChatJSON` | Cross-encoder / LLM rerank adapters |

`FromAI(*ai.Manager, model)` implements `Embedder`. `FromAIChat` implements LLM rerank chat.

### 2.3 Stores

| Store | Search | Persistence |
|---|---|---|
| `MemoryStore` | Brute-force cosine | RAM + `RWMutex` |
| `JSONFileStore` | Same | Atomic JSON file |
| `SQLStore` | Load all rows, cosine in Go | `database/sql`, embeddings as JSON text |
| `PGVectorStore` | pgvector `<=>` / `<->` / `<#>` | `vector(N)`, optional IVFFlat |

Redis vector store: README mentions as an exercise; **not implemented**.

### 2.4 Query flow (call path)

`Pipeline.Query` → embed → `Store.Search`.  
`Pipeline.QueryWith` → `Query(fetch)` → `Rerank` → truncate.  
`FormatContext` concatenates `[n] (score=… doc=…)` + text, `maxChars` default 6000. Metadata is **not** rendered.

### 2.5 Embedding integration

Compile-time import of `ai` in `ai_embed.go` and `rerank_llm.go`. Direction: **`rag → ai` only**. A `FuncEmbedder` can avoid `ai` entirely.

### 2.6 Reranking

`KeywordReranker`, `CrossEncoderReranker`, `LLMReranker`. LLM rerank **fails open** (keeps retrieval order). Cross-encoder scores **sequentially**. Agent `RAGRetrieve` calls `Query`, **not** `QueryWith`.

### 2.7 Persistence, metadata, failure, concurrency

- Metadata: `map[string]string` copied onto chunks; stored; **never queried**
- Index requires Embed+Store; empty question errors; vector count mismatch errors
- `ctx.Err()` checked in stores/rerank
- Pipeline itself has no lock; stores that need it use mutex or the database
- No ingestion job, no incremental loader, no document-format parsers

### 2.8 Extensibility

Correct: swap Chunker / Embedder / Store / Reranker.  
Missing as first-class stages: ingest adapters, hybrid search, metadata filters, retrieval-time ACL.

### 2.9 Library vs service

**Library.** No `From(app)`, no enablement. The application owns the `Pipeline` value.

---

## 3. Agent

Package path: `agent/`. Import-only. README is accurate for the implemented surface.

### 3.1 Execution loop (`Agent.Run`)

Source: `agent/agent.go`.

1. Require `Chat`, non-empty user message; nil ctx → `Background`
2. Memory nil → ephemeral `BufferMemory`
3. `MaxSteps <= 0` → **6**
4. If memory empty and `System` set → append system message
5. Optional `Retrieve.Retrieve`; on success prefix `Context:\n…\n\nQuestion:`; **retrieve error aborts Run**
6. Append user message
7. For `step = 1..maxSteps`:
   - Build `ai.ChatRequest` with tools
   - Last step: `ToolChoiceNone` unless `AllowToolsOnFinal`
   - `Chat.Chat` (**synchronous only**)
   - Chat error → abort
   - No tool calls → append assistant, return `Result`
   - Tool calls → append assistant tool-calls; `Registry.ExecuteResult` per call; append tool message; **tool error does not abort**
8. Exhaustion → `Result` + `error: max steps reached`

**There is no `ctx.Err()` check at the top of the loop.** Cancellation works only if `Chat` or the tool handler observes `ctx`.

### 3.2 Terminology as implemented

| Name | Implementation meaning |
|---|---|
| Agent | Struct of Chat + Tools + Memory + Retrieve + limits. Not a process. |
| Agent Runtime | None. `Run` is a function call. |
| Agent State | The `Memory` message list. No typed working state. |
| Agent Context | Concatenated user string (+ optional RAG blob). |
| Agent Memory | `Memory` interface: `Messages` / `Append` / `Clear`. One impl: `BufferMemory`. |
| Agent Tool | `ai.Tool` schema + `Handler` in `Registry`. |
| Agent Result | `Result{Response, Steps, Messages, ToolResults}` |
| Agent Execution | One `Run` invocation. Queue jobs wrap that. |

There is **no planner**, no explicit goal object, no scratch working memory distinct from the transcript.

### 3.3 Tool execution

`Registry.ExecuteResult`: lookup by name; missing → `ToolInvalid`; handler error classified (`ToolTimeout` if context deadline/cancel, else `ToolError` unless `*ExecError`). `Retryable` is **metadata only** — the loop never retries tools.

**Authorization:** no Gate/Policy. Built-ins only: `web_fetch` HTTPS-only (`ToolDenied`); `file_search` sandboxed under `Root`. An agent with a registry can run **every registered tool**.

### 3.4 Memory

`BufferMemory`: in-process slice, `MaxKeep` trim, leading system message preserved. No Redis/SQL/session memory.  
Separate: `ResultStore` persists **job outcomes** (`StoredRun`), not transcripts.

### 3.5 Retrieve

`Retriever` returns a string. `RAGRetrieve` → `Pipeline.Query` + `FormatContext`. Rerank is unused on this path. `FuncRetriever` is the extension point.

### 3.6 Planning, termination, errors, retries, streaming

- Planning: none (model-driven tool choice only)
- Termination: text reply, or max steps error
- Chat/retrieve errors fail the run
- LLM retries: only whatever `*ai.Client` already does
- Streaming: **not used** (`Chatter` is `Chat` only)

### 3.7 Queue integration

`agent/queue.go` **reuses** `packages/queue` correctly.

- `Catalog` of named `*Agent`
- `Runner.RegisterQueue` handler default `agent.run`
- `PushRun` enqueues `RunJob`
- Optional `ResultStore` + `OnResult`

Graph and Chain are **not** queue-backed. Queue is async **job transport**, not orchestration.

### 3.8 Multi-agent, HITL, durability

| Pattern | Present? |
|---|---|
| Sequential (`Chain`) | Yes, agent-only |
| Conditional hops (`Graph`) | Yes, agent-only |
| Concurrent / fan-out | No |
| Supervisor / delegation API | No |
| Structured handoff | No (raw previous assistant text) |
| Human-in-the-loop | No |
| Checkpoint / resume | No |
| Mid-run cancel hook | No |

---

## 4. Graph / Chain

**Do not read the names as “workflow.”**

### 4.1 Graph

`agent.Graph`: named nodes, each **must** have `*Agent`. `Route(ctx, GraphState) (next, error)` picks the next name; empty/`nil` ends. `MaxHops` default 8 is the only cycle brake (`Visits` is counted, not enforced as a DAG). `GraphState` is `Input/Previous/Output` plus `Data map[string]string`.

**Semantic responsibility:** in-process **conditional multi-agent router**. Not a general executor graph, not durable, not parallel, not LangGraph.

### 4.2 Chain

`agent.Chain`: fixed `[]ChainStep`, each `*Agent`. Prompt fn or previous assistant text. Stops on first error. Fresh memory per step unless Agents share a `Memory`. `CatalogChain` builds from names.

**Semantic responsibility:** sequential multi-agent pipeline. Not an LLM “chain” library, not a workflow DSL.

### 4.3 Name collision

`queue.Manager.Chain` runs **queue jobs** sequentially. Unrelated to `agent.Chain`.

---

## 5. Queue

`packages/queue` is **infrastructure**: named handlers, `sync` / `database` / `redis` backends, `queue:work`, failed jobs.

Relative to AI:

| Question | Answer |
|---|---|
| Infrastructure? | Yes |
| Workflow execution engine? | No |
| Agent execution engine? | No; it hosts `agent.run` jobs |
| Asynchronous job execution? | Yes |
| Orchestration infrastructure? | Only as “run this handler later,” including `PushChain` of jobs |

**Do not duplicate queue.** Agent already consumes it. A future workflow runner should do the same.

---

## 6. Existing infrastructure vs AI stack

| Package | Kind | Used by AI stack? | Notes |
|---|---|---|---|
| `queue` | service | **Yes** (`agent.Runner`) | Correct reuse |
| `concurrency` | library | No | `Map`/`Pool` exist; Graph/Chain are serial |
| `observability` | service | No | HTTP/Prometheus metrics; AI has `Observer`/`UsageMeter` |
| `health` | service | No | App health checks; AI has provider `Healthy` |
| `circuit` | service | No | AI uses fallback lists instead of breakers |
| `ratelimit` | service | No | No provider RPM limiter |
| `idempotency` | library | No | HTTP Idempotency-Key middleware; not execution IDs |
| `cache` | service | No | RAG/result stores are their own |
| `httpclient` | service | No | Drivers and `web_fetch` use `net/http` |
| `bus` | service | No | Sync command bus; not workflow |
| `lock` / `facts` | service | No | |

Domain-specific AI health/usage observers are reasonable. Unwired **HTTP client, circuit, app health, and rate limit** are the real gaps — not missing a second queue.

---

## 7. Defects that matter architecturally

1. **Agent owns Graph/Chain**, so every composed step must be an LLM agent. HTTP, SQL, RAG index, and human approval cannot be first-class steps.
2. **No workflow boundary.** Production processes that mix deterministic work with agents have nowhere correct to live.
3. **Context handoff is a string.** Chains pass previous assistant text; graphs pass `Previous`/`Output`. No task/goal/permissions envelope.
4. **Memory kinds are collapsed** into one transcript buffer.
5. **Tool authorization is absent** at the agent boundary.
6. **Cancellation is incomplete** in the agent loop.
7. **Durability is a final JSON snapshot** of queue jobs, not execution checkpointing.
8. **Infrastructure twins:** retry, health, metrics, HTTP.
9. **RAGRetrieve ignores rerank** and metadata filters do not exist.
10. **Streaming, HITL, concurrent fan-out, supervisor** are unimplemented — and Graph cannot grow into them without becoming a disguised workflow engine.

These are boundary problems. They are not fixed by adding more methods to `Agent`.
