# Changelog

All notable changes to `github.com/zatrano/packages` are documented in this file.

This module is versioned independently of `github.com/zatrano/framework/v2`.
Historical `2.0.x` headings below recorded the framework pin, not a packages `/v2` module.

## Unreleased

## 1.10.0 - 2026-09-15

Why now: Drop the `api` library (path versioning is kernel routing) and ship auth email-verify env plus JSON surface parity with web.

### Breaking

- Removed the `api` library. Path versioning is `routing.Version` in the kernel (`github.com/zatrano/framework/v2@v2.6.0`). `make:auth` JSON routes call `routing.Version`.

### Added

- `auth`: `AUTH_MUST_VERIFY_EMAIL` (default `false`) gates verification mail and `VerifyEmailMiddleware`. `package:enable auth` merges the env key; `make:auth` / `make:panel` wire the middleware.
- `auth`: `make:auth` JSON surface includes verify/resend/`GET /user`/2FA status to match the web flows. Signed links use `url.From(app)` (`make:auth` enables `url`).

Install with `go get github.com/zatrano/packages@v1.10.0`.

## 1.9.1 - 2026-09-15

Why now: Restore the markdown `docs` addon and ship a single `seo` addon so applications that served sitemaps, robots.txt, and LLM discovery files can pin without tracking `main`.

### Added

- Restored `docs` (`From`, `Repository`, `Register`). Markdown rendering uses `toolkit/markdown`.
- `seo` is one addon for classic crawlers and LLM discovery: sitemap.xml, robots.txt (no `LLMs-Txt` directive), OG/JSON-LD `ViewData`, security.txt, `/llms.txt`, `/llms-full.txt`, `/.well-known/ai-plugin.json`. Optional `docs` fills sitemap and llms-full. It does not import `ai`. There is no `sitemap` or `wellknown` package.

Install with `go get github.com/zatrano/packages@v1.9.1`.

## 1.9.0 - 2026-09-15

Why now: Pin the catalog freeze so applications can import toolkit libraries, nested pagination/TOTP/OTP, and experimental `workflow` without tracking `main`.

### Breaking

- Removed the `billing` addon.
- `enums` is now `toolkit/enums` (import-only, like `toolkit/str`). There is no `make:enum` and no `package:enable enums`.
- Moved `collection`, `bloom`, `concurrency`, `cron`, `debug`, `process`, and `timing` under `toolkit/` (same import-only rule).
- Moved `markdown` to `toolkit/markdown` and `archive/zipx` to `toolkit/zip` (`package zipx`). `schedule` uses `toolkit/cron`; the nested `schedule/cron` copy is gone.
- Removed the `useragent` addon. Parse with `http.ParseUserAgent` / `req.Agent()` in the kernel (the framework cannot import this module).
- Moved `jsonschema` to `toolkit/jsonschema`. `circuit` is `toolkit/circuit` (`New` / `Breaker`); no `package:enable circuit`, no `From(app)`.
- Removed `octane`, `pulse`, and `inspector`.
- Removed `version` (`app.Version()` / kernel CLI remain). Go.mod pin tests live at module root.
- Removed `search`, `shorturl`, `sitemap`, `wellknown`, `geo`, and `docs`.
- ORM paginators live in `orm/pagination`. There is no standalone `pagination` package.
- TOTP lives in `auth/totp` (MFA). Numeric codes live in `notification/otp`. There is no standalone `totp` or `otp` package.
- `hashid` and `lock` are `toolkit/hashid` and `toolkit/lock` (import-only; no enablement).
- Removed `bus`, `features`, and `tenancy`. GraphQL no longer ships a demo `feature` query.
- `make:dashboard` (and its stubs) are removed; use `make:panel`.

### Added

- `workflow` is catalogued as an experimental intelligence library (import-only; agents enter with `agent.AsExecutor`).

### Changed

- Broadcasting no longer declares Optional `auth` (private channels resolve `auth` at request time). Official addon Requires/Optional graph is cycle-tested.
- `make:auth` writes `controllers/auth/{web,api}`, `routes/auth/{web,api}`, `views/auth` pages, `views/layout/{auth,mail}`, and `views/mail/auth`. Default output has no social files, routes, views, or lang keys. `--social` / `--social=google` adds them. GitHub is not generated.
- `package:enable view` (and `make:auth`) write `views/layout/app.html` and `views/web/welcome.html` when missing, and switch the starter `HomeController` from `http.HTML` to `http.View("web.welcome")`.
- `make:panel {name}` scaffolds a named HTML surface.

This module still requires `github.com/zatrano/framework/v2 v2.0.28` (minimum). Nested-module publication remains a separate tagging operation.

## 1.8.0 - 2026-09-15

Why now: Publish toolkit libraries and experimental AI/RAG/agent labels so framework v2.4.0 consumers can import them without tracking `main`.

### Added

- `toolkit/{arr,color,date,html,money,num,str}`: general-purpose helpers moved out of the kernel (`kernel/support/...`). Opt-in libraries (`LayerAddon`). Import `github.com/zatrano/packages/toolkit/str` (and siblings). No kernel re-export. Validation `IsSemver` uses `toolkit/str`.
- View expressions (`@if` / `@elseif` / `{{ }}` / `{!! !!}`) compile ternary (`? :`), null coalescing (`??`), elvis (`?:`), concat (`.`), `xor` / `**`, indexing, and common helpers (`empty`, `isset`, `count`, `in_array`, `str_contains`, `filled`, `blank`, `data_get`, …).
- `ai`, `rag`, and `agent` are catalogued as `Stability: experimental` until they complete the same security review as the rest of the ecosystem.

### Changed

- OAuth, ORM, and WebAuthn check `crypto/rand` errors instead of discarding them.

This module still requires `github.com/zatrano/framework/v2 v2.0.28` (minimum). Nested-module publication remains a separate tagging operation.

## 1.7.2 - 2026-09-10

### Fixed

- `unique` / `exists` fail closed when no PresenceChecker is bound, the rule is missing a table or column, or the checker returns an error. An infrastructure failure is no longer treated as a successful lookup. The database package's default checker also returns an error for empty table or column instead of `(false, nil)`. `unique` / `exists` no longer silently succeed when the database checker or required infrastructure cannot determine the result.

This module still requires `github.com/zatrano/framework/v2 v2.0.28` (minimum). Nested-module publication remains a separate tagging operation.

## 1.7.1 - 2026-09-09

Correction release after `v1.7.0`. That tag stays published and must not be moved.

### Fixed

- Root `go.mod` no longer requires `github.com/zatrano/packages/database/driver/sqlite`. That nested driver is a separate module, linked by `db:setup`, and was not publicly resolvable (`unknown revision database/driver/sqlite/v1.0.0`). Nested Go tags must follow the module path (`database/driver/sqlite/v1.0.0`), not a `packages/` prefix.
- Cache and queue fail closed when `CACHE_STORE=redis` or `QUEUE_CONNECTION=redis` but no usable Redis binding exists. Default `file` / `sync` still boot if Redis is absent. Cache remains the Redis connection owner; queue only consumes `"redis"`; `redisx` stays a library.

In-tree ORM/query tests may still use public `modernc.org/sqlite`.

## 1.7.0 - 2026-09-09

Require `github.com/zatrano/framework/v2 v2.0.28`. Public module path stays `github.com/zatrano/packages` (no `/v2` suffix).

### Changed

- Enablement graph: declare real `Requires` / `Optional` (auth, apitoken, queue, notification, backup, broadcasting, database). Optional names are not auto-enabled.
- `rag` and `agent` are import-only libraries (no addon `Register` / no-op `ServiceProvider`).
- `redisx` is a library. Cache owns the Redis client and publishes `"redis"`; queue reads that binding.
- Webhooks, filesystem signed URLs, HashID salt, and social OAuth fail closed on missing or placeholder secrets.
- ORM connection (`Configure` / `DB`) is mutex-guarded. The database service still owns the connection.
- AI `LogDriver` does not log prompts or replies unless `log_prompts` is opted in. Gemini sends the API key in `x-goog-api-key`, not the query string.

## 2.0.27 - 2026-09-08

Require `github.com/zatrano/framework/v2 v2.0.27`. This packages module is still not tagged `v2.x`.

## 2.0.26 - 2026-09-08

Require `github.com/zatrano/framework/v2 v2.0.26`. This packages module is still not tagged `v2.x`.

## 2.0.25 - 2026-09-08

Require `github.com/zatrano/framework/v2 v2.0.25`. This packages module is still not tagged `v2.x`.

## 2.0.24 - 2026-09-08

Require `github.com/zatrano/framework/v2 v2.0.24`. This packages module is still not tagged `v2.x`.

## 2.0.22 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.22`. This packages module is still not tagged `v2.x`.

## 2.0.21 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.21`. This packages module is still not tagged `v2.x`.

## 2.0.20 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.20`. This packages module is still not tagged `v2.x`.

## 2.0.19 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.19`. This packages module is still not tagged `v2.x`.

## 2.0.18 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.18`. This packages module is still not tagged `v2.x`.

## 2.0.17 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.17`. This packages module is still not tagged `v2.x`.

## 2.0.16 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.16`. This packages module is still not tagged `v2.x`.

## 2.0.14 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.14`. This packages module is still not tagged `v2.x`.

## 2.0.13 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.13`. This packages module is still not tagged `v2.x`.

## 2.0.12 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.12`. This packages module is still not tagged `v2.x`.

## 2.0.11 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.11`. This packages module is still not tagged `v2.x`.

## 2.0.10 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.10`. This packages module is still not tagged `v2.x`.

## 2.0.9 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.9`. This packages module is still not tagged `v2.x`.

## 2.0.8 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.8`. This packages module is still not tagged `v2.x`.

## 2.0.7 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.7`. This packages module is still not tagged `v2.x`.

## 2.0.6 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.6`. This packages module is still not tagged `v2.x`.

## 2.0.5 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.5`. This packages module is still not tagged `v2.x`.

## 2.0.4 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.4`. This packages module is still not tagged `v2.x`.

## 2.0.3 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.3` (registry data model on the kernel). This packages module is still not tagged `v2.x`.

## 2.0.2 - 2026-09-07

Require `github.com/zatrano/framework/v2 v2.0.2`. Bind `apitoken` on Register so `From(app)` works. Document Enabled âˆ© Imported. Tests for hashing, health, redisx, and testing.

This packages module is still not tagged `v2.x`.

## 2.0.1 - 2026-09-06

Require `github.com/zatrano/framework/v2 v2.0.1` (GOPROXY-valid kernel module). This packages module is still not tagged `v2.x`.

## 2.0.0 - 2026-09-06

v2 packages are the default line on `main`. This module stays `github.com/zatrano/packages` (no `/v2` suffix) so it must not be tagged `v2.x`. Use `go get github.com/zatrano/packages@main`. The kernel it requires is `github.com/zatrano/framework/v2`.

### Added

- Agent tool execution returns typed `ToolResult` (`ok` / `error` / `timeout` / `denied` / `invalid`, retryable, model `Content()`). `Registry.Execute` still returns `(string, error)`.
- First-party addon packages split from the ZATRANO framework module (Stage B).
- Consumers blank-import a package (for example `github.com/zatrano/packages/billing`) and enable it with `bootstrap.WithAddons(...)`.

### Changed

- Imports that resolve `app/views`, `app/localization`, and `app/database` now use `github.com/zatrano/framework/v2/kernel/dirs` (was `kernel/layout`).
- CI: tests, coding style, static analysis, and security (same set as the framework). Linux jobs check out `zatrano/framework@main` as a sibling.
- `validation` no longer imports `flash` (old input is flashed on the session directly). Importing validation/database/billing must not register the flash addon.
- Browser feature tests register probe routes before `Bootstrap` (router is frozen after boot).

### Notes

- Go module path is `github.com/zatrano/packages` (no `/v2` suffix). Do **not** tag `v2.0.0-alpha`: the Go toolchain would require `github.com/zatrano/packages/v2`. Use `v0.x` / `v1.x` tags (for example `v1.0.0-alpha`) until a real v1/v2 module decision.
- Local development: `replace github.com/zatrano/framework/v2 => ../framework`.
