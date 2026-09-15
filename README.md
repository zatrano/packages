<p align="center">
  <strong>ZATRANO packages</strong>
</p>

<p align="center">
  <em>Official addons for the ZATRANO kernel. Services boot when they are both imported and enabled. Leave a package out of the binary, and it cannot boot.</em>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/zatrano/packages"><img src="https://img.shields.io/badge/golang-1.25+-00ADD8?logo=go&logoColor=white" alt="Golang"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/zatrano/framework"><img src="https://img.shields.io/badge/kernel-zatrano%2Fframework-222?logo=go" alt="Framework"></a>
</p>

<p align="center">
  <a href="https://github.com/zatrano/framework">Framework</a>
  ·
  <a href="https://zatrano.com/docs">Docs</a>
  ·
  <a href="https://github.com/zatrano/framework/blob/main/PACKAGES.md">Package guide</a>
</p>

<p align="center">
  Current stable: <a href="https://github.com/zatrano/packages/releases/tag/v1.9.0"><code>v1.9.0</code></a>
  ·
  Framework pin: <code>github.com/zatrano/framework/v2@v2.0.28</code>
</p>

---

This module is [github.com/zatrano/packages](https://github.com/zatrano/packages). It is not the kernel.

`ai`, `rag`, `agent`, and `workflow` are **experimental**: they have not completed the same security review as the rest of the ecosystem. See [PACKAGES.md](PACKAGES.md).

It is a **v1** Go module: the import path has no `/v2` suffix and must not be tagged `v2.x`. Current stable release: **`v1.9.0`**. It requires `github.com/zatrano/framework/v2 v2.0.28`. The two modules version independently. Nested package `VERSION` files (drivers, auth, …) are informational only.

Releases are created only with `scripts/release.sh`. Do not run `git tag` by hand. Preview with `scripts/release.sh --dry-run vX.Y.Z` (nested: `scripts/release.sh --dry-run database/driver/sqlite/vX.Y.Z`).

The kernel lives in [github.com/zatrano/framework/v2](https://github.com/zatrano/framework): HTTP, routing, middleware, config, the CLI, `zatrano new`. Everything that used to look like “the rest of the framework” — sessions, auth, database, views, queues, AI — lives **here**, next to OAuth and the import-only helpers.

The two modules cannot be merged: this one already requires the framework.

```text
  github.com/zatrano/framework/v2          github.com/zatrano/packages
  ────────────────────────────          ──────────────────────────
  kernel/http  kernel/routing           session  auth  database  view
  contracts    bootstrap.App()          queue    ai    oauth     social
  zatrano new                           toolkit   resources
          │                                        ▲
          │         blank-import                   │
          └────────────────────────────────────────┘
```

## How a package turns on

Each service package registers itself in `init()`:

```go
func init() {
    addons.Register(addons.Meta{
        Name:    "session",
        Factory: func() contracts.Provider { return &ServiceProvider{} },
    })
}
```

Environment keys for a package live in that package's `.env.example`. They are not part of the kernel example file.

The **application** blank-imports the package and lists it in `bootstrap/enabled.go`. `bootstrap.App()` boots **Enabled ∩ Imported**. If you do not import it, it is not compiled in and it does not boot. Without an enablement manifest, `App()` falls back to every imported bootable addon.

```go
import (
    _ "github.com/zatrano/packages/session"
    _ "github.com/zatrano/packages/auth"
    _ "github.com/zatrano/packages/database"
    _ "github.com/zatrano/packages/view"
)

app := bootstrap.App(bootstrap.WithProviders(providers.All()...))
sess := session.From(app)
```

From the app CLI:

```bash
go get github.com/zatrano/framework/v2@v2.5.0
go get github.com/zatrano/packages@v1.9.0
go run ./cmd/app package:enable auth
go run ./cmd/app package:list
go run ./cmd/app package:doctor
```

`v1.9.0` is the current public packages tag. Do not use `@v1.7.0` for new apps: that historical tag requires an unpublished nested SQLite module and must not be retagged. Upgrade path: `v1.7.0` → `v1.7.1` → `v1.7.2` → `v1.8.0` → `v1.9.0`. Do not use `@main` for application consumption. A sibling framework checkout is only for this repository's development (`go.work` / `replace`).

`package:enable` writes the blank-import into `bootstrap/addons.go` and merges `packages/<name>/.env.example` into the app `.env.example` (existing keys are not overwritten). Libraries are never enabled — you just `import` them.

| Kind | In the binary when | Examples |
| --- | --- | --- |
| **Service** | Blank-import **and** list in `EnabledAddons` (`Enabled ∩ Imported`) | `auth`, `database`, `queue`, `ai`, `oauth` |
| **Library** | You `import` it in your code | `toolkit/str`, `resources`, `rag` |
| **Heavy** | Own `go.mod`, only when needed | `webauthn`, `mongo`, `qr`, SQL drivers |

Resolve services with `From(app)` helpers. Do not expect `app.Auth()` on the kernel.

```go
auth.From(app)
session.From(app)
database.Migrator(app)
notification.From(app)
```

There is no `mail` package. Send email through `notification` with `Channels: ["mail"]`.

## Packages

HTTP, routing, middleware, config, and the CLI live in the [framework](https://github.com/zatrano/framework). Everything below is this module.

**Service** — blank-import (`package:enable`). **Library** — `import` in your code. **Heavy** — own `go.mod`, only when needed.

### Web

| Package | Kind | What it does |
| --- | --- | --- |
| [`session`](session) | service | Per-visitor server-side sessions (file driver by default) |
| [`flash`](flash) | service | One-request success/error messages and old input |
| [`validation`](validation) | service | Form and request validation (pipe rules, FormRequest) |
| [`view`](view) | service | HTML templates (`views/`) |
| [`assets`](assets) | service | Vite/Mix manifest URLs in views |
| [`localization`](localization) | service | JSON translations under `lang/` |
| [`filesystem`](filesystem) | service | Named disks (`local`, `public`, …) |
| [`pages`](pages) | library | File-based static pages registered on the router |

### Identity

| Package | Kind | What it does |
| --- | --- | --- |
| [`auth`](auth) | service | Session login, register, password reset, email verify, lockout, MFA, remember-me |
| [`authorization`](authorization) | service | Gates and policies after authentication |
| [`hashing`](hashing) | service | bcrypt password hashes |
| [`apitoken`](apitoken) | service | Personal access tokens (Bearer) |
| [`social`](social) | service | GitHub/Google OAuth **client** login |
| [`oauth`](oauth) | service | OAuth2 **authorization server** (not social login) |
| [`webauthn`](webauthn) | heavy | Passkey registration and login (own `go.mod`) |
| [`consent`](consent) | library | Cookie-consent helpers |
| [`fingerprint`](fingerprint) | library | Device fingerprint helpers |
| [`honeypot`](honeypot) | library | Hidden spam-trap fields on forms |

### Data

| Package | Kind | What it does |
| --- | --- | --- |
| [`database`](database) | service | SQL connections, query builder, schema, migrations, seeders |
| [`orm`](orm) | service | Models, relations, eager loading, soft deletes |
| [`factory`](factory) | library | Model factories for tests and seeders |
| [`cache`](cache) | service | Temporary key/value store (file / memory / redis) |
| [`redisx`](redisx) | library | Redis client helper; cache owns the connection |
| [`mongo`](mongo) | heavy | Document store client, not SQL ORM (own `go.mod`) |

### Async

| Package | Kind | What it does |
| --- | --- | --- |
| [`queue`](queue) | service | Named background jobs (sync / database / redis) |
| [`events`](events) | service | Sync event dispatch and model observers |
| [`notification`](notification) | service | Mail, SMS, push, database inbox, broadcast — this is how you send email |
| [`broadcasting`](broadcasting) | service | Channel events to log/file/null drivers (not a WebSocket server) |
| [`schedule`](schedule) | service | Cron-like tasks via `schedule:run` (no long-running daemon) |

### HTTP extras

| Package | Kind | What it does |
| --- | --- | --- |
| [`httpclient`](httpclient) | service | Outbound HTTP with JSON, retries, and fakes |
| [`ratelimit`](ratelimit) | service | Named in-process rate limiters |
| [`url`](url) | service | Absolute URLs, named routes, signed links |
| [`maintenance`](maintenance) | service | Downtime page (`down` / `up`) |
| [`health`](health) | service | `/health` style checks |
| [`observability`](observability) | service | Metrics collection |
| [`idempotency`](idempotency) | library | Idempotent POST keys |
| [`negotiate`](negotiate) | library | `Accept` content negotiation |
| [`websocket`](websocket) | library | WebSocket upgrade helpers |

### Intelligence

| Package | Kind | What it does |
| --- | --- | --- |
| [`ai`](ai) | service | Chat / completion providers |
| [`rag`](rag) | library | Chunking, embed pipeline, vector store helpers |
| [`agent`](agent) | library | Agent loop, tools, conversation memory |
| [`workflow`](workflow) | library | Generic process graphs (`agent.AsExecutor`; not durable) |

### Product addons

| Package | Kind | What it does |
| --- | --- | --- |
| [`audit`](audit) | service | Request and audit event logging |
| [`backup`](backup) | service | Database backup/restore via native CLIs |
| [`graphql`](graphql) | service | GraphQL schema and queries |
| [`webhooks`](webhooks) | service | Signed outbound webhook delivery |

### Libraries

| Package | Kind | What it does |
| --- | --- | --- |
| [`api`](api) | library | API versioning helpers |
| [`browser`](browser) | library | Headless browser test helpers |
| [`export`](export) | library | CSV/XLSX import and export |
| [`image`](image) | library | Resize and encode images |
| [`jsonapi`](jsonapi) | library | JSON:API document helpers |
| [`openapi`](openapi) | library | OpenAPI generate/serve helpers |
| [`pdf`](pdf) | library | PDF generate and inline view |
| [`qr`](qr) | heavy | QR code images (own `go.mod`) |
| [`resources`](resources) | library | API resource transformers |
| [`testing`](testing) | library | Feature tests (`Get("/").AssertOK()`) |
| [`toolkit/arr`](toolkit/arr) | library | Array/slice helpers (moved from kernel `support/arr`) |
| [`toolkit/bloom`](toolkit/bloom) | library | Bloom filter |
| [`toolkit/circuit`](toolkit/circuit) | library | Circuit breaker (`New` / `Breaker` / `Execute`) |
| [`toolkit/collection`](toolkit/collection) | library | In-memory collections (`Filter`, `Map`, …) |
| [`toolkit/color`](toolkit/color) | library | Color helpers |
| [`toolkit/concurrency`](toolkit/concurrency) | library | Parallel tasks: `Run` / `Map` / `Pool` |
| [`toolkit/cron`](toolkit/cron) | library | Cron expression parse and match |
| [`toolkit/date`](toolkit/date) | library | Date/time helpers |
| [`toolkit/debug`](toolkit/debug) | library | Dump helpers |
| [`toolkit/enums`](toolkit/enums) | library | String-backed enums with labels |
| [`toolkit/hashid`](toolkit/hashid) | library | Obfuscated reversible public IDs |
| [`toolkit/html`](toolkit/html) | library | HTML helpers |
| [`toolkit/jsonschema`](toolkit/jsonschema) | library | JSON Schema subset validation |
| [`toolkit/lock`](toolkit/lock) | library | Process-local named locks |
| [`toolkit/markdown`](toolkit/markdown) | library | Markdown → HTML |
| [`toolkit/money`](toolkit/money) | library | Money helpers |
| [`toolkit/num`](toolkit/num) | library | Number helpers |
| [`toolkit/process`](toolkit/process) | library | Run OS commands |
| [`toolkit/str`](toolkit/str) | library | String helpers (moved from kernel `support/str`) |
| [`toolkit/timing`](toolkit/timing) | library | Server-Timing measurements |
| [`toolkit/zip`](toolkit/zip) | library | ZIP create and extract (`package zipx`) |

`bootutil` is an internal coerce/CLI helper. It is not a consumer package.

The name list lives in the framework CLI catalog (`console/describe/catalog.go`). Kernel `kernel/catalog.go` is primitives only. This tree is the code.

## Nested modules

These have their own `go.mod`:

```text
mongo/
webauthn/
qr/
database/driver/sqlite/
database/driver/mysql/
database/driver/pgsql/
database/driver/mssql/
database/driver/oracle/
database/driver/mongo/
```

Public consumption of a nested driver (after that module is tagged correctly) is a separate `go get`, for example:

```text
go get github.com/zatrano/packages/database/driver/sqlite@v1.0.0
```

The Git tag for that module must be `database/driver/sqlite/v1.0.0`. A tag named `packages/database/driver/sqlite/v1.0.0` is not a valid Go module version. Publishing those nested tags is a **separate release operation**; it is not part of root `v1.9.0`.

`db:setup` in the app pulls the driver you choose. None of them are linked until then — including SQLite. Root `github.com/zatrano/packages@v1.9.0` does not require those nested module paths.

## Local development

This repository may use a sibling framework checkout and `go.work`:

```text
replace github.com/zatrano/framework/v2 => ../framework
```

That replace is **development-only**. Public consumers resolve `github.com/zatrano/framework/v2@v2.0.28` from the module proxy; they do not clone this tree next to the framework.

```bash
go test ./...
```

Work lands on **`main`**, same default branch as the framework.

## Import path

```go
import "github.com/zatrano/packages/auth"
import "github.com/zatrano/framework/v2/kernel/http"   // kernel, not this module
```

Kernel types (`http.Request`, the router, CSRF) stay in the framework. This module implements the rest and talks to the kernel through `contracts.App`.

## Docs

| | |
| --- | --- |
| [Framework README](https://github.com/zatrano/framework) | Kernel, `zatrano new`, boot rule |
| [PACKAGES.md](https://github.com/zatrano/framework/blob/main/PACKAGES.md) | Purpose and usage per package |
| [zatrano.com/docs](https://zatrano.com/docs) | Product guides |
| [Package ecosystem](https://zatrano.com/docs/package-ecosystem) | Enable, doctor, presets |

## License

Same terms as the framework: MIT · Copyright (c) 2026 Serhan KARAKOÇ
