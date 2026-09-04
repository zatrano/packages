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
  <a href="https://github.com/zatrano/framework/blob/v2-dev/PACKAGES.md">Package guide</a>
</p>

<p align="center">
  Active line: <a href="https://github.com/zatrano/packages/tree/v2-dev"><code>v2-dev</code></a>
  ·
  <code>go get github.com/zatrano/packages@v2-dev</code>
</p>

---

This module is **`github.com/zatrano/packages`**. It is not the kernel.

The kernel lives in [`github.com/zatrano/framework`](https://github.com/zatrano/framework): HTTP, routing, middleware, config, the CLI, `zatrano new`. Everything that used to look like “the rest of the framework” — sessions, auth, database, views, queues, AI — lives **here**, next to OAuth, billing, and the import-only helpers.

The two modules cannot be merged: this one already requires the framework.

```text
  github.com/zatrano/framework          github.com/zatrano/packages
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
go get github.com/zatrano/packages@v2-dev
go run ./cmd/app package:enable auth
go run ./cmd/app package:list
go run ./cmd/app package:doctor
```

`package:enable` writes the blank-import into `bootstrap/addons.go`. That is the only “enable” step. Libraries are never enabled — you just `import` them.

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

## What is in this repo

**Web stack** — `session`, `flash`, `validation`, `view`, `assets`, `localization`, `filesystem`

**Identity** — `auth`, `authorization`, `hashing`, `apitoken`

**Data** — `database`, `orm`, `factory`, `cache`, `redisx`

**Async** — `queue`, `events`, `notification`, `broadcasting`, `schedule`

**HTTP extras** — `httpclient`, `ratelimit`, `url`, `maintenance`, `health`, `observability`

**Intelligence** — `ai` (service), `rag` and `agent` (libraries)

**Opt-in services** — `oauth`, `social`, `billing`, `webauthn`, `audit`, `backup`, `mongo`, `graphql`, `inspector`, `features`, `lock`, `octane`, `pulse`, `search`, `shorturl`, `sitemap`, `tenancy`, `webhooks`, `wellknown`, …

**Libraries** — `collection`, `totp`, `resources`, `pagination`, `export`, `openapi`, `testing`, `timing`, `useragent`, `websocket`, `markdown`, `image`, `pdf`, …

The catalog in the framework (`kernel/catalog.go`) is the name list. This tree is the code.

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
replace github.com/zatrano/framework => ../framework
```

```bash
go test ./...
```

Work lands on **`v2-dev`**, same branch name as the framework.

## Import path

```go
import "github.com/zatrano/packages/billing"
import "github.com/zatrano/packages/auth"
import "github.com/zatrano/framework/kernel/http"   // kernel, not this module
```

Kernel types (`http.Request`, the router, CSRF) stay in the framework. This module implements the rest and talks to the kernel through `contracts.App`.

## Docs

| | |
| --- | --- |
| [Framework README](https://github.com/zatrano/framework) | Kernel, `zatrano new`, boot rule |
| [PACKAGES.md](https://github.com/zatrano/framework/blob/v2-dev/PACKAGES.md) | Purpose and usage per package |
| [zatrano.com/docs](https://zatrano.com/docs) | Product guides |
| [Package ecosystem](https://zatrano.com/docs/package-ecosystem) | Enable, doctor, presets |

## License

Same terms as the framework: MIT · Copyright (c) 2026 Serhan KARAKOÇ
