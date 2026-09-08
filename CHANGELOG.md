# Changelog

All notable changes to `github.com/zatrano/packages` are documented in this file.

This module is versioned independently of `github.com/zatrano/framework/v2`.
Historical `2.0.x` headings below recorded the framework pin, not a packages `/v2` module.

## 1.7.0 Unreleased

Require `github.com/zatrano/framework/v2 v2.0.28`. Public module path stays `github.com/zatrano/packages` (no `/v2` suffix). Last published tag remains `v1.6.6` until this line is tagged.

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
