# Changelog

All notable changes to `github.com/zatrano/packages` are documented in this file.

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
