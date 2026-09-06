<p align="center">
  <strong>ZATRANO packages</strong>
</p>

<p align="center">
  <em>Official addons for the ZATRANO kernel. Import a package, it boots. Leave it out, it is not in the binary.</em>
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
  Active line: <a href="https://github.com/zatrano/packages/tree/main"><code>main</code></a>
  ·
  <code>go get github.com/zatrano/packages@main</code>
</p>

---

This module is **`github.com/zatrano/packages`**. It is not the kernel.

The kernel lives in [`github.com/zatrano/framework/v2`](https://github.com/zatrano/framework): HTTP, routing, middleware, config, the CLI, `zatrano new`. Everything that used to look like “the rest of the framework” — sessions, auth, database, views, queues, AI — lives **here**, next to OAuth, billing, and the import-only helpers.

The two modules cannot be merged: this one already requires the framework.

```text
  github.com/zatrano/framework/v2          github.com/zatrano/packages
  ────────────────────────────          ──────────────────────────
  kernel/http  kernel/routing           session  auth  database  view
  contracts    bootstrap.App()          queue    ai    oauth     billing
  zatrano new                           collection  totp  resources
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

The **application** blank-imports the package. `bootstrap.App()` then loads every registered provider. If you do not import it, it is not compiled in and it does not boot.

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
go get github.com/zatrano/packages@main
go run ./cmd/app package:enable auth
go run ./cmd/app package:list
go run ./cmd/app package:doctor
```

`package:enable` writes the blank-import into `bootstrap/addons.go` and merges `packages/<name>/.env.example` into the app `.env.example` (existing keys are not overwritten). Libraries are never enabled — you just `import` them.

| Kind | In the binary when | Examples |
| --- | --- | --- |
| **Service** | You blank-import it | `auth`, `database`, `queue`, `ai`, `oauth` |
| **Library** | You `import` it in your code | `collection`, `totp`, `resources`, `rag` |
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
| [`otp`](otp) | service | Short numeric OTPs (you deliver via SMS/mail) |
| [`totp`](totp) | library | Authenticator-app TOTP secrets and codes |
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
| [`redisx`](redisx) | service | Shared Redis client used by cache and queue |
| [`mongo`](mongo) | heavy | Document store client, not SQL ORM (own `go.mod`) |
| [`search`](search) | service | In-memory search index |
| [`hashid`](hashid) | service | Obfuscate numeric IDs for public URLs |
| [`enums`](enums) | service | String-backed enums with labels |

### Async

| Package | Kind | What it does |
| --- | --- | --- |
| [`queue`](queue) | service | Named background jobs (sync / database / redis) |
| [`events`](events) | service | Sync event dispatch and model observers |
| [`notification`](notification) | service | Mail, SMS, push, database inbox, broadcast — this is how you send email |
| [`broadcasting`](broadcasting) | service | Channel events to log/file/null drivers (not a WebSocket server) |
| [`schedule`](schedule) | service | Cron-like tasks via `schedule:run` (no long-running daemon) |
| [`bus`](bus) | service | Sync command bus (`Dispatch` → handler). Not a queue |
| [`lock`](lock) | service | Process-local atomic locks (not distributed) |
| [`cron`](cron) | library | Cron expression parse and match |

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
| [`timing`](timing) | library | Server-Timing measurements |
| [`websocket`](websocket) | library | WebSocket upgrade helpers |
| [`useragent`](useragent) | library | Parse browser/OS from User-Agent |
| [`geo`](geo) | service | Resolve client geolocation |
| [`wellknown`](wellknown) | service | `/.well-known` and `security.txt` |

### Intelligence

| Package | Kind | What it does |
| --- | --- | --- |
| [`ai`](ai) | service | Chat / completion providers |
| [`rag`](rag) | library | Chunking, embed pipeline, vector store helpers |
| [`agent`](agent) | library | Agent loop, tools, conversation memory |

### Product addons

| Package | Kind | What it does |
| --- | --- | --- |
| [`billing`](billing) | service | Subscriptions / Stripe-style billing and webhooks |
| [`audit`](audit) | service | Request and audit event logging |
| [`backup`](backup) | service | Database backup/restore via native CLIs |
| [`graphql`](graphql) | service | GraphQL schema and queries |
| [`inspector`](inspector) | service | Request inspector toolbar data |
| [`features`](features) | service | In-memory feature flags and % rollouts |
| [`octane`](octane) | service | Concurrent request metrics + `GOMAXPROCS` hint. Not a multi-process app server |
| [`pulse`](pulse) | service | Metrics pulse dashboard |
| [`shorturl`](shorturl) | service | Create and resolve short URLs |
| [`sitemap`](sitemap) | service | Build XML sitemaps |
| [`tenancy`](tenancy) | service | Resolve current tenant from header/query/host (no auto DB isolation) |
| [`webhooks`](webhooks) | service | Signed outbound webhook delivery |
| [`circuit`](circuit) | service | Circuit breaker around flaky dependencies |
| [`docs`](docs) | service | Markdown documentation repository (docs sites) |
| [`version`](version) | service | Runtime version helper |

### Libraries

| Package | Kind | What it does |
| --- | --- | --- |
| [`api`](api) | library | API versioning helpers |
| [`archive/zipx`](archive/zipx) | library | ZIP create and extract |
| [`bloom`](bloom) | library | Bloom filter |
| [`browser`](browser) | library | Headless browser test helpers |
| [`collection`](collection) | library | In-memory collections (`Filter`, `Map`, …) |
| [`concurrency`](concurrency) | library | Parallel tasks: `Run` / `Map` / `Pool` |
| [`debug`](debug) | library | Dump helpers |
| [`export`](export) | library | CSV/XLSX import and export |
| [`image`](image) | library | Resize and encode images |
| [`jsonapi`](jsonapi) | library | JSON:API document helpers |
| [`jsonschema`](jsonschema) | library | JSON Schema validation |
| [`markdown`](markdown) | library | Markdown → HTML |
| [`openapi`](openapi) | library | OpenAPI generate/serve helpers |
| [`pagination`](pagination) | library | Page metadata for list endpoints |
| [`pdf`](pdf) | library | PDF generate and inline view |
| [`process`](process) | library | Run OS commands |
| [`qr`](qr) | heavy | QR code images (own `go.mod`) |
| [`resources`](resources) | library | API resource transformers |
| [`testing`](testing) | library | Feature tests (`Get("/").AssertOK()`) |

`bootutil` is an internal coerce/CLI helper. It is not a consumer package.

The name list also lives in the framework catalog (`kernel/catalog.go`). This tree is the code.

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

`db:setup` in the app pulls the driver you choose. None of them are linked until then — including SQLite.

## Local development

Clone next to the framework as `framework` (or `ZATRANO` via junction):

```text
replace github.com/zatrano/framework/v2 => ../framework
```

```bash
go test ./...
```

Work lands on **`main`**, same default branch as the framework.

## Import path

```go
import "github.com/zatrano/packages/billing"
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
