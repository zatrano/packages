<p align="center">
  <strong>ZATRANO packages</strong>
</p>

<p align="center">
  <em>Official addons for the ZATRANO kernel. Import a package, it boots. Leave it out, it is not in the binary.</em>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/zatrano/packages"><img src="https://img.shields.io/badge/golang-1.25+-00ADD8?logo=go&logoColor=white" alt="Golang"></a>
  <a href="https://github.com/zatrano/packages/blob/v2-dev/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/zatrano/framework"><img src="https://img.shields.io/badge/kernel-zatrano%2Fframework-222?logo=go" alt="Framework"></a>
</p>

<p align="center">
  <a href="https://github.com/zatrano/packages/tree/v2-dev"><code>v2-dev</code></a>
  ·
  <a href="https://github.com/zatrano/framework">Framework</a>
  ·
  <a href="https://zatrano.com/docs">Docs</a>
  ·
  <a href="https://github.com/zatrano/framework/blob/v2-dev/PACKAGES.md">Package guide</a>
</p>

---

**This `main` branch is the GitHub landing page.** The module source — every addon — is on **[`v2-dev`](https://github.com/zatrano/packages/tree/v2-dev)**.

```bash
git clone -b v2-dev https://github.com/zatrano/packages.git
go get github.com/zatrano/packages@v2-dev
```

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

Each service package registers itself in `init()`. The **application** blank-imports it. `bootstrap.App()` loads every registered provider. If you do not import it, it is not compiled in and it does not boot.

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

```bash
go get github.com/zatrano/packages@v2-dev
go run ./cmd/app package:enable auth
go run ./cmd/app package:list
```

| Kind | In the binary when | Examples |
| --- | --- | --- |
| **Service** | You blank-import it | `auth`, `database`, `queue`, `ai`, `oauth` |
| **Library** | You `import` it in your code | `collection`, `totp`, `resources`, `rag` |
| **Heavy** | Own `go.mod`, only when needed | `webauthn`, `mongo`, `qr`, SQL drivers |

Resolve with `From(app)` — not `app.Auth()` on the kernel. There is no `mail` package; send email through `notification` with `Channels: ["mail"]`.

## What is in this repo

**Web stack** — `session`, `flash`, `validation`, `view`, `assets`, `localization`, `filesystem`

**Identity** — `auth`, `authorization`, `hashing`, `apitoken`

**Data** — `database`, `orm`, `factory`, `cache`, `redisx`

**Async** — `queue`, `events`, `notification`, `broadcasting`, `schedule`

**Intelligence** — `ai` (service), `rag` and `agent` (libraries)

**Opt-in services** — `oauth`, `social`, `billing`, `webauthn`, `audit`, `backup`, `mongo`, `graphql`, `inspector`, …

**Libraries** — `collection`, `totp`, `resources`, `pagination`, `export`, `openapi`, `testing`, …

SQL drivers (including SQLite) are nested modules under `database/driver/`. None are linked until `db:setup`.

## Docs

| | |
| --- | --- |
| [Browse the code](https://github.com/zatrano/packages/tree/v2-dev) | `v2-dev` tree |
| [Framework README](https://github.com/zatrano/framework) | Kernel, `zatrano new`, boot rule |
| [PACKAGES.md](https://github.com/zatrano/framework/blob/v2-dev/PACKAGES.md) | Purpose and usage per package |
| [zatrano.com/docs](https://zatrano.com/docs) | Product guides |

## License

MIT · Copyright (c) 2026 Serhan KARAKOÇ — see [`LICENSE` on v2-dev](https://github.com/zatrano/packages/blob/v2-dev/LICENSE).
